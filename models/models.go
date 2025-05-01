/*
(C) Copyright Hewlett Packard Enterprise Development LP
*/

// Package models provides types definition
package models

import (
	"time"
)

const (
	//BaseURL is the base URL for Fabric Manager
	BaseURL = "https://api-gw-service-nmn.local/apis/fabric-manager"

	//WaitTime is the wait time in between the requests
	WaitTime = 100 * time.Millisecond

	//CAPublicKey is the path to the CA certificate
	CAPublicKey = "/var/run/configmap/ca-public-key.pem"

	OperatorConstFabric   = "/fabric"
	OperatorConstSwitches = "/switches/"
	OperatorConstPorts    = "/ports/"

	ReconciliationTime = 60 * time.Second
)

var (
	//AccessToken is the access token for Fabric Manager
	AccessToken string

	//NamespaceForClientData is the namespace for CSM client data
	NamespaceForClientData string

	//SecretForClientData is the secret for CSM client data
	SecretForClientData string

	//ClientID is the client ID for CSM client data
	ClientID string

	//SkipTLSVerify is the flag to skip TLS verification
	SkipTLSVerify string
)

// VNIRequestData defines the Payload for VNI configuration
type VNIRequestData struct {
	PartitionName string   `json:"partitionName,omitempty"`
	VNICount      int      `json:"vniCount,omitempty"`
	VNIRange      []string `json:"vniRanges,omitempty"`
	EdgePortDFA   []int    `json:"edgePortDFAs,omitempty"`
}

// VNIPartitionResponse defines the response for VNI configuration
type VNIPartitionResponse struct {
	PartitionName                string   `json:"partitionName"`
	VNICount                     int      `json:"vniCount"`
	VNIRange                     []string `json:"vniRanges"`
	EdgePortDFA                  []int    `json:"edgePortDFA"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentEpoch                int      `json:"documentEpoch"`
	DocumentKind                 string   `json:"documentKind"`
	DocumentSelfLink             string   `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string   `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
	Message                      string   `json:"message"`
	StatusCode                   int      `json:"statusCode"`
	ErrorCode                    int      `json:"errorCode"`
}

// AllVNIPartitionsResponse defines the response for all VNI partitions
type AllVNIPartitionsResponse struct {
	DocumentLinks                []string `json:"documentLinks"`
	DocumentCount                int      `json:"documentCount"`
	QueryTimeMicros              int      `json:"queryTimeMicros"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// Partition represents each dynamic partition key in the partition map
type Partition struct {
	VniRanges []VniRange `json:"vniRanges,omitempty"`
	VniCount  int        `json:"vniCount,omitempty"`
}

// VniRange represents a range of VNIs
type VniRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// SwitchesResponse defines the response for all switches
type SwitchesResponse struct {
	TotalCount                   int      `json:"totalCount"`
	DocumentLinks                []string `json:"documentLinks"`
	DocumentCount                int      `json:"documentCount"`
	QueryTimeMicros              int      `json:"queryTimeMicros"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// SwitchResponse defines the response for a switch
type SwitchResponse struct {
	IP                           string       `json:"IP"`
	GrpID                        int          `json:"grpId"`
	SwcNum                       int          `json:"swcNum"`
	FirmwareVersion              string       `json:"firmwareVersion"`
	FirmwareImage                string       `json:"firmwareImage"`
	SerialNumber                 string       `json:"serialNumber"`
	SwitchType                   string       `json:"switchType"`
	DisplayName                  string       `json:"displayName"`
	AgentLink                    string       `json:"agentLink"`
	EdgePortLinks                []string     `json:"edgePortLinks"`
	FabricPortLinks              []string     `json:"fabricPortLinks"`
	SwitchLabelList              []string     `json:"switchLabelList"`
	EdgePorts                    []EdgePort   `json:"edgePorts"`
	FabricPorts                  []FabricPort `json:"fabricPorts"`
	SwitchPolicyLink             string       `json:"switchPolicyLink"`
	DocumentVersion              int          `json:"documentVersion"`
	DocumentEpoch                int          `json:"documentEpoch"`
	DocumentKind                 string       `json:"documentKind"`
	DocumentSelfLink             string       `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int          `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string       `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int          `json:"documentExpirationTimeMicros"`
	DocumentOwner                string       `json:"documentOwner"`
	DocumentAuthPrincipalLink    string       `json:"documentAuthPrincipalLink"`
}

// EdgePort defines the edge port for a switch
type EdgePort struct {
	PortNum  int    `json:"portNum"`
	ConnPort string `json:"conn_port"`
}

// FabricPort defines the fabric port for a switch
type FabricPort struct {
	PortNum  int    `json:"portNum"`
	ConnPort string `json:"conn_port"`
}

// PortResponse defines the response for a port
type PortResponse struct {
	SwitchLink                   string   `json:"switchLink"`
	PortNumber                   int      `json:"portNumber"`
	ConnPort                     string   `json:"conn_port"`
	DstPort                      string   `json:"dst_port"`
	HsnIP                        string   `json:"hsnIp"`
	PortPolicyLinks              []string `json:"portPolicyLinks"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentEpoch                int      `json:"documentEpoch"`
	DocumentKind                 string   `json:"documentKind"`
	DocumentSelfLink             string   `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string   `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
	DocumentAuthPrincipalLink    string   `json:"documentAuthPrincipalLink"`
}

// DFAComponents defines the response for DFA components
type DFAComponents struct {
	GroupID       int             `json:"groupId,omitempty"`
	SwitchID      int             `json:"switchId,omitempty"`
	EdgePortsInfo []EdgePortsInfo `json:"edgePortsInfo,omitempty"`
}

// EdgePortsInfo defines the response for edge ports
type EdgePortsInfo struct {
	PortID   int    `json:"portNum"`
	EdgePort string `json:"conn_port"`
}

// VLANPortPolicyRequest defines the VLAN port policy
type VLANPortPolicyRequest struct {
	AllowedVlans      []string `json:"allowedVlans,omitempty"`
	NativeVlanID      string   `json:"nativeVlanId,omitempty"`
	IsUntaggedAllowed bool     `json:"isUntaggedAllowed,omitempty"`
	DocumentSelfLink  string   `json:"documentSelfLink,omitempty"`
}

// VLANPortPolicyResponse defines the response for VLAN port policy
type VLANPortPolicyResponse struct {
	AutoRetry                    AutoRetry      `json:"autoRetry"`
	HeadShellReset               HeadShellReset `json:"headShellReset"`
	AllowedVlans                 []string       `json:"allowedVlans"`
	NativeVlanID                 string         `json:"nativeVlanId"`
	IsUntaggedAllowed            bool           `json:"isUntaggedAllowed"`
	DocumentVersion              int            `json:"documentVersion"`
	DocumentEpoch                int            `json:"documentEpoch"`
	DocumentKind                 string         `json:"documentKind"`
	DocumentSelfLink             string         `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int            `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string         `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int            `json:"documentExpirationTimeMicros"`
	DocumentOwner                string         `json:"documentOwner"`
}

// AutoRetry defines the auto retry policy
type AutoRetry struct {
	Enabled     bool `json:"enabled"`
	Always      bool `json:"always"`
	NumRetries  int  `json:"num_retries"`
	DurationSec int  `json:"duration_sec"`
}

// HeadShellReset defines the head shell reset policy
type HeadShellReset struct {
	Enabled     bool `json:"enabled"`
	Always      bool `json:"always"`
	NumRetries  int  `json:"num_retries"`
	DurationSec int  `json:"duration_sec"`
}

// VLANsResponse defines the response for VLANs
type VLANsResponse struct {
	DocumentLinks                []string `json:"documentLinks"`
	DocumentCount                int      `json:"documentCount"`
	QueryTimeMicros              int      `json:"queryTimeMicros"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// VLANRequestData defines the payload for VLAN configuration
type VLANRequestData struct {
	VLANID   int    `json:"id,omitempty"`
	VLANName string `json:"name,omitempty"`
	Status   string `json:"status,omitempty"`
}

// VLANResponse defines the response for VLAN configuration
type VLANResponse struct {
	VLANID                       int    `json:"id"`
	VLANName                     string `json:"name"`
	Status                       string `json:"status"`
	DocumentVersion              int    `json:"documentVersion"`
	DocumentEpoch                int    `json:"documentEpoch"`
	DocumentKind                 string `json:"documentKind"`
	DocumentSelfLink             string `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int    `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int    `json:"documentExpirationTimeMicros"`
	DocumentOwner                string `json:"documentOwner"`
}

// PortPATCHRequest defines the payload for PATCH request
type PortPATCHRequest struct {
	PortPolicyLinks []string `json:"portPolicyLinks"`
}

// PortPoliciesResponse defines the response for port policies
type PortPoliciesResponse struct {
	DocumentLinks                []string `json:"documentLinks"`
	DocumentCount                int      `json:"documentCount"`
	QueryTimeMicros              int      `json:"queryTimeMicros"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// PortPolicyResponse defines the response for a policy
type PortPolicyResponse struct {
	AutoRetry                    AutoRetry      `json:"autoRetry"`
	HeadShellReset               HeadShellReset `json:"headShellReset"`
	AllowedVlans                 []string       `json:"allowedVlans"`
	NativeVlanID                 string         `json:"nativeVlanId"`
	IsUntaggedAllowed            bool           `json:"isUntaggedAllowed"`
	DocumentVersion              int            `json:"documentVersion"`
	DocumentEpoch                int            `json:"documentEpoch"`
	DocumentKind                 string         `json:"documentKind"`
	DocumentSelfLink             string         `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int            `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string         `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int            `json:"documentExpirationTimeMicros"`
	DocumentOwner                string         `json:"documentOwner"`
}

// VNIBlockRequestData defines the payload for VNI block configuration
type VNIBlockRequestData struct {
	VNIBlockName     string   `json:"vniBlockName,omitempty"`
	VNIPartitionName string   `json:"partitionName,omitempty"`
	VNIBlockRange    []string `json:"vniRanges,omitempty"`
	PortDFAs         []int    `json:"portDFAs,omitempty"`
}

// VNIBlockResponse defines the response for VNI block configuration
type VNIBlockResponse struct {
	VNIBlockName                 string   `json:"vniBlockName"`
	PartitionName                string   `json:"partitionName"`
	VNIBlockRange                []string `json:"vniRanges"`
	PortDFAs                     []int    `json:"portDFAs"`
	EnforcementTaskServiceLink   string   `json:"enforcementTaskServiceLink"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentEpoch                int      `json:"documentEpoch"`
	DocumentKind                 string   `json:"documentKind"`
	DocumentSelfLink             string   `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string   `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// VNIBlock PATCH Request
type VNIBlockPatchRequest struct {
	PortDFAs      []int    `json:"portDFAs"`
	VNIBlockRange []string `json:"vniRanges"`
}

// TokenResponse defines the response for token
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	Scope            string `json:"scope"`
}

// ErrorResponse defines the response for error
type ErrorResponse struct {
	Message      string `json:"message"`
	StatusCode   int    `json:"statusCode"`
	DocumentKind string `json:"documentKind"`
	ErrorCode    int    `json:"errorCode"`
}

// AllVNIBlocksResponse defines the response for all VNI blocks
type AllVNIBlocksResponse struct {
	DocumentLinks                []string `json:"documentLinks"`
	DocumentCount                int      `json:"documentCount"`
	QueryTimeMicros              int      `json:"queryTimeMicros"`
	DocumentVersion              int      `json:"documentVersion"`
	DocumentUpdateTimeMicros     int      `json:"documentUpdateTimeMicros"`
	DocumentExpirationTimeMicros int      `json:"documentExpirationTimeMicros"`
	DocumentOwner                string   `json:"documentOwner"`
}

// VniBlockEnforcementTaskServiceState defines the response for VNI block enforcement task
type VniBlockEnforcementTaskServiceState struct {
	EdgePortDFAs                 []int             `json:"edgePortDFAs"`
	RemoveEdgePortDFAs           []int             `json:"removeEdgePortDFAs"`
	VniList                      []int             `json:"vniList"`
	RemoveVniList                []int             `json:"removeVniList"`
	AddVniList                   []int             `json:"addVniList"`
	SwitchPortMap                map[string]string `json:"switchPortMap"`
	RemoveSwitchPortMap          map[string]string `json:"removeSwitchPortMap"`
	SubStage                     string            `json:"subStage"`
	TaskInfo                     TaskInfo          `json:"taskInfo"`
	DocumentVersion              int               `json:"documentVersion"`
	DocumentKind                 string            `json:"documentKind"`
	DocumentSelfLink             string            `json:"documentSelfLink"`
	DocumentUpdateTimeMicros     int               `json:"documentUpdateTimeMicros"`
	DocumentUpdateAction         string            `json:"documentUpdateAction"`
	DocumentExpirationTimeMicros int               `json:"documentExpirationTimeMicros"`
}

// TaskInfo defines the task information
type TaskInfo struct {
	Stage    string `json:"stage"`
	IsDirect bool   `json:"isDirect"`
}
