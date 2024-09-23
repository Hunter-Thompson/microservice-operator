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
	ErrMicroserviceAlreadyExists = "microservice already exists"
)

func (e *Env) Create(w http.ResponseWriter, r *http.Request) {
	microservice := &v1beta1.Microservice{}

	err := json.NewDecoder(r.Body).Decode(microservice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = e.Client.Get(context.Background(), types.NamespacedName{
		Namespace: microservice.Namespace,
		Name:      microservice.Name,
	}, microservice)
	if err != nil && !k8sErrors.IsNotFound(err) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if err == nil {
		http.Error(w, ErrMicroserviceAlreadyExists, http.StatusConflict)
		return
	}

	if microservice.ObjectMeta.Labels == nil {
		microservice.ObjectMeta.Labels = make(map[string]string)
		microservice.ObjectMeta.Labels["created-by"] = "microservice-operator"
	} else {
		microservice.ObjectMeta.Labels["created-by"] = "microservice-operator"
	}

	err = e.Client.Create(context.Background(), microservice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
