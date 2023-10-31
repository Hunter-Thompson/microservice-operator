package microservice

import (
	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateIngressV1(deployment *microservicev1beta1.Microservice) *networking.Ingress {
	ingress := newNetworkingV1Ingress(deployment, deployment.Name)
	return configureIngress(deployment, deployment.Spec.Ingress, ingress)
}

func newNetworkingV1Ingress(deployment *microservicev1beta1.Microservice, name string) *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
			Labels:          deployment.Spec.Labels,
			Annotations:     deployment.Spec.IngressAnnotations,
		},
	}
}

func configureIngress(deployment *microservicev1beta1.Microservice, ing []microservicev1beta1.Ingress, ingress *networking.Ingress) *networking.Ingress {
	for _, v := range ing {
		configureIngressRules(deployment, &v, ingress)
	}

	return ingress
}

func configureIngressRules(deployment *microservicev1beta1.Microservice, ing *microservicev1beta1.Ingress, ingress *networking.Ingress) *networking.Ingress {
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
