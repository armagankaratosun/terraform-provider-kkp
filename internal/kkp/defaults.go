package kkp

// Default values used across resources
const (
	// CNI defaults
	DefaultCNIType    = "cilium"
	DefaultCNIVersion = "1.16.9"

	// Kubernetes defaults
	DefaultK8sVersion = "1.32.7"

	// Replica defaults and limits
	DefaultReplicas        = int64(1)
	MaxReplicas            = int64(100)
	MaxAutoscalingReplicas = int64(1000)

	// Disk defaults and limits
	DefaultDiskSize = int64(25)   // 25GB
	MaxDiskSize     = int64(1000) // 1TB

	// Application defaults
	DefaultNamespace = "default"

	// Cloud provider constants
	CloudOpenStack = "openstack"
	CloudAWS       = "aws"
	CloudVSphere   = "vsphere"
	CloudAzure     = "azure"

	// Status constants
	StatusFailed     = "failed"
	StatusReady      = "ready"
	StatusInstalling = "installing"

	// KKP compatibility (minor series)
	SupportedKKPMinor = "2.28"

	// Value constants
	NullValue = "null"
)

// SupportedCloudProviders lists all supported cloud providers.
var SupportedCloudProviders = []string{
	CloudOpenStack,
	CloudAWS,
	CloudVSphere,
	CloudAzure,
}
