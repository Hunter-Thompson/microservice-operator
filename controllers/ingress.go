package controllers

import (
	"context"
	"strings"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	"github.com/Hunter-Thompson/microservice-operator/pkg/microservice"
	"github.com/Hunter-Thompson/microservice-operator/pkg/resources"
	"github.com/go-logr/logr"
	networking "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (r *MicroserviceReconciler) checkIngress(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	ingresses := networking.IngressList{}
	err := r.Client.List(context.TODO(), &ingresses, &client.ListOptions{
		Namespace: deployment.GetNamespace(),
	})
	if err != nil {
		return err
	}

	if len(deployment.Spec.Ingress) < 1 {
		for _, ing := range ingresses.Items {
			if strings.Contains(ing.GetName(), deployment.GetName()) {
				err := r.Resources.DeleteIngress(types.NamespacedName{Name: ing.GetName(), Namespace: ing.GetNamespace()}, reqLogger)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	if !deployment.Spec.IngressEnabled {
		for _, ing := range ingresses.Items {
			if strings.Contains(ing.GetName(), deployment.GetName()) {
				err := r.Resources.DeleteIngress(types.NamespacedName{Name: ing.GetName(), Namespace: ing.GetNamespace()}, reqLogger)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}

	desiredIngresses := microservice.GenerateIngressesV1(deployment)

	for _, desired := range desiredIngresses {
		err := r.Resources.CreateIngressIfNotExists(deployment, desired, reqLogger)
		if err != nil {
			return err
		}

		current := &networking.Ingress{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desired.Name, Namespace: desired.Namespace}, current)
		if err != nil {
			return err
		}

		err = r.Resources.Update(current, desired, reqLogger)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *MicroserviceReconciler) checkService(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	if len(deployment.Spec.Ingress) < 1 {
		return r.Resources.DeleteService(types.NamespacedName{Name: deployment.GetName(), Namespace: deployment.GetNamespace()}, reqLogger)
	}

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
