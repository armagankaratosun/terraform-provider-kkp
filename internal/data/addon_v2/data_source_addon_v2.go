// Package addon_v2 implements the Terraform data source for KKP addons.
package addon_v2

import (
	"context"
	"sort"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	acli "github.com/kubermatic/go-kubermatic/client/addon"
)

var (
	_ datasource.DataSource              = &dataSourceAddons{}
	_ datasource.DataSourceWithConfigure = &dataSourceAddons{}
)

// NewDataSource creates a new addon v2 data source.
func NewDataSource() datasource.DataSource { return &dataSourceAddons{} }

func (d *dataSourceAddons) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addons_v2"
}

func (d *dataSourceAddons) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Retrieve information about installed and available addons for a KKP cluster using V2 API.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"cluster_id": dsschema.StringAttribute{
				Required:    true,
				Description: "Cluster ID to retrieve addons for.",
			},
			"addons": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of currently installed addons.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id": dsschema.StringAttribute{
							Computed:    true,
							Description: "Addon ID.",
						},
						"name": dsschema.StringAttribute{
							Computed:    true,
							Description: "Addon name.",
						},
						"continuously_reconcile": dsschema.BoolAttribute{
							Computed:    true,
							Description: "Whether the addon is continuously reconciled.",
						},
						"is_default": dsschema.BoolAttribute{
							Computed:    true,
							Description: "Whether the addon is a default addon.",
						},
						"variables": dsschema.StringAttribute{
							Computed:    true,
							Description: "Addon variables as JSON string.",
						},
					},
				},
			},
			"available": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of addons available for installation.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"name": dsschema.StringAttribute{
							Computed:    true,
							Description: "Available addon name.",
						},
						"description": dsschema.StringAttribute{
							Computed:    true,
							Description: "Addon description.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceAddons) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceAddons) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config addonsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := kkp.TrimmedStringValue(config.ClusterID)
	if clusterID == "" {
		resp.Diagnostics.AddError("Missing cluster ID", "cluster_id cannot be empty")
		return
	}

	aclient := acli.New(d.Client.Transport, nil)

	// Get installed addons
	listParams := acli.NewListAddonsV2Params().
		WithProjectID(d.DefaultProjectID).
		WithClusterID(clusterID)

	listResp, err := aclient.ListAddonsV2(listParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list installed addons", err.Error())
		return
	}

	// Convert installed addons
	var installedAddons []addonDataModel
	if listResp.Payload != nil {
		for _, addon := range listResp.Payload {
			if addon != nil {
				addonModel := addonDataModel{
					ID:   types.StringValue(addon.ID),
					Name: types.StringValue(addon.Name),
				}

				if addon.Spec != nil {
					addonModel.ContinuouslyReconcile = types.BoolValue(addon.Spec.ContinuouslyReconcile)
					addonModel.IsDefault = types.BoolValue(addon.Spec.IsDefault)

					// Convert variables to JSON string using shared utility
					if addon.Spec.Variables != nil {
						variablesJSON, err := kkp.VariablesToJSON(addon.Spec.Variables)
						if err == nil {
							addonModel.Variables = types.StringValue(variablesJSON)
						} else {
							addonModel.Variables = types.StringValue("{}")
						}
					} else {
						addonModel.Variables = types.StringValue("{}")
					}
				} else {
					addonModel.ContinuouslyReconcile = types.BoolValue(false)
					addonModel.IsDefault = types.BoolValue(false)
					addonModel.Variables = types.StringValue("{}")
				}

				installedAddons = append(installedAddons, addonModel)
			}
		}
	}

	// Get available addons
	availableParams := acli.NewListInstallableAddonsV2Params().
		WithProjectID(d.DefaultProjectID).
		WithClusterID(clusterID)

	availableResp, err := aclient.ListInstallableAddonsV2(availableParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list available addons", err.Error())
		return
	}

	// Convert available addons - AccessibleAddons is []string
	var availableAddons []availableAddonModel
	if availableResp.Payload != nil {
		// Sort the addon names to ensure consistent ordering
		addonNames := make([]string, len(availableResp.Payload))
		copy(addonNames, availableResp.Payload)
		sort.Strings(addonNames)

		for _, addonName := range addonNames {
			availableModel := availableAddonModel{
				Name:        types.StringValue(addonName),
				Description: types.StringValue(""), // No description available in V2 API
			}
			availableAddons = append(availableAddons, availableModel)
		}
	}

	// Sort installed addons by name to ensure consistent ordering
	sort.Slice(installedAddons, func(i, j int) bool {
		return installedAddons[i].Name.ValueString() < installedAddons[j].Name.ValueString()
	})

	tflog.Info(ctx, "retrieved addon information", map[string]any{
		"cluster_id":      clusterID,
		"installed_count": len(installedAddons),
		"available_count": len(availableAddons),
	})

	// Set the data
	state := addonsDataSourceModel{
		ID:        types.StringValue("addons-" + clusterID),
		ClusterID: config.ClusterID,
		Addons:    installedAddons,
		Available: availableAddons,
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
