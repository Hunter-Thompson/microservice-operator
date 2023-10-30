package microservice

import (
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func DeploymentOwnerReference(deployment *microservicev1beta1.Deployment) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(deployment, schema.GroupVersionKind{
			Group:   microservicev1beta1.GroupVersion.Group,
			Version: microservicev1beta1.GroupVersion.Version,
			Kind:    "Deployment",
		}),
	}
}

func GenerateIngressesV1(deployment *microservicev1beta1.Deployment) []*networking.Ingress {
	ingresses := []*networking.Ingress{}

	for _, ing := range deployment.Spec.Ingress {
		name := fmt.Sprintf("%s-%s", deployment.Name, ing.Name)
		ingress := newNetworkingV1Ingress(deployment, name)
		ingresses = append(ingresses, configureIngress(deployment, &ing, ingress))
	}

	return ingresses
}

func newNetworkingV1Ingress(deployment *microservicev1beta1.Deployment, name string) *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
		},
	}
}

func configureIngress(deployment *microservicev1beta1.Deployment, ing *microservicev1beta1.Ingress, ingress *networking.Ingress) *networking.Ingress {
	return configureIngressRules(deployment, ing, ingress)
}

func configureIngressRules(deployment *microservicev1beta1.Deployment, ing *microservicev1beta1.Ingress, ingress *networking.Ingress) *networking.Ingress {
	paths := []networking.HTTPIngressPath{}
	pathType := networking.PathTypeImplementationSpecific

	for _, path := range ing.Paths {
		paths = append(paths, networking.HTTPIngressPath{
			Path:     path,
			PathType: &pathType,
			Backend: networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: deployment.Name,
					Port: networking.ServiceBackendPort{
						Number: ing.ContainerPort,
					},
				},
			},
		})
	}

	if len(paths) == 0 {
		paths = append(paths, networking.HTTPIngressPath{
			PathType: &pathType,
			Backend: networking.IngressBackend{
				Service: &networking.IngressServiceBackend{
					Name: deployment.Name,
					Port: networking.ServiceBackendPort{
						Number: ing.ContainerPort,
					},
				},
			},
		})
	}

	ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
		Host: ing.Host,
		IngressRuleValue: networking.IngressRuleValue{
			HTTP: &networking.HTTPIngressRuleValue{
				Paths: paths,
			},
		},
	})

	return ingress
}

func GenerateServiceV1(deployment *microservicev1beta1.Deployment) *corev1.Service {
	service := newServiceV1Beta(deployment, map[string]string{})

	return configureService(deployment, service)
}

func newServiceV1Beta(deployment *microservicev1beta1.Deployment, annotations map[string]string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deployment.Name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
			Annotations:     annotations,
		},
	}
}

func configureService(deployment *microservicev1beta1.Deployment, service *corev1.Service) *corev1.Service {
	ports := []corev1.ServicePort{}
	for _, ingress := range deployment.Spec.Ingress {
		ports = append(ports, corev1.ServicePort{
			Port:       ingress.ContainerPort,
			TargetPort: intstr.FromInt(int(ingress.ContainerPort)),
			Protocol:   corev1.ProtocolTCP,
			Name:       ingress.Name,
		})
	}

	service.Spec.Ports = ports
	service.Spec.Type = corev1.ServiceTypeNodePort

	return service
}
