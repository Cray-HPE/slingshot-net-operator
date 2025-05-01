/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package v1alpha2 defines Tenant with v1alpha2 version
package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TenantSpec defines the desired state of Tenant
type TenantSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ChildNamespaces []string          `json:"childnamespaces"`
	State           string            `json:"state,omitempty"`
	TenantName      string            `json:"tenantname"`
	TenantResources []TenantResources `json:"tenantresources"`
	TenantKMS       TenantKMS         `json:"tenantkms"`
}

// TenantKMS defines TenantKMS type
type TenantKMS struct {
	EnableKMS bool   `json:"enablekms"`
	KeyName   string `json:"keyname"`
	KeyType   string `json:"keytype"`
}

// TenantResources defines the desired state of Tenant resources
type TenantResources struct {
	EnforceExclusiveHSMGroups bool     `json:"enforceexclusivehsmgroups"`
	HSMGroupLabel             string   `json:"hsmgrouplabel,omitempty"`
	HSMPartitionName          string   `json:"hsmpartitionname,omitempty"`
	Type                      string   `json:"type"`
	XNames                    []string `json:"xnames"`
	ForcePowerOff             bool     `json:"forcepoweroff"`
}

// TenantStatus defines the observed state of Tenant
type TenantStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ChildNamespaces []string          `json:"childnamespaces,omitempty"`
	TenantResources []TenantResources `json:"tenantresources,omitempty"`
	TenantKMS       TenantKMSStatus   `json:"tenantkms"`
	UUID            string            `json:"uuid,omitempty"`
}

// TenantKMSStatus defines the status of the TenantKMS
type TenantKMSStatus struct {
	KeyName     string `json:"keyname"`
	KeyType     string `json:"keytype"`
	PublicKey   string `json:"publickey"`
	TransitName string `json:"transitname"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Tenant is the Schema for the tenants API
type Tenant struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TenantSpec   `json:"spec,omitempty"`
	Status TenantStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TenantList contains a list of Tenant
type TenantList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Tenant `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Tenant{}, &TenantList{})
}
