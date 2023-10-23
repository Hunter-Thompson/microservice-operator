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
func (r *DeploymentReconciler) updateStatusReconciling(deployment *microservicev1beta1.Deployment, status microservicev1beta1.DeploymentStatus, reqLogger logr.Logger) error {
	status.State = microservicev1beta1.Reconciling
	return r.updateStatus(deployment, status, reqLogger)
}

func (r *DeploymentReconciler) updateStatus(deployment *microservicev1beta1.Deployment, status microservicev1beta1.DeploymentStatus, reqLogger logr.Logger) error {
	if reflect.DeepEqual(deployment.Status, status) {
		return nil
	}

	if deployment.Status.State != status.State {
		reqLogger.Info(fmt.Sprintf("Updating Deployment state from '%s' to '%s'", deployment.Status.State, status.State))
	}

	deployment.Status = status
	err := r.Client.Status().Update(context.TODO(), deployment)
	if err != nil {
		return errors.Wrap(err, "failed to update the Deployment status")
	}

	return nil
}

func (r *DeploymentReconciler) updateStatusReconcilingAndLogError(deployment *microservicev1beta1.Deployment, status microservicev1beta1.DeploymentStatus, reqLogger logr.Logger, statusErr error) {
	if statusErr != nil {
		status.Error = statusErr.Error()
	}

	err := r.updateStatusReconciling(deployment, status, reqLogger)
	if err != nil {
		reqLogger.Error(err, "Failed to set reconciling")
	}
}
