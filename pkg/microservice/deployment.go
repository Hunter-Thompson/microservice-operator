package microservice

import (
	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func GenerateDeployment(deployment *microservicev1beta1.Microservice) *appsv1.Deployment {
	desired := newDeployment(deployment)

	return configureDeployment(deployment, desired)
}

func newDeployment(deployment *microservicev1beta1.Microservice) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deployment.Name,
			Namespace:       deployment.Namespace,
			OwnerReferences: DeploymentOwnerReference(deployment),
			Labels:          deployment.Spec.Labels,
		},
	}
}

func configureDeployment(micdeployment *microservicev1beta1.Microservice, deployment *appsv1.Deployment) *appsv1.Deployment {
	revHistoryLimit := int32(defaultRevHistoryLimit)
	maxUnavailable := intstr.FromInt(defaultMaxUnavailable)
	maxSurge := intstr.FromInt(defaultMaxSurge)

	envVar := []v1.EnvVar{}
	for envName, env := range micdeployment.Spec.Env {
		envVar = append(envVar, v1.EnvVar{
			Name:  envName,
			Value: env,
		})
	}

	ports := []v1.ContainerPort{}
	for _, ingress := range micdeployment.Spec.Ingress {
		ports = append(ports, v1.ContainerPort{
			Name:          ingress.Name,
			ContainerPort: ingress.ContainerPort,
		})
	}

	deployment.Spec = appsv1.DeploymentSpec{
		Strategy: appsv1.DeploymentStrategy{
			Type: appsv1.RollingUpdateDeploymentStrategyType,
			RollingUpdate: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: &maxUnavailable,
				MaxSurge:       &maxSurge,
			},
		},
		RevisionHistoryLimit: &revHistoryLimit,
		Replicas:             &micdeployment.Spec.Replicas,
		Selector: &metav1.LabelSelector{
			MatchLabels: micdeployment.Spec.Labels,
		},
		Template: v1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: micdeployment.Spec.Labels,
			},
			Spec: v1.PodSpec{
				Tolerations:  micdeployment.Spec.Tolerations,
				NodeSelector: micdeployment.Spec.NodeSelector,
				Containers: []v1.Container{
					{
						Name:      micdeployment.Name,
						Image:     micdeployment.Spec.Image,
						Resources: micdeployment.Spec.Resources,
						Env:       envVar,
						Ports:     ports,
					},
				},
			},
		},
	}

	return deployment
}
