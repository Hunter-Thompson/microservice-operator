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
	"path/filepath"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"

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

	return logger, &r
}

func TestDeploymentController(t *testing.T) {
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
		err := r.checkService(ms, currentStatus, logger)
		assert.NoError(t, err)

		current := &corev1.Service{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: msName, Namespace: msNamespace}, current)
		assert.Error(t, err)
		assert.True(t, k8sErrors.IsNotFound(err))

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
	})
}
