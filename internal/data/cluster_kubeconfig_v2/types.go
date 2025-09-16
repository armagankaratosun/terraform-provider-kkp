package cluster_kubeconfig_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// Data source implementation struct
type dataSourceClusterKubeconfig struct {
	kkp.DataSourceBase
}

// Terraform state/config model
type kubeconfigDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	ClusterID     types.String `tfsdk:"cluster_id"`
	Content       types.String `tfsdk:"content"`
	ContentBase64 types.String `tfsdk:"content_base64"`
}
