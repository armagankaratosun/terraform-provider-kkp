package ssh_key_v2

import (
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

// ---------- Resource-specific types ----------

type dataSourceSSHKeys struct {
	kkp.DataSourceBase
}

// ---------- Local state structs to read config/state ----------

type sshKeysDataSourceModel struct {
	ID      types.String    `tfsdk:"id"`
	SSHKeys []sshKeySummary `tfsdk:"ssh_keys"`
}

type sshKeySummary struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Fingerprint  types.String `tfsdk:"fingerprint"`
	PublicKey    types.String `tfsdk:"public_key"`
	CreationTime types.String `tfsdk:"creation_time"`
	Labels       types.Map    `tfsdk:"labels"`
}
