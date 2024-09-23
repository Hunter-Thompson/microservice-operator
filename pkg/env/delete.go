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
	ErrMicroserviceNotOwnedByOperator = "this microservice was not created by the microservice-operator and cannot be deleted"
)

func (e *Env) Delete(w http.ResponseWriter, r *http.Request) {
	nameAndNamespace := struct {
		Namespace string `json:"namespace"`
		Name      string `json:"name"`
	}{}

	err := json.NewDecoder(r.Body).Decode(&nameAndNamespace)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	microservice := &v1beta1.Microservice{}
	err = e.Client.Get(context.Background(), types.NamespacedName{
		Namespace: nameAndNamespace.Namespace,
		Name:      nameAndNamespace.Name,
	}, microservice)
	if err != nil && k8sErrors.IsNotFound(err) {
		w.WriteHeader(http.StatusOK)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if microservice.GetLabels()["created-by"] != "microservice-operator" {
		http.Error(w, ErrMicroserviceNotOwnedByOperator, http.StatusBadRequest)
		return
	}

	err = e.Client.Delete(context.Background(), microservice)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
