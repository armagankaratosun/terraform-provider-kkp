package addon_v2

import (
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Common fields ----------

// Plan represents the configuration plan for an addon resource.
type Plan struct {
	Name      string // Addon name (e.g., "prometheus", "grafana", etc.)
	ClusterID string // Required: which cluster to install the addon to

	// Addon configuration
	ContinuouslyReconcile bool        // Indicates that the addon cannot be deleted or modified outside of the UI after installation
	IsDefault             bool        // Indicates whether the addon is default
	Variables             interface{} // Free form data to use for parsing the manifest templates
}

// ---------- Resource-specific types ----------

type resourceAddon struct {
	kkp.ResourceBase
}

// ---------- Local state structs to read plan/state ----------

type addonState struct {
	ID                    tftypes.String `tfsdk:"id"`
	ClusterID             tftypes.String `tfsdk:"cluster_id"`
	Name                  tftypes.String `tfsdk:"name"`
	ContinuouslyReconcile tftypes.Bool   `tfsdk:"continuously_reconcile"`
	IsDefault             tftypes.Bool   `tfsdk:"is_default"`
	Variables             tftypes.String `tfsdk:"variables"` // JSON string for variables

	// Installation control
	WaitForReady   tftypes.Bool  `tfsdk:"wait_for_ready"`  // Wait for addon to be ready during creation
	TimeoutMinutes tftypes.Int64 `tfsdk:"timeout_minutes"` // Timeout for waiting (default: 2 minutes)

	// Status tracking fields
	Status        tftypes.String `tfsdk:"status"`         // Installation status: installing, ready, failed, deleting
	LastChecked   tftypes.String `tfsdk:"last_checked"`   // Last time status was checked (RFC3339)
	CreatedAt     tftypes.String `tfsdk:"created_at"`     // When addon was created (RFC3339)
	StatusMessage tftypes.String `tfsdk:"status_message"` // Additional status details
}
