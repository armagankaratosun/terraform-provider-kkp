// Package ssh_key_v2 implements the Terraform data source for KKP SSH keys.
package ssh_key_v2

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kapi "github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

var _ datasource.DataSource = &dataSourceSSHKeys{}
var _ datasource.DataSourceWithConfigure = &dataSourceSSHKeys{}

// NewDataSource creates a new SSH key v2 data source.
func NewDataSource() datasource.DataSource {
	return &dataSourceSSHKeys{}
}

func (d *dataSourceSSHKeys) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys_v2"
}

func (d *dataSourceSSHKeys) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Retrieves a list of SSH keys from KKP using V2 API.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"ssh_keys": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of SSH keys available in the project.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"id": dschema.StringAttribute{
							Computed:    true,
							Description: "SSH key ID.",
						},
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "SSH key name.",
						},
						"fingerprint": dschema.StringAttribute{
							Computed:    true,
							Description: "SSH key fingerprint.",
						},
						"public_key": dschema.StringAttribute{
							Computed:    true,
							Description: "SSH public key content.",
						},
						"creation_time": dschema.StringAttribute{
							Computed:    true,
							Description: "SSH key creation timestamp (RFC3339).",
						},
						"labels": dschema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "SSH key labels as key-value pairs.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceSSHKeys) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceSSHKeys) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.ValidateDataSourceBase(resp) {
		return
	}

	var config sshKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshKeysPayload, err := d.fetchSSHKeys()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list SSH keys", err.Error())
		return
	}

	if sshKeysPayload == nil {
		resp.Diagnostics.AddError("Invalid response", "SSH keys list response was empty")
		return
	}

	sshKeys := d.convertSSHKeysToSummaries(sshKeysPayload)

	state := sshKeysDataSourceModel{
		ID:      types.StringValue("ssh-keys-" + d.DefaultProjectID),
		SSHKeys: sshKeys,
	}

	tflog.Info(ctx, "successfully listed SSH keys", map[string]any{
		"project_id":    d.DefaultProjectID,
		"ssh_key_count": len(sshKeys),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *dataSourceSSHKeys) fetchSSHKeys() ([]*models.SSHKey, error) {
	pcli := kapi.New(d.Client.Transport, nil)
	params := kapi.NewListSSHKeysParams().WithProjectID(d.DefaultProjectID)
	resp, err := pcli.ListSSHKeys(params, nil)
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

func (d *dataSourceSSHKeys) convertSSHKeysToSummaries(sshKeys []*models.SSHKey) []sshKeySummary {
	summaries := make([]sshKeySummary, 0, len(sshKeys))

	for _, sshKey := range sshKeys {
		if sshKey == nil {
			continue
		}
		summaries = append(summaries, d.convertSSHKeyToSummary(sshKey))
	}

	// Sort SSH keys alphabetically by name for consistent ordering
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name.ValueString() < summaries[j].Name.ValueString()
	})

	return summaries
}

func (d *dataSourceSSHKeys) convertSSHKeyToSummary(sshKey *models.SSHKey) sshKeySummary {
	summary := sshKeySummary{
		ID:   types.StringValue(sshKey.ID),
		Name: types.StringValue(sshKey.Name),
	}

	// Set fingerprint if available
	if sshKey.Spec != nil && sshKey.Spec.Fingerprint != "" {
		summary.Fingerprint = types.StringValue(sshKey.Spec.Fingerprint)
	}

	// Set public key if available
	if sshKey.Spec != nil && sshKey.Spec.PublicKey != "" {
		summary.PublicKey = types.StringValue(sshKey.Spec.PublicKey)
	}

	// Set creation time if available
	if !sshKey.CreationTimestamp.IsZero() {
		summary.CreationTime = types.StringValue(sshKey.CreationTimestamp.String())
	}

	// Labels - simplified empty map for now
	summary.Labels = types.MapValueMust(types.StringType, map[string]attr.Value{})

	return summary
}
