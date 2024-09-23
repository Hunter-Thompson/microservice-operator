package env

import (
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func New(client client.Client, resources *resources.ResourceHelper) *Env {
	return &Env{client, resources}
}

type Env struct {
	client.Client
	Resources *resources.ResourceHelper
}
