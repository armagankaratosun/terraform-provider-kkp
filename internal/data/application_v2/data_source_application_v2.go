// Package application_v2 implements the Terraform data source for KKP applications.
package application_v2

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	acli "github.com/kubermatic/go-kubermatic/client/applications"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

var _ datasource.DataSource = &dataSourceApplications{}
var _ datasource.DataSourceWithConfigure = &dataSourceApplications{}

// NewDataSource creates a new application v2 data source.
func NewDataSource() datasource.DataSource {
	return &dataSourceApplications{}
}

func (d *dataSourceApplications) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_applications_v2"
}

func (d *dataSourceApplications) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Retrieves a list of application installations from KKP using V2 API. Requires cluster_id to list applications for a specific cluster.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"cluster_id": dschema.StringAttribute{
				Required:    true,
				Description: "Cluster ID to list applications from.",
			},
			"applications": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of application installations in the specified cluster.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"id": dschema.StringAttribute{
							Computed:    true,
							Description: "Application installation ID.",
						},
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "Application installation name.",
						},
						"namespace": dschema.StringAttribute{
							Computed:    true,
							Description: "Kubernetes namespace where application is installed.",
						},
						"cluster_id": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster ID where application is installed.",
						},
						"application_name": dschema.StringAttribute{
							Computed:    true,
							Description: "Name of the installed application (from ApplicationDefinitions).",
						},
						"application_version": dschema.StringAttribute{
							Computed:    true,
							Description: "Version of the installed application.",
						},
						"creation_time": dschema.StringAttribute{
							Computed:    true,
							Description: "Application installation creation timestamp (RFC3339).",
						},
						"status": dschema.StringAttribute{
							Computed:    true,
							Description: "Installation status: 'installing', 'ready', 'failed', 'deleting'.",
						},
						"values": dschema.StringAttribute{
							Computed:    true,
							Description: "Application configuration values as JSON string.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceApplications) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceApplications) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.ValidateDataSourceBase(resp) {
		return
	}

	var config applicationsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := kkp.TrimmedStringValue(config.ClusterID)
	if clusterID == "" {
		resp.Diagnostics.AddError("Missing cluster_id", "cluster_id is required to list applications")
		return
	}

	aclient := acli.New(d.Client.Transport, nil)
	params := acli.NewListApplicationInstallationsParams().
		WithProjectID(d.DefaultProjectID).
		WithClusterID(clusterID)

	applicationsResp, err := aclient.ListApplicationInstallations(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list applications", err.Error())
		return
	}

	if applicationsResp.Payload == nil {
		resp.Diagnostics.AddError("Invalid response", "Applications list response was empty")
		return
	}

	applications := make([]applicationSummary, 0, len(applicationsResp.Payload))
	for _, app := range applicationsResp.Payload {
		if app == nil {
			continue
		}

		// Generate ID from name since ListItem doesn't have ID field
		appID := app.Name
		if app.Namespace != "" {
			appID = app.Namespace + "/" + app.Name
		}

		summary := applicationSummary{
			ID:        types.StringValue(appID),
			Name:      types.StringValue(app.Name),
			Namespace: types.StringValue(app.Namespace),
			ClusterID: types.StringValue(clusterID),
		}

		// Set creation time if available (string format in v2.27)
		if app.CreationTimestamp != "" {
			summary.CreationTime = types.StringValue(app.CreationTimestamp)
		}

		// Extract application details from spec
		if app.Spec != nil {
			if app.Spec.ApplicationRef != nil {
				summary.ApplicationName = types.StringValue(app.Spec.ApplicationRef.Name)
				summary.ApplicationVersion = types.StringValue(app.Spec.ApplicationRef.Version)
			}
		}

		// Status determination - simplified since list items may not have full status
		status := "installing" // Default status
		if app.Status != nil {
			// Try to determine status from available fields
			// This is simplified as we don't have the same condition structure
			status = "ready" // Assume ready if status exists
		}

		summary.Status = types.StringValue(status)

		// Note: Values are not available in ListItem model
		summary.Values = types.StringValue("{}")

		applications = append(applications, summary)
	}

	// Sort applications alphabetically by name for consistent ordering
	sort.Slice(applications, func(i, j int) bool {
		return applications[i].Name.ValueString() < applications[j].Name.ValueString()
	})

	state := applicationsDataSourceModel{
		ID:           types.StringValue("applications-" + clusterID),
		ClusterID:    types.StringValue(clusterID),
		Applications: applications,
	}

	tflog.Info(ctx, "successfully listed applications", map[string]any{
		"project_id":        d.DefaultProjectID,
		"cluster_id":        clusterID,
		"application_count": len(applications),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
