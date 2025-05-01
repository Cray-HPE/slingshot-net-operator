/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package v1alpha1 for slingshot tenant provides types definition corresponding to its
// custom resource definition
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SlingshotTenantSpec defines the desired state of SlingshotTenant
type SlingshotTenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// TapmsTenantName is the name of the Tenant.
	TenantName string `json:"tenantname"`

	// TapmsTenantVersion specifies the version of the Tenant resource.
	TenantVersion string `json:"tenantversion,omitempty"`

	// IP is the IP address associated with the Tenant network.
	IP string `json:"ip,omitempty"`

	// Host specifies the hostname for the Tenant network.
	Host string `json:"host,omitempty"`

	// VNIPartition contains information about the VNI partition.
	VNIPartition VNIPartition `json:"vnipartition"`

	// VNIBlockName specifies the name of the VNI block.
	VNIBlockName string `json:"vniBlockName"`
}

// VNIPartition represents the VNI partition configuration for the Tenant network.
type VNIPartition struct {
	VNICount    int      `json:"vniCount,omitempty"`
	VNIRange    []string `json:"vniRanges,omitempty"`
	EdgePortDFA []int    `json:"edgePortDFA,omitempty"`
}

// SlingshotTenantStatus defines the observed state of SlingshotTenant
type SlingshotTenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Message provides a simple description of the current status of the SlingshotTenant resource.
	// This can be used to communicate the operational state to users.
	Message string `json:"message,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// SlingshotTenant is the Schema for the slingshottenants API
type SlingshotTenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SlingshotTenantSpec   `json:"spec,omitempty"`
	Status SlingshotTenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// SlingshotTenantList contains a list of SlingshotTenant
type SlingshotTenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SlingshotTenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SlingshotTenant{}, &SlingshotTenantList{})
}
