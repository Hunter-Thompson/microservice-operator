package microservice

import (
	microservicev1beta1 "github.com/Hunter-Thompson/microservice-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func DeploymentOwnerReference(deployment *microservicev1beta1.Microservice) []metav1.OwnerReference {
	return []metav1.OwnerReference{
		*metav1.NewControllerRef(deployment, schema.GroupVersionKind{
			Group:   microservicev1beta1.GroupVersion.Group,
			Version: microservicev1beta1.GroupVersion.Version,
			Kind:    "Microservice",
		}),
	}
}
