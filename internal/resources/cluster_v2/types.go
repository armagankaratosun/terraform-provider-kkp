package cluster_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------- Common fields (provider-agnostic) ----------

// CNI represents container network interface configuration.
type CNI struct {
	Type    string // "cilium", "canal", ...
	Version string // "v1.14", ...
}

// Plan represents the configuration plan for a cluster resource.
type Plan struct {
	Name       string
	K8sVersion string // e.g. "1.28.5"
	Datacenter string // e.g. "ewc-eumetsat"
	Preset     string // KKP preset/credential; MAY be empty when using app creds
	Cloud      string // "openstack" | "aws" | "vsphere" | "azure"

	CNI CNI

	// exactly one of these per Provider
	OpenStack *OpenStack
	AWS       *AWS
	VSphere   *VSphere
	Azure     *Azure
}

// ---------- Per-cloud options ----------

// OpenStack represents OpenStack-specific cluster configuration.
type OpenStack struct {
	// Auth option A (preset): set Plan.Preset and (optionally) UseToken=true
	UseToken bool // only meaningful with presets; ignored when app creds are provided

	// Auth option B (no preset): application credentials
	ApplicationCredentialID     string
	ApplicationCredentialSecret string

	// Domain (required for both auth options)
	Domain string // OpenStack domain name (e.g. "default")

	// Networking (often required either way)
	Network        string // neutron network name/ID
	SecurityGroups string // at least one when no preset
	SubnetID       string // IPv4 subnet ID
	FloatingIPPool string // external network name (e.g. "public")
}

// ---------- Cloud-specific types for cluster ----------

// AWS represents AWS-specific cluster configuration.
type AWS = kkp.AWS
// VSphere represents VSphere-specific cluster configuration.
type VSphere = kkp.VSphere
// Azure represents Azure-specific cluster configuration.
type Azure = kkp.Azure

// ---------- Resource-specific types ----------

type resourceCluster struct {
	kkp.ResourceBase
}

// ---------- Local state structs to read plan/state ----------

type stateOpenStack struct {
	UseToken                    tftypes.Bool   `tfsdk:"use_token"`
	ApplicationCredentialID     tftypes.String `tfsdk:"application_credential_id"`
	ApplicationCredentialSecret tftypes.String `tfsdk:"application_credential_secret"`
	Domain                      tftypes.String `tfsdk:"domain"`
	Network                     tftypes.String `tfsdk:"network"`
	SecurityGroups              tftypes.String `tfsdk:"security_groups"`
	SubnetID                    tftypes.String `tfsdk:"subnet_id"`
	FloatingIPPool              tftypes.String `tfsdk:"floating_ip_pool"`
}

type stateAWS struct{}
type stateVSphere struct{}
type stateAzure struct{}

type clusterState struct {
	ID         tftypes.String `tfsdk:"id"`
	Name       tftypes.String `tfsdk:"name"`
	K8sVersion tftypes.String `tfsdk:"k8s_version"`
	Datacenter tftypes.String `tfsdk:"datacenter"`
	Preset     tftypes.String `tfsdk:"preset"`
	Cloud      tftypes.String `tfsdk:"cloud"`
	CNIType    tftypes.String `tfsdk:"cni_type"`
	CNIVersion tftypes.String `tfsdk:"cni_version"`

	OpenStack *stateOpenStack `tfsdk:"openstack"`
	AWS       *stateAWS       `tfsdk:"aws"`
	VSphere   *stateVSphere   `tfsdk:"vsphere"`
	Azure     *stateAzure     `tfsdk:"azure"`
}
