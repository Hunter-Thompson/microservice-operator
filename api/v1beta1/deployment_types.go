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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DeploymentSpec defines the desired state of Deployment
type DeploymentSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Ingress []Ingress `json:"ingress,omitempty"`
}

type Ingress struct {
	// +optional
	Host string `json:"host,omitempty"`
	// +optional
	InternetFacing bool `json:"internetFacing,omitempty"`
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
	// +optional
	Type Type `json:"type"`
	// +optional
	Paths         []string `json:"paths"`
	Name          string   `json:"name"`
	ContainerPort int32    `json:"containerPort"`
}

type Type string

const (
	HTTPS     Type = "HTTPS"
	GRPC      Type = "GRPC"
	HTTP      Type = "HTTP"
	TCP       Type = "TCP"
	WEBSOCKET Type = "WEBSOCKET"
)

// DeploymentStatus defines the observed state of Deployment
type DeploymentStatus struct {
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
type Deployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DeploymentSpec   `json:"spec,omitempty"`
	Status DeploymentStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DeploymentList contains a list of Deployment
type DeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Deployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Deployment{}, &DeploymentList{})
}
