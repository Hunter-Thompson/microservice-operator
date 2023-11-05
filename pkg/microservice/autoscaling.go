package microservice

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
)

func GenerateAutoscalingv2(mic *microservicev1beta1.Microservice) *autoscalingv2.HorizontalPodAutoscaler {
	if mic.Spec.Autoscaling == nil {
		return nil
	}
	return newAutoscalingv2(mic)
}

func newAutoscalingv2(mic *microservicev1beta1.Microservice) *autoscalingv2.HorizontalPodAutoscaler {
	return &autoscalingv2.HorizontalPodAutoscaler{
		ObjectMeta: v1.ObjectMeta{
			Name:            mic.Name,
			Namespace:       mic.Namespace,
			OwnerReferences: DeploymentOwnerReference(mic),
			Labels:          mic.Spec.Labels,
			Annotations:     mic.GetAnnotations(),
		},
		Spec: *mic.Spec.Autoscaling,
	}
}
