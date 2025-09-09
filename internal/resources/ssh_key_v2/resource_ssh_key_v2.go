// Package ssh_key_v2 implements the Terraform resource for KKP SSH keys.
package ssh_key_v2

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"

	kapi "github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

var (
	_ resource.Resource                = &resourceSSHKey{}
	_ resource.ResourceWithConfigure   = &resourceSSHKey{}
	_ resource.ResourceWithImportState = &resourceSSHKey{}
)

// New creates a new SSH key v2 resource.
func New() resource.Resource { return &resourceSSHKey{} }

// fromAPIProjectSSHKey maps API SSHKey -> TF state (without project_id).
// Note: KKP usually doesn't echo the public key on reads; caller keeps it in state.
func fromAPIProjectSSHKey(p *models.SSHKey) projectSSHKeyState {
	if p == nil {
		return projectSSHKeyState{
			ID:        tftypes.StringNull(),
			Name:      tftypes.StringNull(),
			PublicKey: tftypes.StringNull(),
		}
	}
	return projectSSHKeyState{
		ID:        tftypes.StringValue(p.ID),
		Name:      tftypes.StringValue(p.Name),
		PublicKey: tftypes.StringNull(),
	}
}

func (r *resourceSSHKey) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_sshkey"
}

func (r *resourceSSHKey) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Project-scoped SSH key in Kubermatic Kubernetes Platform (KKP). Uses the provider-level project_id.",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "SSH key ID.",
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Human-friendly SSH key name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": rschema.StringAttribute{
				Required:    true,
				Description: "OpenSSH public key (one line: '<type> <base64> [comment]').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`^\S+\s+[A-Za-z0-9+/=]+(?:\s+.+)?\s*$`),
						"must be a valid OpenSSH public key",
					),
				},
			},
		},
	}
}

func (r *resourceSSHKey) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

func (r *resourceSSHKey) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ValidateResourceBase(resp) {
		return
	}

	plan, ok := kkp.ExtractPlan[projectSSHKeyState](ctx, req, resp)
	if !ok {
		return
	}

	pub := plan.PublicKey.ValueString()
	body := &models.SSHKey{
		Name: plan.Name.ValueString(),
		Spec: &models.SSHKeySpec{
			PublicKey: pub,
		},
	}

	pcli := kapi.New(r.Client.Transport, nil)
	params := kapi.NewCreateSSHKeyParams().
		WithProjectID(r.DefaultProjectID).
		WithKey(body)

	out, err := pcli.CreateSSHKey(params, nil)
	if err != nil {
		// Helpful hint for permission issues.
		if strings.Contains(err.Error(), "createSshKeyForbidden") {
			resp.Diagnostics.AddError(
				"Permission denied (403)",
				"The Service Account must be in the 'Editor' group to create SSH keys. "+
					"Recreate the SA as Editor (not Project Manager/Viewer) and retry.",
			)
			return
		}
		// Handle 409 Conflict - resource already exists with the same name.
		if strings.Contains(err.Error(), "[409]") || strings.Contains(err.Error(), "createSSHKey default") {
			resp.Diagnostics.AddError(
				"SSH key name already exists",
				fmt.Sprintf("An SSH key with the name '%s' already exists in this project. Please choose a different name or remove the existing key first.", plan.Name.ValueString()),
			)
			return
		}
		resp.Diagnostics.AddError("Create project SSH key failed", err.Error())
		return
	}

	state := fromAPIProjectSSHKey(out.Payload)
	state.PublicKey = tftypes.StringValue(pub)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceSSHKey) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ValidateResourceBaseRead(resp) {
		return
	}

	state, ok := kkp.ExtractState[projectSSHKeyState](ctx, req, resp)
	if !ok {
		return
	}

	keyID := state.ID.ValueString()
	if keyID == "" {
		resp.Diagnostics.AddError("Missing id", "State did not contain SSH key id.")
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	listParams := kapi.NewListSSHKeysParams().WithProjectID(r.DefaultProjectID)
	listOut, err := pcli.ListSSHKeys(listParams, nil)
	if err != nil {
		resp.Diagnostics.AddWarning("Read project SSH keys failed",
			fmt.Sprintf("project %q: %v", r.DefaultProjectID, err))
		resp.State.RemoveResource(ctx)
		return
	}

	var found *models.SSHKey
	for _, k := range listOut.Payload {
		if k != nil && k.ID == keyID {
			found = k
			break
		}
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := fromAPIProjectSSHKey(found)
	newState.PublicKey = state.PublicKey // preserve pubkey in state
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *resourceSSHKey) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	plan, ok := kkp.ExtractStateForUpdate[projectSSHKeyState](ctx, req, resp)
	if !ok {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourceSSHKey) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ValidateResourceBaseDelete(resp) {
		return
	}

	state, ok := kkp.ExtractStateForDelete[projectSSHKeyState](ctx, req, resp)
	if !ok {
		return
	}

	keyID := state.ID.ValueString()
	if keyID == "" {
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	del := kapi.NewDeleteSSHKeyParams().
		WithProjectID(r.DefaultProjectID).
		WithSSHKeyID(keyID)

	if _, err := pcli.DeleteSSHKey(del, nil); err != nil {
		resp.Diagnostics.AddWarning("Delete project SSH key warning",
			fmt.Sprintf("project %q key %q: %v", r.DefaultProjectID, keyID, err))
	}
}

func (r *resourceSSHKey) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Accept either "<key_id>" or "<project_id>/<key_id>"; we always keep only the key id in state.
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Unexpected import ID", "Expected '<key_id>'")
		return
	}
	// If the user provides "<project>/<key>", take the last part.
	if i := strings.LastIndex(id, "/"); i >= 0 {
		id = id[i+1:]
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}
