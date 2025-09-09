package application_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Resource-specific types ----------

type dataSourceApplications struct {
	kkp.DataSourceBase
}

// ---------- Local state structs to read config/state ----------

type applicationsDataSourceModel struct {
	ID           types.String         `tfsdk:"id"`
	ClusterID    types.String         `tfsdk:"cluster_id"`
	Applications []applicationSummary `tfsdk:"applications"`
}

type applicationSummary struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Namespace          types.String `tfsdk:"namespace"`
	ClusterID          types.String `tfsdk:"cluster_id"`
	ApplicationName    types.String `tfsdk:"application_name"`
	ApplicationVersion types.String `tfsdk:"application_version"`
	CreationTime       types.String `tfsdk:"creation_time"`
	Status             types.String `tfsdk:"status"`
	Values             types.String `tfsdk:"values"`
}
