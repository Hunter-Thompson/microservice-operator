package env

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/controllers"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	cfg       *rest.Config
	k8sClient client.Client
	testEnv   *envtest.Environment
)

func setupSuite(tb testing.TB) func(tb testing.TB) {
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("../..", "config", "crd", "bases")},
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

func setupTestDeps(t *testing.T) (logr.Logger, *controllers.MicroserviceReconciler) {
	s := prepareSchema(t, scheme.Scheme)
	r := controllers.MicroserviceReconciler{
		Client:    k8sClient,
		Scheme:    s,
		Resources: resources.NewResourceHelper(k8sClient, s),
	}

	logger := log.FromContext(context.TODO())

	logf.SetLogger(zap.New(zap.WriteTo(os.Stdout), zap.UseDevMode(true)))

	return logger, &r
}

func TestCreateEnv(t *testing.T) {
	teardown := setupSuite(t)
	defer teardown(t)

	_, r := setupTestDeps(t)
	e := New(r.Client, resources.NewResourceHelper(r.Client, r.Scheme))

	cases := []struct {
		name         string
		code         int
		error        string
		microservice *microservicev1beta1.Microservice
	}{
		{
			name:  "working env creation",
			code:  http.StatusCreated,
			error: "",
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 1,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:  "working env creation with labels",
			code:  http.StatusCreated,
			error: "",
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "default",
					Labels: map[string]string{
						"app": "nginx",
					},
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 1,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:  "env already exists",
			code:  http.StatusConflict,
			error: ErrMicroserviceAlreadyExists,
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 1,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:  "missing labels",
			code:  http.StatusInternalServerError,
			error: "Microservice.microservice.microservice.example.com \"test3\" is invalid: spec.labels: Required value",
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test3",
					Namespace: "default",
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Replicas: 1,
					Image:    "nginx:latest",
				},
			},
		},
		{
			name:         "empty json",
			code:         http.StatusInternalServerError,
			error:        "resource name may not be empty",
			microservice: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {

			req := httptest.NewRequest(http.MethodPost, "/env/create", nil)
			req.Header.Set("Content-Type", "application/json")
			jb, err := json.Marshal(c.microservice)
			assert.NoError(t, err)
			req.Body = io.NopCloser(bytes.NewReader(jb))

			w := httptest.NewRecorder()

			e.Create(w, req)
			res := w.Result()

			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if c.error != "" {
				c.error = c.error + "\n"
			}

			assert.Equal(t, c.code, w.Code)
			assert.Equal(t, c.error, string(data))
		})
	}
}

func TestDeleteEnv(t *testing.T) {
	teardown := setupSuite(t)
	defer teardown(t)

	_, r := setupTestDeps(t)
	e := New(r.Client, resources.NewResourceHelper(r.Client, r.Scheme))

	err := r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
			Labels: map[string]string{
				"created-by": "microservice-operator",
			},
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	err = r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	cases := []struct {
		name             string
		code             int
		error            string
		nameAndNamespace *struct {
			Namespace string `json:"namespace"`
			Name      string `json:"name"`
		}
	}{
		{
			name:  "working env deletion",
			code:  http.StatusOK,
			error: "",
			nameAndNamespace: &struct {
				Namespace string "json:\"namespace\""
				Name      string "json:\"name\""
			}{
				Namespace: "default",
				Name:      "test1",
			},
		},
		{
			name:  "env not owned by operator",
			code:  http.StatusBadRequest,
			error: ErrMicroserviceNotOwnedByOperator,
			nameAndNamespace: &struct {
				Namespace string "json:\"namespace\""
				Name      string "json:\"name\""
			}{
				Namespace: "default",
				Name:      "test2",
			},
		},
		{
			name:  "does not exist",
			code:  http.StatusOK,
			error: "",
			nameAndNamespace: &struct {
				Namespace string "json:\"namespace\""
				Name      string "json:\"name\""
			}{
				Namespace: "default",
				Name:      "test3",
			},
		},
		{
			name:             "empty json",
			code:             http.StatusInternalServerError,
			nameAndNamespace: nil,
			error:            "resource name may not be empty",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/env/delete", nil)
			req.Header.Set("Content-Type", "application/json")
			jb, err := json.Marshal(c.nameAndNamespace)
			assert.NoError(t, err)
			req.Body = io.NopCloser(bytes.NewReader(jb))

			w := httptest.NewRecorder()

			e.Delete(w, req)
			res := w.Result()

			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if c.error != "" {
				c.error = c.error + "\n"
			}

			assert.Equal(t, c.code, w.Code)
			assert.Equal(t, c.error, string(data))
		})
	}
}

func TestGetEnv(t *testing.T) {
	teardown := setupSuite(t)
	defer teardown(t)

	_, r := setupTestDeps(t)
	e := New(r.Client, resources.NewResourceHelper(r.Client, r.Scheme))

	err := r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
			Labels: map[string]string{
				"created-by": "microservice-operator",
			},
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	cases := []struct {
		name     string
		code     int
		error    string
		query    string
		response *microservicev1beta1.Microservice
	}{
		{
			name:  "working env get",
			code:  http.StatusOK,
			error: "",
			query: "namespace=default&name=test1",
			response: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
					Labels: map[string]string{
						"created-by": "microservice-operator",
					},
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 1,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:     "does not exist",
			code:     http.StatusNotFound,
			error:    ErrMicroserviceNotFound,
			query:    "namespace=default&name=test2",
			response: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/env/get?"+c.query, nil)
			w := httptest.NewRecorder()

			e.Get(w, req)
			res := w.Result()

			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if c.error != "" {
				c.error = c.error + "\n"
			}

			assert.Equal(t, c.code, w.Code)

			if c.response != nil {
				m := microservicev1beta1.Microservice{}
				err = json.Unmarshal(data, &m)
				assert.NoError(t, err)

				assert.Equal(t, c.response.Name, m.Name)
				assert.Equal(t, c.response.Namespace, m.Namespace)
			} else {
				assert.Equal(t, c.error, string(data))
			}
		})
	}
}

func TestListEnv(t *testing.T) {
	teardown := setupSuite(t)
	defer teardown(t)

	_, r := setupTestDeps(t)
	e := New(r.Client, resources.NewResourceHelper(r.Client, r.Scheme))

	err := r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
			Labels: map[string]string{
				"created-by": "microservice-operator",
			},
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	err = r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "default",
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	cases := []struct {
		name  string
		code  int
		error string
		query string
	}{
		{
			name:  "working env list",
			code:  http.StatusOK,
			error: "",
			query: "namespace=default",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/env/list"+c.query, nil)
			w := httptest.NewRecorder()

			e.List(w, req)
			res := w.Result()

			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if c.error != "" {
				c.error = c.error + "\n"
			}

			assert.Equal(t, c.code, w.Code)

			microservices := []microservicev1beta1.Microservice{}
			err = json.Unmarshal(data, &microservices)
			assert.NoError(t, err)

			assert.Len(t, microservices, 2)
			assert.Equal(t, "test1", microservices[0].Name)
			assert.Equal(t, "default", microservices[0].Namespace)

			assert.Equal(t, "test2", microservices[1].Name)
			assert.Equal(t, "default", microservices[1].Namespace)

		})
	}
}

func TestEditEnv(t *testing.T) {

	teardown := setupSuite(t)
	defer teardown(t)

	_, r := setupTestDeps(t)

	e := New(r.Client, resources.NewResourceHelper(r.Client, r.Scheme))

	err := r.Client.Create(context.TODO(), &microservicev1beta1.Microservice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "default",
			Labels: map[string]string{
				"created-by": "microservice-operator",
			},
		},
		Spec: microservicev1beta1.MicroserviceSpec{
			Image:    "nginx:latest",
			Replicas: 1,
			Labels: map[string]string{
				"app": "nginx",
			},
		},
	})
	assert.NoError(t, err)

	cases := []struct {
		name         string
		code         int
		error        string
		microservice *microservicev1beta1.Microservice
	}{
		{
			name:  "working env edit",
			code:  http.StatusOK,
			error: "",
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test1",
					Namespace: "default",
					Labels: map[string]string{
						"created-by": "microservice-operator",
					},
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 2,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:  "does not exist",
			code:  http.StatusNotFound,
			error: ErrMicroserviceNotFound,
			microservice: &microservicev1beta1.Microservice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test2",
					Namespace: "default",
				},
				Spec: microservicev1beta1.MicroserviceSpec{
					Image:    "nginx:latest",
					Replicas: 2,
					Labels: map[string]string{
						"app": "nginx",
					},
				},
			},
		},
		{
			name:         "empty json",
			code:         http.StatusInternalServerError,
			error:        "resource name may not be empty",
			microservice: nil,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/env/edit", nil)
			req.Header.Set("Content-Type", "application/json")
			jb, err := json.Marshal(c.microservice)
			assert.NoError(t, err)
			req.Body = io.NopCloser(bytes.NewReader(jb))

			w := httptest.NewRecorder()

			e.Edit(w, req)
			res := w.Result()

			defer res.Body.Close()
			data, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			if c.error != "" {
				c.error = c.error + "\n"
			}

			assert.Equal(t, c.code, w.Code)
			assert.Equal(t, c.error, string(data))

			if c.code == http.StatusOK {
				m := microservicev1beta1.Microservice{}
				err = r.Client.Get(context.TODO(), client.ObjectKey{
					Namespace: c.microservice.Namespace,
					Name:      c.microservice.Name,
				}, &m)
				assert.NoError(t, err)

				assert.Equal(t, c.microservice.Name, m.Name)
				assert.Equal(t, c.microservice.Namespace, m.Namespace)
				assert.Equal(t, c.microservice.Spec.Replicas, m.Spec.Replicas)
			}

		})
	}

}
