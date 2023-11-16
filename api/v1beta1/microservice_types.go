/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	corev1 "k8s.io/api/core/v1"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MicroserviceSpec defines the desired state of Microservice
type MicroserviceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Ingress []Ingress `json:"ingress,omitempty"`
	// +optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`
	// +optional
	Env   map[string]string `json:"env,omitempty"`
	Image string            `json:"image"`
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	Replicas       int32         `json:"replicas"`
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Labels    map[string]string           `json:"labels"`
	// +optional
	IngressEnabled bool `json:"ingressEnabled,omitempty"`
	// +optional
	Autoscaling *autoscalingv2.HorizontalPodAutoscalerSpec `json:"autoscaling,omitempty"`
	// +optional
	DisableServiceAccountCreation bool `json:"disableServiceAccountCreation,omitempty"`
}

type Ingress struct {
	// +optional
	Hosts []string `json:"host,omitempty"`
	// +optional
	Annotations   map[string]string `json:"annotations,omitempty"`
	Paths         []string          `json:"paths"`
	Name          string            `json:"name"`
	ContainerPort int32             `json:"containerPort"`
}

type Type string

const (
	HTTPS     Type = "HTTPS"
	GRPC      Type = "GRPC"
	HTTP      Type = "HTTP"
	TCP       Type = "TCP"
	WEBSOCKET Type = "WEBSOCKET"
)

// MicroserviceStatus defines the observed state of Microservice
type MicroserviceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Represents the running state of the Mattermost instance
	// +optional
	State RunningState `json:"state,omitempty"`
	// The last observed error in the deployment of this Mattermost instance
	// +optional
	Error string `json:"error,omitempty"`
}

// RunningState is the state of the Mattermost instance
type RunningState string

// Running States:
// Two types of instance running states are implemented: reconciling and stable.
// If any changes are being made on the mattermost instance, the state will be
// set to reconciling. If the reconcile loop reaches the end without requeuing
// then the state will be set to stable.
const (
	// Reconciling is the state when the Mattermost instance is being updated
	Reconciling RunningState = "reconciling"
	// Ready is the state when the Mattermost instance is ready to start serving
	// traffic but not fully stable.
	Ready RunningState = "ready"
	// Stable is the state when the Mattermost instance is fully running
	Stable RunningState = "stable"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Microservice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MicroserviceSpec   `json:"spec,omitempty"`
	Status MicroserviceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MicroserviceList contains a list of Microservice
type MicroserviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Microservice `json:"items"`
}

//func (d *Microservice) SetDefaults() error {
//	if d.Spec.Replicas == nil {
//		return fmt.Errorf("spec.replicas is required")
//	}
//
//	if d.Spec.Image == nil {
//		return fmt.Errorf("spec.image is required")
//	}
//
//	for _, ingress := range d.Spec.Ingress {
//		if ingress.ContainerPort == nil {
//
//		}
//
//	}
//
//	return nil
//}

func init() {
	SchemeBuilder.Register(&Microservice{}, &MicroserviceList{})
}
