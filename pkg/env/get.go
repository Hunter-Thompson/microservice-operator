package env

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

var (
	ErrMicroserviceNotFound = "microservice not found"
)

func (e *Env) Get(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")
	name := r.URL.Query().Get("name")

	microservice := &v1beta1.Microservice{}

	err := e.Client.Get(context.Background(), types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, microservice)

	if err != nil && k8sErrors.IsNotFound(err) {
		http.Error(w, ErrMicroserviceNotFound, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(microservice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
