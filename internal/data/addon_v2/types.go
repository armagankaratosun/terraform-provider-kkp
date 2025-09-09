package addon_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Resource-specific types ----------

type dataSourceAddons struct {
	kkp.DataSourceBase
}

// ---------- Local state structs to read config/state ----------

type addonsDataSourceModel struct {
	ID        types.String          `tfsdk:"id"`
	ClusterID types.String          `tfsdk:"cluster_id"`
	Addons    []addonDataModel      `tfsdk:"addons"`
	Available []availableAddonModel `tfsdk:"available"`
}

type addonDataModel struct {
	ID                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	ContinuouslyReconcile types.Bool   `tfsdk:"continuously_reconcile"`
	IsDefault             types.Bool   `tfsdk:"is_default"`
	Variables             types.String `tfsdk:"variables"`
}

type availableAddonModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}
