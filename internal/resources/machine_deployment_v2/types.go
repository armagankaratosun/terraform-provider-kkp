package machine_deployment_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------- Common fields (provider-agnostic) ----------

// Plan represents the configuration plan for a machine deployment resource.
type Plan struct {
	Name      string
	ClusterID string // Required: which cluster to deploy to
	Replicas  int32  // Number of worker nodes (default: 1)

	// Kubernetes version for worker nodes (defaults to cluster version if empty)
	K8sVersion string

	// Deployment settings
	MinReadySeconds int32
	Paused          bool

	// Autoscaling settings (optional)
	MinReplicas *int32 // Minimum number of replicas for autoscaling (1-1000)
	MaxReplicas *int32 // Maximum number of replicas for autoscaling (1-1000)

	// Cloud provider
	Cloud string // "openstack" | "aws" | "vsphere" | "azure"

	// Cloud-specific configurations - exactly one should be set
	OpenStack *OpenStack
	AWS       *AWS
	VSphere   *VSphere
	Azure     *Azure
}

// ---------- Per-cloud options ----------

// OpenStack represents OpenStack-specific machine deployment configuration.
type OpenStack struct {
	// Machine specifications
	Flavor string // OpenStack flavor (e.g., "m1.small", "standard.medium")
	Image  string // Image name or UUID

	// Networking
	UseFloatingIP bool // Whether to assign floating IP to nodes

	// Storage
	DiskSize int32 // Root disk size in GB

	// Optional: Availability zone
	AvailabilityZone string
}

// ---------- Cloud-specific types for machine deployment ----------

// AWS represents AWS-specific machine deployment configuration.
type AWS = kkp.AWS

// VSphere represents VSphere-specific machine deployment configuration.
type VSphere = kkp.VSphere

// Azure represents Azure-specific machine deployment configuration.
type Azure = kkp.Azure

// ---------- Resource-specific types ----------

type resourceMachineDeployment struct {
	kkp.ResourceBase
}

// ---------- Local state structs to read plan/state ----------

type blockOpenStack struct {
	Flavor           tftypes.String `tfsdk:"flavor"`
	Image            tftypes.String `tfsdk:"image"`
	UseFloatingIP    tftypes.Bool   `tfsdk:"use_floating_ip"`
	DiskSize         tftypes.Int64  `tfsdk:"disk_size"`
	AvailabilityZone tftypes.String `tfsdk:"availability_zone"`
}

type blockAWS struct{}
type blockVSphere struct{}
type blockAzure struct{}

type machineDeploymentState struct {
	ID              tftypes.String `tfsdk:"id"`
	ClusterID       tftypes.String `tfsdk:"cluster_id"`
	Name            tftypes.String `tfsdk:"name"`
	Replicas        tftypes.Int64  `tfsdk:"replicas"`
	K8sVersion      tftypes.String `tfsdk:"k8s_version"`
	MinReadySeconds tftypes.Int64  `tfsdk:"min_ready_seconds"`
	Paused          tftypes.Bool   `tfsdk:"paused"`
	MinReplicas     tftypes.Int64  `tfsdk:"min_replicas"`
	MaxReplicas     tftypes.Int64  `tfsdk:"max_replicas"`
	Cloud           tftypes.String `tfsdk:"cloud"`

	OpenStack *blockOpenStack `tfsdk:"openstack"`
	AWS       *blockAWS       `tfsdk:"aws"`
	VSphere   *blockVSphere   `tfsdk:"vsphere"`
	Azure     *blockAzure     `tfsdk:"azure"`
}
