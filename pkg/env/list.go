package env

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (e *Env) List(w http.ResponseWriter, r *http.Request) {
	namespace := r.URL.Query().Get("namespace")

	microservices := &v1beta1.MicroserviceList{}
	err := e.Client.List(context.Background(), microservices, &client.ListOptions{
		Namespace: namespace,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(microservices.Items)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
