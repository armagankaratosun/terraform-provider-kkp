package cluster_template_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Data source implementation holder
type dataSourceClusterTemplates struct {
	kkp.DataSourceBase
}

// Data model for the data source state
type clusterTemplatesDataSourceModel struct {
	ID        types.String        `tfsdk:"id"`
	Templates []clusterTemplateEl `tfsdk:"templates"`
}

// Template summary element
type clusterTemplateEl struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Scope types.String `tfsdk:"scope"`
}
