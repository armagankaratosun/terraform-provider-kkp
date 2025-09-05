package kkp

import (
	"net/http"
	"net/url"
	"time"

	httptransport "github.com/go-openapi/runtime/client"
)

// ProviderMeta is passed from provider.Configure to resources/datasources.
type ProviderMeta struct {
	Client           *Client
	DefaultProjectID string
}

// headerAuthWriter injects Authorization/User-Agent/extra headers on each request.
type headerAuthWriter struct {
	token       string
	userAgent   string
	extraHeader map[string]string
}

// Config holds connection/auth options for KKP.
type Config struct {
	// KKP base URL; with or without /api (we’ll normalize).
	// e.g. https://kkp.example.com or https://kkp.example.com/api
	Endpoint string

	// Bearer token for KKP REST.
	Token string

	// TLS
	InsecureSkipVerify bool
	CAFile             string // optional PEM bundle appended to system pool

	// HTTP
	Timeout      time.Duration     // default 60s
	UserAgent    string            // default "terraform-provider-kkp"
	ExtraHeaders map[string]string // optional extra headers for every request
}

// Client is a thin wrapper around the generated Go client.
type Client struct {
	// API is the generated client instance (branch-dependent type).
	// Avoid hardcoding the concrete type name here; resources can type assert.
	API any

	// Underlying pieces in case you need them.
	Transport  *httptransport.Runtime
	HTTPClient *http.Client
	BaseURL    *url.URL
}

// ClusterHealthChecker provides cluster health checking functionality
type ClusterHealthChecker struct {
	Client    *Client
	ProjectID string
	ClusterID string
}

// ClusterUpdateSpec holds the expected values after an update
type ClusterUpdateSpec struct {
	K8sVersion string
	CNIType    string
	CNIVersion string
}

// MachineDeploymentHealthChecker provides machine deployment health checking functionality
type MachineDeploymentHealthChecker struct {
	Client              *Client
	ProjectID           string
	ClusterID           string
	MachineDeploymentID string
	ExpectedReplicas    int64 // Optional: if set, wait for this many replicas instead of just matching desired==available
}

// ---------- Common Resource Types ----------

// ResourceBase provides common fields for all resource implementations
type ResourceBase struct {
	Client           *Client
	DefaultProjectID string
}

// DataSourceBase provides common fields for all data source implementations
type DataSourceBase struct {
	Client           *Client
	DefaultProjectID string
}

// ---------- Common Cloud Provider Types ----------

// CloudProvider represents supported cloud providers
type CloudProvider string

// Cloud provider constants
const (
	// CloudProviderOpenStack represents the OpenStack cloud provider.
	CloudProviderOpenStack CloudProvider = "openstack"
	// CloudProviderAWS represents the AWS cloud provider.
	CloudProviderAWS       CloudProvider = "aws"
	// CloudProviderVSphere represents the VSphere cloud provider.
	CloudProviderVSphere   CloudProvider = "vsphere"
	// CloudProviderAzure represents the Azure cloud provider.
	CloudProviderAzure     CloudProvider = "azure"
)

// EmptyCloudConfig represents placeholder for not-yet-implemented cloud providers
type EmptyCloudConfig struct{}

// AWS represents AWS-specific configuration (placeholder)
type AWS EmptyCloudConfig

// VSphere represents VSphere-specific configuration (placeholder)
type VSphere EmptyCloudConfig

// Azure represents Azure-specific configuration (placeholder)
type Azure EmptyCloudConfig

// ---------- Plan Interface ----------

// PlanValidator interface for resources that follow the common plan pattern
type PlanValidator interface {
	SetDefaults()
	Validate() error
}
