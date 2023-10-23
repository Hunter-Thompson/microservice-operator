package controllers

import (
	"context"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"
	"github.com/go-logr/logr"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *DeploymentReconciler) checkIngress(deployment *microservicev1beta1.Deployment, status microservicev1beta1.DeploymentStatus, reqLogger logr.Logger) error {
	return nil
}

func (r *DeploymentReconciler) checkService(deployment *microservicev1beta1.Deployment, status microservicev1beta1.DeploymentStatus, reqLogger logr.Logger) error {
	desired := microservice.GenerateServiceV1(deployment)

	err := r.Resources.CreateServiceIfNotExists(deployment, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &corev1.Service{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}
