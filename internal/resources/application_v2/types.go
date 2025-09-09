package application_v2

import (
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Common fields ----------

// Plan represents the configuration plan for an application resource.
type Plan struct {
	Name      string // Application installation name
	ClusterID string // Required: which cluster to install the application to
	Namespace string // Kubernetes namespace for installation

	// Application reference
	ApplicationName    string // Name of the application (from ApplicationDefinitions)
	ApplicationVersion string // Version to install

	// Configuration
	Values interface{} // Helm values or other configuration as JSON
}

// ---------- Resource-specific types ----------

type resourceApplication struct {
	kkp.ResourceBase
}

// ---------- Local state structs to read plan/state ----------

type applicationState struct {
	ID        tftypes.String `tfsdk:"id"`
	ClusterID tftypes.String `tfsdk:"cluster_id"`
	Name      tftypes.String `tfsdk:"name"`
	Namespace tftypes.String `tfsdk:"namespace"`

	// Application reference
	ApplicationName    tftypes.String `tfsdk:"application_name"`
	ApplicationVersion tftypes.String `tfsdk:"application_version"`

	// Configuration
	Values tftypes.String `tfsdk:"values"` // JSON string for values

	// Installation control
	WaitForReady   tftypes.Bool  `tfsdk:"wait_for_ready"`  // Wait for application to be ready during creation
	TimeoutMinutes tftypes.Int64 `tfsdk:"timeout_minutes"` // Timeout for waiting (default: 5 minutes)

	// Status tracking fields
	Status        tftypes.String `tfsdk:"status"`         // Installation status: installing, ready, failed, deleting
	LastChecked   tftypes.String `tfsdk:"last_checked"`   // Last time status was checked (RFC3339)
	CreatedAt     tftypes.String `tfsdk:"created_at"`     // When application was created (RFC3339)
	StatusMessage tftypes.String `tfsdk:"status_message"` // Additional status details
}
