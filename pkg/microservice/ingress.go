package microservice

import (
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateIngressesV1(deployment *microservicev1beta1.Microservice) []*networking.Ingress {
	ingresses := []*networking.Ingress{}
	for _, ing := range deployment.Spec.Ingress {
		ingName := fmt.Sprintf("%s-%s", deployment.GetName(), ing.Name)
		ingresses = append(ingresses, configureIngressRules(deployment, &ing, newNetworkingV1Ingress(deployment, ingName, ing.Annotations)))
	}

	return ingresses
}

func newNetworkingV1Ingress(deployment *microservicev1beta1.Microservice, name string, annotations map[string]string) *networking.Ingress {
	return &networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
			Labels:          deployment.Spec.Labels,
			Annotations:     annotations,
		},
	}
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
					Name: ing.Name,
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
					Name: ing.Name,
					Port: networking.ServiceBackendPort{
						Number: ing.ContainerPort,
					},
				},
			},
		})
	}

	if len(ing.Hosts) == 0 {
		ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
			IngressRuleValue: networking.IngressRuleValue{
				HTTP: &networking.HTTPIngressRuleValue{
					Paths: paths,
				},
			},
		})
	} else {
		for _, host := range ing.Hosts {
			ingress.Spec.Rules = append(ingress.Spec.Rules, networking.IngressRule{
				Host: host,
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: paths,
					},
				},
			})
		}
	}

	return ingress
}
