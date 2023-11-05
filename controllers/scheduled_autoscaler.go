package controllers

import (
	"context"
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/go-logr/logr"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/types"
)

// updateStatusReconciling sets the Mattermost state to reconciling.
func (r *ScheduledAutoscalerReconciler) scale(sa *microservicev1beta1.ScheduledAutoscaler, schedule *microservicev1beta1.Schedule, status microservicev1beta1.ScheduledAutoscalerStatus, reqLogger logr.Logger) func() {
	return func() {
		l := fmt.Sprintf("autoscaling %s", sa.Spec.MicroserviceName)
		reqLogger.Info(l)

		current := &autoscalingv2.HorizontalPodAutoscaler{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: sa.Spec.MicroserviceName, Namespace: sa.Namespace}, current)
		if err != nil {
			reqLogger.Error(err, "failed to get HPA")
			return
		}

		current.Spec.MaxReplicas = schedule.MaxReplicas
		current.Spec.MinReplicas = &schedule.MinReplicas

		err = r.Client.Update(context.TODO(), current)
		if err != nil {
			reqLogger.Error(err, "failed to update HPA")
			return
		}

		mic := &microservicev1beta1.Microservice{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: sa.Spec.MicroserviceName, Namespace: sa.Namespace}, mic)
		if err != nil {
			reqLogger.Error(err, "failed to get microservice")
			return
		}

		mic.ObjectMeta.Annotations["scheduledautoscaler.override"] = "true"
		err = r.Client.Update(context.TODO(), mic)
		if err != nil {
			reqLogger.Error(err, "failed to update microservice annotations")
			return
		}

		l = fmt.Sprintf("successfully scaled %s", sa.Spec.MicroserviceName)
		reqLogger.Info(l)
	}
}
