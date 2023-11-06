/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	cron "github.com/robfig/cron/v3"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
)

// ScheduledAutoscalerReconciler reconciles a ScheduledAutoscaler object
type ScheduledAutoscalerReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	allcron   *cron.Cron
	Resources *resources.ResourceHelper
}

func NewScheduledAutoscalerReconciler(mgr ctrl.Manager, allcron *cron.Cron) *ScheduledAutoscalerReconciler {
	return &ScheduledAutoscalerReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		allcron:   allcron,
		Resources: resources.NewResourceHelper(mgr.GetClient(), mgr.GetScheme()),
	}
}

//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=scheduledautoscalers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=scheduledautoscalers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=scheduledautoscalers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ScheduledAutoscaler object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *ScheduledAutoscalerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	sa := &microservicev1beta1.ScheduledAutoscaler{}
	err := r.Client.Get(ctx, req.NamespacedName, sa)
	if err != nil && k8sErrors.IsNotFound(err) {
		// Request object not found, could have been deleted after reconcile
		// request. Owned objects are automatically garbage collected.
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// We copy status to not to refetch the resource
	status := sa.Status

	// name of our custom finalizer
	myFinalizerName := "microservice.example.com/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if sa.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(sa, myFinalizerName) {
			controllerutil.AddFinalizer(sa, myFinalizerName)
			if err := r.Update(ctx, sa); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(sa, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			l := fmt.Sprintf("%s is being deleted", sa.GetName())
			reqLogger.Info(l)

			for _, scheduledID := range status.ScheduledCrons {
				r.allcron.Remove(cron.EntryID(scheduledID))
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(sa, myFinalizerName)
			if err := r.Update(ctx, sa); err != nil {
				return ctrl.Result{}, err
			}

			l = fmt.Sprintf("removed all crons for %s from schedule", sa.GetName())
			reqLogger.Info(l)
		}

		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	if status.State != microservicev1beta1.Reconciling {
		err = r.updateStatusReconciling(sa, status, reqLogger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	for _, scheduledID := range status.ScheduledCrons {
		r.allcron.Remove(cron.EntryID(scheduledID))
	}

	allSchedules := sa.Spec.Schedules

	scheduledStatusUpdate := []int{}
	for _, toSchedule := range allSchedules {
		l := fmt.Sprintf("adding %s cron to schedule for %s", toSchedule.Name, sa.Spec.MicroserviceName)
		reqLogger.Info(l)

		id, err := r.allcron.AddFunc(toSchedule.Cron, r.scale(sa, &toSchedule, status, reqLogger))
		if err != nil {
			r.updateStatusReconcilingAndLogError(sa, status, reqLogger, err)
			return reconcile.Result{}, err
		}

		scheduledStatusUpdate = append(scheduledStatusUpdate, int(id))
	}

	status.State = microservicev1beta1.Stable
	err = r.updateStatus(sa, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(sa, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	status.ScheduledCrons = scheduledStatusUpdate
	err = r.updateStatusWithCronID(sa, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(sa, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ScheduledAutoscalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&microservicev1beta1.ScheduledAutoscaler{}).
		WithEventFilter(pred).
		Complete(r)
}
