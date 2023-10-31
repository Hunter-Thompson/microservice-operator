package controllers

import (
	"context"
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
)

// updateStatusReconciling sets the Mattermost state to reconciling.
func (r *MicroserviceReconciler) updateStatusReconciling(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	status.State = microservicev1beta1.Reconciling
	return r.updateStatus(deployment, status, reqLogger)
}

func (r *MicroserviceReconciler) updateStatus(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger) error {
	if reflect.DeepEqual(deployment.Status, status) {
		return nil
	}

	if deployment.Status.State != status.State {
		reqLogger.Info(fmt.Sprintf("Updating Microservice state from '%s' to '%s'", deployment.Status.State, status.State))
	}

	deployment.Status = status
	err := r.Client.Status().Update(context.TODO(), deployment)
	if err != nil {
		return errors.Wrap(err, "failed to update the Microservice status")
	}

	return nil
}

func (r *MicroserviceReconciler) updateStatusReconcilingAndLogError(deployment *microservicev1beta1.Microservice, status microservicev1beta1.MicroserviceStatus, reqLogger logr.Logger, statusErr error) {
	if statusErr != nil {
		status.Error = statusErr.Error()
	}

	err := r.updateStatusReconciling(deployment, status, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Failed to set reconciling")
	}
}
