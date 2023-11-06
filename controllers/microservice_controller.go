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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

// MicroserviceReconciler reconciles a Microservice object
type MicroserviceReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	Resources *resources.ResourceHelper
}

func NewMicroserviceReconciler(mgr ctrl.Manager) *MicroserviceReconciler {
	return &MicroserviceReconciler{
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
		Resources: resources.NewResourceHelper(mgr.GetClient(), mgr.GetScheme()),
	}
}

//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=deployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=microservice.microservice.example.com,resources=deployments/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Microservice object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *MicroserviceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reqLogger := log.FromContext(ctx)

	deployment := &microservicev1beta1.Microservice{}
	err := r.Client.Get(ctx, req.NamespacedName, deployment)
	if err != nil && k8sErrors.IsNotFound(err) {
		// Request object not found, could have been deleted after reconcile
		// request. Owned objects are automatically garbage collected.
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// We copy status to not to refetch the resource
	status := deployment.Status

	if status.State != microservicev1beta1.Reconciling {
		err = r.updateStatusReconciling(deployment, status, reqLogger)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = r.checkServiceAccount(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	err = r.checkServiceAccountSecret(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	err = r.checkDeployment(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	err = r.checkAutoscaling(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	err = r.checkService(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	err = r.checkIngress(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	status.State = microservicev1beta1.Stable
	err = r.updateStatus(deployment, status, reqLogger)
	if err != nil {
		r.updateStatusReconcilingAndLogError(deployment, status, reqLogger, err)
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MicroserviceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	pred := predicate.GenerationChangedPredicate{}
	return ctrl.NewControllerManagedBy(mgr).
		For(&microservicev1beta1.Microservice{}).
		WithEventFilter(pred).
		Complete(r)
}
