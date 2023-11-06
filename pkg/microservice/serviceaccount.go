package microservice

import (
	"fmt"

	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GenerateServiceAccount(mic *microservicev1beta1.Microservice) *corev1.ServiceAccount {
	return newServiceAccount(mic)
}

func newServiceAccount(mic *microservicev1beta1.Microservice) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:            mic.Name,
			Namespace:       mic.Namespace,
			OwnerReferences: DeploymentOwnerReference(mic),
			Labels:          mic.Spec.Labels,
			Annotations:     mic.GetAnnotations(),
		},
	}
}

func GenerateServiceAccountSecret(mic *microservicev1beta1.Microservice) *corev1.Secret {
	return newServiceAccountSecret(mic)
}

func newServiceAccountSecret(mic *microservicev1beta1.Microservice) *corev1.Secret {
	annotations := map[string]string{}

	for k, v := range mic.GetAnnotations() {
		annotations[k] = v
	}

	annotations["kubernetes.io/service-account.name"] = mic.GetName()
	name := fmt.Sprintf("%s-sa", mic.GetName())

	return &corev1.Secret{
		Type: corev1.SecretTypeServiceAccountToken,
		ObjectMeta: metav1.ObjectMeta{
			Name:            name,
			Namespace:       mic.Namespace,
			OwnerReferences: DeploymentOwnerReference(mic),
			Labels:          mic.Spec.Labels,
			Annotations:     annotations,
		},
	}
}
