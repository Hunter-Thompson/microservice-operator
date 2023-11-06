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
	"os"
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"

	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/autoscaling/v2"
	networking "k8s.io/api/networking/v1"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func setupSuite(tb testing.TB) func(tb testing.TB) {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	assert.NoError(tb, err)

	err = microservicev1beta1.AddToScheme(scheme.Scheme)
	assert.NoError(tb, err)

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	assert.NoError(tb, err)

	return func(tb testing.TB) {
		err := testEnv.Stop()
		assert.NoError(tb, err)
	}
}

func prepareSchema(t *testing.T, scheme *runtime.Scheme) *runtime.Scheme {
	err := microservicev1beta1.AddToScheme(scheme)
	assert.NoError(t, err)

	return scheme
}

func setupTestDeps(t *testing.T) (logr.Logger, *MicroserviceReconciler) {
	s := prepareSchema(t, scheme.Scheme)
	r := MicroserviceReconciler{
		Client:    k8sClient,
		Scheme:    s,
		Resources: resources.NewResourceHelper(k8sClient, s),
	}

	logger := log.FromContext(context.TODO())

	logf.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	return logger, &r
}

func TestAllChecks(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	logger, r := setupTestDeps(t)

	msName := "foo"
	msNamespace := "default"

	ms := &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      msName,
			Namespace: msNamespace,
			UID:       types.UID("test"),
		},
	}

	currentStatus := microservicev1beta1.MicroserviceStatus{}

	t.Run("service", func(t *testing.T) {
		// ---
		err := r.checkService(ms, currentStatus, logger)
		assert.NoError(t, err)

		current := &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))

		// ---
		ms.Spec = microservicev1beta1.MicroserviceSpec{
			Ingress: []microservicev1beta1.Ingress{
				{
					ContainerPort: 8080,
					Name:          "test-1",
				},
			},
		}
		err = r.checkService(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, msName, current.GetName())
		assert.EqualValues(t, msNamespace, current.GetNamespace())
		assert.Equal(t, int32(8080), current.Spec.Ports[0].Port)
		assert.Equal(t, "test-1", current.Spec.Ports[0].Name)

		// ---
		ms.Spec = microservicev1beta1.MicroserviceSpec{
			Ingress: []microservicev1beta1.Ingress{
				{
					ContainerPort: 8090,
					Name:          "test-2",
				},
			},
		}
		err = r.checkService(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, msName, current.GetName())
		assert.EqualValues(t, msNamespace, current.GetNamespace())
		assert.Equal(t, int32(8090), current.Spec.Ports[0].Port)
		assert.Equal(t, "test-2", current.Spec.Ports[0].Name)

		// ---
		ms.Spec.Ingress = []microservicev1beta1.Ingress{}
		err = r.checkService(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))
	})

	t.Run("ingress", func(t *testing.T) {
		// ---
		err := r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current := &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))

		// ---
		ms.Spec.Ingress = []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
				Host:          "example.com",
			},
		}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))

		// ---
		pathType := networking.PathTypeImplementationSpecific
		ms.Spec.IngressEnabled = true

		ms.Spec.Ingress = []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
			},
		}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, []networking.IngressRule{
			{
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								PathType: &pathType,
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
						},
					},
				},
			},
		}, current.Spec.Rules)

		ms.Spec.Ingress = []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
				Host:          "example.com",
			},
		}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, []networking.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								PathType: &pathType,
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
						},
					},
				},
			},
		}, current.Spec.Rules)

		ms.Spec.Ingress = []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
				Host:          "example.com",
				Paths: []string{
					"/asd",
					"/dsa",
				},
			},
		}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, []networking.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								PathType: &pathType,
								Path:     "/asd",
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
							{
								PathType: &pathType,
								Path:     "/dsa",
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
						},
					},
				},
			},
		}, current.Spec.Rules)

		// ---
		ms.Spec.IngressEnabled = false

		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))

		// ---
		ms.Spec.IngressEnabled = true
		ms.Spec.Ingress = []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
				Host:          "example.com",
				Paths: []string{
					"/asd",
					"/dsa",
				},
			},
		}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.EqualValues(t, []networking.IngressRule{
			{
				Host: "example.com",
				IngressRuleValue: networking.IngressRuleValue{
					HTTP: &networking.HTTPIngressRuleValue{
						Paths: []networking.HTTPIngressPath{
							{
								PathType: &pathType,
								Path:     "/asd",
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
							{
								PathType: &pathType,
								Path:     "/dsa",
								Backend: networking.IngressBackend{
									Service: &networking.IngressServiceBackend{
										Name: "test-2",
										Port: networking.ServiceBackendPort{
											Number: int32(8090),
										},
									},
								},
							},
						},
					},
				},
			},
		}, current.Spec.Rules)

		ms.Spec.Ingress = []microservicev1beta1.Ingress{}
		err = r.checkIngress(ms, currentStatus, logger)
		assert.NoError(t, err)

		current = &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-2", Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))
	})

	t.Run("deployment", func(t *testing.T) {
		image := "image:latest"
		labels := map[string]string{
			"app": "test",
		}
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

		ingress := []microservicev1beta1.Ingress{
			{
				ContainerPort: 8090,
				Name:          "test-2",
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
			Ingress:        ingress,
		}

		err := r.checkDeployment(ms, currentStatus, logger)
		assert.NoError(t, err)

		current := &appsv1.Deployment{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.NoError(t, err)
		assert.Equal(t, []v1.Container{
			{
				Name:  ms.GetName(),
				Image: image,
				Ports: []v1.ContainerPort{
					{
						Name:          "test-2",
						ContainerPort: 8090,
						Protocol:      v1.ProtocolTCP,
					},
				},
				Env: []v1.EnvVar{
					{
						Name:  "test",
						Value: "env",
					},
				},
				TerminationMessagePath:   "/dev/termination-log",
				TerminationMessagePolicy: "File",
				ImagePullPolicy:          v1.PullAlways,
			},
		}, current.Spec.Template.Spec.Containers)

		assert.Equal(t, labels, current.Labels)
		assert.Equal(t, labels, current.Spec.Template.ObjectMeta.Labels)
		assert.Equal(t, podAnnotations, current.Spec.Template.ObjectMeta.Annotations)
		assert.Equal(t, nodeSelector, current.Spec.Template.Spec.NodeSelector)
		assert.Equal(t, tolerations, current.Spec.Template.Spec.Tolerations)
		assert.Equal(t, &metav1.LabelSelector{
			MatchLabels: labels,
		}, current.Spec.Selector)
		assert.Equal(t, &replicas, current.Spec.Replicas)
	})
}

func TestMicroserviceController(t *testing.T) {
	teardownSuite := setupSuite(t)
	defer teardownSuite(t)

	_, r := setupTestDeps(t)

	msName := "foo"
	msNamespace := "default"
	image := "image:latest"
	labels := map[string]string{
		"app": "test",
	}
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
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:          image,
			Labels:         labels,
			Replicas:       replicas,
			Env:            env,
			NodeSelector:   nodeSelector,
			Tolerations:    tolerations,
			IngressEnabled: true,
			PodAnnotations: podAnnotations,
			Autoscaling: &v2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: v2.CrossVersionObjectReference{
					APIVersion: "apps/v1",
					Kind:       "Deployment",
					Name:       "foo",
				},
				MinReplicas: &replicas,
				MaxReplicas: replicas,
			},
			Ingress: []microservicev1beta1.Ingress{
				{
					Host:          "example.com",
					Name:          "test-1",
					ContainerPort: 8090,
					Paths: []string{
						"/asd",
					},
				},
			},
		},
	}

	err := r.Client.Create(context.TODO(), ms)
	assert.NoError(t, err)

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: msName, Namespace: msNamespace}}

	result, err := r.Reconcile(context.TODO(), req)
	assert.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)

	svc := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, svc)
	assert.NoError(t, err)

	ing := &networking.Ingress{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName + "-test-1", Namespace: msNamespace}, ing)
	assert.NoError(t, err)

	deployment := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, deployment)
	assert.NoError(t, err)

	as := &v2.HorizontalPodAutoscaler{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, as)
	assert.NoError(t, err)
}
