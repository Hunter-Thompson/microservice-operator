package controllers

import (
	"context"
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"

	"github.com/go-logr/logr"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MicroserviceReconciler) checkAutoscaling(mic *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	if mic.Spec.Autoscaling == nil {
		return r.Resources.DeleteHPA(types.NamespacedName{Name: mic.GetName(), Namespace: mic.GetNamespace()}, reqLogger)
	}

	if mic.Annotations["scheduledautoscaler.override"] == "true" {
		l := fmt.Sprintf("scheduledautoscaler override found, skipping checkAutoscaling for %s", mic.GetName())
		reqLogger.Info(l)
		return nil
	}

	desired := microservice.GenerateAutoscalingv2(mic)

	err := r.Resources.CreateHPAIfNotExists(mic, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &autoscalingv2.HorizontalPodAutoscaler{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}
