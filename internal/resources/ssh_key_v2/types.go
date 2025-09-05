package ssh_key_v2

import (
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
)

// ---------- Resource-specific types ----------

type resourceSSHKey struct {
	kkp.ResourceBase
}

// ---------- Local state structs to read plan/state ----------

type projectSSHKeyState struct {
	ID        tftypes.String `tfsdk:"id"`         // computed
	Name      tftypes.String `tfsdk:"name"`       // required
	PublicKey tftypes.String `tfsdk:"public_key"` // required
}
