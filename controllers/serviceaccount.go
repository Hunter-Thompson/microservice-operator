package controllers

import (
	"context"
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"
	corev1 "k8s.io/api/core/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MicroserviceReconciler) checkServiceAccount(mic *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	if mic.Spec.DisableServiceAccountCreation {
		return r.Resources.DeleteServiceAccount(types.NamespacedName{Name: mic.GetName(), Namespace: mic.GetNamespace()}, reqLogger)
	}

	desired := microservice.GenerateServiceAccount(mic)
	err := r.Resources.CreateServiceAccountIfNotExists(mic, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}

func (r *MicroserviceReconciler) checkServiceAccountSecret(mic *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	if mic.Spec.DisableServiceAccountCreation {
		secretName := fmt.Sprintf("%s-sa", mic.GetName())
		return r.Resources.DeleteSecret(types.NamespacedName{Name: secretName, Namespace: mic.GetNamespace()}, reqLogger)
	}

	desired := microservice.GenerateServiceAccountSecret(mic)
	err := r.Resources.CreateSecretIfNotExists(mic, desired, reqLogger)
	if err != nil {
		return err
	}

	current := &corev1.Secret{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
	if err != nil {
		return err
	}

	return r.Resources.Update(current, desired, reqLogger)
}
