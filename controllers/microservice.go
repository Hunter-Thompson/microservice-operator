package controllers

import (
	"context"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"
	"github.com/go-logr/logr"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MicroserviceReconciler) checkDeployment(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	desired := microservice.GenerateDeployment(deployment)

	err := r.Resources.CreateDeploymentIfNotExists(deployment, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}
