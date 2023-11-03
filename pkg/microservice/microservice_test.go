package microservice

import (
	"testing"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGenerate(t *testing.T) {
	msName := "foo"
	msNamespace := "default"
	labels := map[string]string{
		"app": "test",
	}
	svcName1 := "test-1"
	svcPort1 := 8080
	ingHost1 := "dsa.example.com"
	pathType1 := networking.PathTypeImplementationSpecific
	path1 := "/asd"
	anotherPath1 := "/dfoig"

	svcName2 := "test-2"
	svcPort2 := 8090
	ingHost2 := "asd.example.com"
	pathType2 := networking.PathTypeImplementationSpecific
	path2 := "/dsa"
	image := "image:latest"
	nodeSelector := map[string]string{
		"app": "test",
	}
	replicas := int32(8)
	env := map[string]string{
		"test": "env",
	}
	tolerations := []v1.Toleration{
		{
			Key:    "test",
			Value:  "toleration",
			Effect: v1.TaintEffectNoExecute,
		},
	}
	podAnnotations := map[string]string{
		"test": "env",
	}

	ms := &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      msName,
			Namespace: msNamespace,
			UID:       types.UID("test"),
		},
	}

	ms.Spec = microservicev1beta1.MicroserviceSpec{
		Image:          image,
		Labels:         labels,
		Replicas:       replicas,
		Env:            env,
		NodeSelector:   nodeSelector,
		Tolerations:    tolerations,
		PodAnnotations: podAnnotations,
		Ingress: []microservicev1beta1.Ingress{
			{
				ContainerPort: int32(svcPort1),
				Name:          svcName1,
				Host:          ingHost1,
				Paths:         []string{path1, anotherPath1},
			},
			{
				ContainerPort: int32(svcPort2),
				Name:          svcName2,
				Host:          ingHost2,
				Paths:         []string{path2},
			},
		},
	}

	t.Run("service", func(t *testing.T) {
		svc := GenerateServiceV1(ms)
		assert.Equal(t, corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeNodePort,
			Ports: []corev1.ServicePort{
				{
					Port:       int32(svcPort1),
					TargetPort: intstr.FromInt(svcPort1),
					Protocol:   corev1.ProtocolTCP,
					Name:       svcName1,
				},
				{
					Port:       int32(svcPort2),
					TargetPort: intstr.FromInt(svcPort2),
					Protocol:   corev1.ProtocolTCP,
					Name:       svcName2,
				},
			},
		}, svc.Spec)
	})

	t.Run("ingress", func(t *testing.T) {
		ing := GenerateIngressesV1(ms)
		assert.Equal(t, networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: ingHost1,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:     path1,
									PathType: &pathType1,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: svcName1,
											Port: networking.ServiceBackendPort{
												Number: int32(svcPort1),
											},
										},
									},
								},
								{
									Path:     anotherPath1,
									PathType: &pathType1,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: svcName1,
											Port: networking.ServiceBackendPort{
												Number: int32(svcPort1),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, ing[0].Spec)

		assert.Equal(t, networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					Host: ingHost2,
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									Path:     path2,
									PathType: &pathType2,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: svcName2,
											Port: networking.ServiceBackendPort{
												Number: int32(svcPort2),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, ing[1].Spec)

		ms.Spec = microservicev1beta1.MicroserviceSpec{
			Image:          image,
			Labels:         labels,
			Replicas:       replicas,
			Env:            env,
			NodeSelector:   nodeSelector,
			Tolerations:    tolerations,
			PodAnnotations: podAnnotations,
			Ingress: []microservicev1beta1.Ingress{
				{
					ContainerPort: int32(svcPort1),
					Name:          svcName1,
				},
				{
					ContainerPort: int32(svcPort2),
					Name:          svcName2,
				},
			},
		}

		ing = GenerateIngressesV1(ms)
		assert.Equal(t, networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									PathType: &pathType1,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: svcName1,
											Port: networking.ServiceBackendPort{
												Number: int32(svcPort1),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, ing[0].Spec)

		assert.Equal(t, networking.IngressSpec{
			Rules: []networking.IngressRule{
				{
					IngressRuleValue: networking.IngressRuleValue{
						HTTP: &networking.HTTPIngressRuleValue{
							Paths: []networking.HTTPIngressPath{
								{
									PathType: &pathType2,
									Backend: networking.IngressBackend{
										Service: &networking.IngressServiceBackend{
											Name: svcName2,
											Port: networking.ServiceBackendPort{
												Number: int32(svcPort2),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}, ing[1].Spec)
	})

	t.Run("deployment", func(t *testing.T) {
		deployment := GenerateDeployment(ms)
		assert.Equal(t, []v1.Container{
			{
				Name:  ms.GetName(),
				Image: image,
				Ports: []v1.ContainerPort{
					{
						Name:          svcName1,
						ContainerPort: int32(svcPort1),
					},
					{
						Name:          svcName2,
						ContainerPort: int32(svcPort2),
					},
				},
				Env: []v1.EnvVar{
					{
						Name:  "test",
						Value: "env",
					},
				},
			},
		}, deployment.Spec.Template.Spec.Containers)

		assert.Equal(t, labels, deployment.Labels)
		assert.Equal(t, labels, deployment.Spec.Template.ObjectMeta.Labels)
		assert.Equal(t, podAnnotations, deployment.Spec.Template.ObjectMeta.Annotations)
		assert.Equal(t, nodeSelector, deployment.Spec.Template.Spec.NodeSelector)
		assert.Equal(t, tolerations, deployment.Spec.Template.Spec.Tolerations)
		assert.Equal(t, &metav1.LabelSelector{
			MatchLabels: labels,
		}, deployment.Spec.Selector)
		assert.Equal(t, &replicas, deployment.Spec.Replicas)
	})
}
