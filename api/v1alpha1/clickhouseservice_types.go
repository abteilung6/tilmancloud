package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClickHouseServiceSpec is empty for now.
// Identity is metadata.name + metadata.namespace. Image, replicas, and
// storage stay hardcoded in the reconciler until we have a reason to change them.
type ClickHouseServiceSpec struct{}

// ClickHouseServiceStatus is what the controller observed.
// ReadyReplicas and Conditions are reserved; nothing writes them yet.
type ClickHouseServiceStatus struct {
	// ReadyReplicas is ready pods on the owned StatefulSet.
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Conditions use the standard Kubernetes condition type. We use Ready.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=chs
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// ClickHouseService is one single-node ClickHouse instance.
// The controller (not this type) creates Service, ConfigMap, Secret, and StatefulSet.
type ClickHouseService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClickHouseServiceSpec   `json:"spec,omitempty"`
	Status ClickHouseServiceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClickHouseServiceList is a list of ClickHouseService.
type ClickHouseServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClickHouseService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClickHouseService{}, &ClickHouseServiceList{})
}
