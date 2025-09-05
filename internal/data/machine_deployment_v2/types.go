package machine_deployment_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------- Resource-specific types ----------

type dataSourceMachineDeployments struct {
	kkp.DataSourceBase
}

// ---------- Local state structs to read config/state ----------

type machineDeploymentsDataSourceModel struct {
	ID                 types.String               `tfsdk:"id"`
	ClusterID          types.String               `tfsdk:"cluster_id"`
	MachineDeployments []machineDeploymentSummary `tfsdk:"machine_deployments"`
}

type machineDeploymentSummary struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ClusterID     types.String `tfsdk:"cluster_id"`
	CreationTime  types.String `tfsdk:"creation_time"`
	Replicas      types.Int64  `tfsdk:"replicas"`
	ReadyReplicas types.Int64  `tfsdk:"ready_replicas"`
	Labels        types.Map    `tfsdk:"labels"`
}
