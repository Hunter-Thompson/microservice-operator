package microservice

import (
	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GenerateServiceV1(deployment *microservicev1beta1.Microservice) *corev1.Service {
	service := newServiceV1Beta(deployment)

	return configureService(deployment, service)
}

func newServiceV1Beta(deployment *microservicev1beta1.Microservice) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deployment.Name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
			Labels:          deployment.Spec.Labels,
			Annotations:     deployment.GetAnnotations(),
		},
	}
}

func configureService(deployment *microservicev1beta1.Microservice, service *corev1.Service) *corev1.Service {
	ports := []corev1.ServicePort{}
	for _, ingress := range deployment.Spec.Ingress {
		ports = append(ports, corev1.ServicePort{
			Port:       ingress.ContainerPort,
			TargetPort: intstr.FromInt(int(ingress.ContainerPort)),
			Protocol:   corev1.ProtocolTCP,
			Name:       ingress.Name,
		})
	}

	service.Spec.Selector = deployment.Spec.Labels
	service.Spec.Ports = ports
	service.Spec.Type = corev1.ServiceTypeNodePort

	return service
}
