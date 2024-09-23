package env

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (e *Env) Edit(w http.ResponseWriter, r *http.Request) {
	desired := &v1beta1.Microservice{}

	err := json.NewDecoder(r.Body).Decode(desired)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	current := &v1beta1.Microservice{}
	err = e.Client.Get(context.Background(), types.NamespacedName{
		Namespace: desired.Namespace,
		Name:      desired.Name,
	}, current)
	if err != nil && k8sErrors.IsNotFound(err) {
		http.Error(w, ErrMicroserviceNotFound, http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = e.Resources.UpdateWithoutLog(current, desired)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
