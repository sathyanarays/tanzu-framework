// Copyright YEAR VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NewCapabilitySpec defines the desired state of NewCapability
type NewCapabilitySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NewCapability. Edit newcapability_types.go to remove/update
	GVRs []GVR `json:"gvks,omitempty"`
}

type GVR struct {
	APIVersion string `json:"apiVersion"`

	Kind string `json:"kind"`

	//+kubebuilder:validation:Optional
	Namespace *string `json:"namespace"`
	Name      string  `json:"name"`
}

// NewCapabilityStatus defines the observed state of NewCapability
type NewCapabilityStatus struct {
	Present bool `json:"present"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NewCapability is the Schema for the newcapabilities API
type NewCapability struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NewCapabilitySpec   `json:"spec,omitempty"`
	Status NewCapabilityStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NewCapabilityList contains a list of NewCapability
type NewCapabilityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NewCapability `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NewCapability{}, &NewCapabilityList{})
}
