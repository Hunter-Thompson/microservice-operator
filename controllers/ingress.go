package controllers

import (
	"context"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	"github.com/go-logr/logr"
	networking "k8s.io/api/networking/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MicroserviceReconciler) checkIngress(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	desired := microservice.GenerateIngressV1(deployment)

	err := r.Resources.CreateIngressIfNotExists(deployment, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &networking.Ingress{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}

func (r *MicroserviceReconciler) checkService(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
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

	resources.CopyServiceEmptyAutoAssignedFields(desired, current)

	return r.Resources.Update(current, desired, reqLogger)
}
