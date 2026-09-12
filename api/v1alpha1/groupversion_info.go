// Package v1alpha1 is the first version of our ClickHouseService API.
// +kubebuilder:object:generate=true
// +groupName=clickhouse.tilmancloud.io
package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

var (
	// GroupVersion is clickhouse.tilmancloud.io/v1alpha1 — not clickhouse.com.
	GroupVersion = schema.GroupVersion{Group: "clickhouse.tilmancloud.io", Version: "v1alpha1"}

	// SchemeBuilder registers our types with a runtime.Scheme.
	SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

	// AddToScheme is called from cmd/operator in the next commit.
	AddToScheme = SchemeBuilder.AddToScheme
)
