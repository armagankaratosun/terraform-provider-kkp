package cluster_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Resource-specific types ----------

type dataSourceClusters struct {
	kkp.DataSourceBase
}

// ---------- Local state structs to read config/state ----------

type clustersDataSourceModel struct {
	ID       types.String     `tfsdk:"id"`
	Clusters []clusterSummary `tfsdk:"clusters"`
}

type clusterSummary struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CreationTime   types.String `tfsdk:"creation_time"`
	Type           types.String `tfsdk:"type"`
	Cloud          types.String `tfsdk:"cloud"`
	DatacenterName types.String `tfsdk:"datacenter_name"`
	Version        types.String `tfsdk:"version"`
	Labels         types.Map    `tfsdk:"labels"`
}
