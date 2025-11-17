package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:path=myresources,scope=Namespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// MyResource is the Hub version (v1)
type MyResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MyResourceSpec   `json:"spec,omitempty"`
	Status MyResourceStatus `json:"status,omitempty"`
}

// MyResourceSpec defines the desired state of MyResource
type MyResourceSpec struct {
	// Name is a required field
	Name string `json:"name"`

	// Replicas is the number of replicas
	Replicas int32 `json:"replicas"`

	// NewField is a field that only exists in v1
	NewField string `json:"newField,omitempty"`
}

// MyResourceStatus defines the observed state of MyResource
type MyResourceStatus struct {
	// Ready indicates if the resource is ready
	Ready *bool `json:"ready"`

	// Conditions represent the latest available observations
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// Hub marks this as the Hub version for conversion
func (*MyResource) Hub() {}
