// +k8s:openapi-gen=true
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:onlyVerbs=create
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleResource represents a generic example resource resource.
//
// TEMPLATE NOTE: This is an example resource. Rename this to your custom resource type
// (e.g., DataProcessing, BackupJob, AnalyticsQuery) and customize the Spec and Status
// fields to match your use case.
//
// The genclient directives above control code generation:
// - nonNamespaced: makes this a cluster-scoped resource (remove for namespaced)
// - onlyVerbs=create: limits to POST requests (modify for full CRUD)
//
// Example usage:
//
//	apiVersion: example.example-org.io/v1alpha1
//	kind: ExampleResource
//	metadata:
//	  name: my-example
//	spec:
//	  name: "example-1"
//	  count: 5
//	  enabled: true
type ExampleResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleResourceSpec   `json:"spec"`
	Status ExampleResourceStatus `json:"status,omitempty"`
}

// ExampleResourceSpec defines the desired state of ExampleResource.
//
// TEMPLATE NOTE: Customize these fields for your use case.
// This example shows common field patterns, but you can add any fields needed
// for your resource (timeouts, priorities, configuration, etc.)
type ExampleResourceSpec struct {
	// Name is an example string field.
	//
	// +required
	Name string `json:"name"`

	// Count is an example integer field.
	// +optional
	// +kubebuilder:default=1
	Count int32 `json:"count,omitempty"`

	// Enabled is an example boolean field.
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`
}

// ExampleResourceStatus defines the observed state of ExampleResource.
//
// TEMPLATE NOTE: Customize the status fields to represent the state of your resource.
// Common patterns include: phase/state, conditions, observedGeneration, results, errors, etc.
type ExampleResourceStatus struct {
	// Phase represents the current phase of the resource.
	// +optional
	Phase string `json:"phase,omitempty"`

	// Message provides human-readable details about the current state.
	// +optional
	Message string `json:"message,omitempty"`

	// ObservedGeneration reflects the generation most recently observed by the controller.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ExampleResourceList is a list of ExampleResource objects
type ExampleResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ExampleResource `json:"items"`
}
