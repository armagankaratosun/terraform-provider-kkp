// Package machine_deployment_v2 implements the Terraform data source for KKP machine deployments.
package machine_deployment_v2

import (
	"context"
	"errors"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	pcli "github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

var _ datasource.DataSource = &dataSourceMachineDeployments{}
var _ datasource.DataSourceWithConfigure = &dataSourceMachineDeployments{}

// NewDataSource creates a new machine deployment v2 data source.
func NewDataSource() datasource.DataSource {
	return &dataSourceMachineDeployments{}
}

func (d *dataSourceMachineDeployments) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_deployments_v2"
}

func (d *dataSourceMachineDeployments) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Retrieves a list of machine deployments from KKP using V2 API. Requires cluster_id to list machine deployments for a specific cluster.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"cluster_id": dschema.StringAttribute{
				Required:    true,
				Description: "Cluster ID to list machine deployments from.",
			},
			"machine_deployments": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of machine deployments in the specified cluster.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"id": dschema.StringAttribute{
							Computed:    true,
							Description: "Machine deployment ID.",
						},
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "Machine deployment name.",
						},
						"cluster_id": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster ID where machine deployment is located.",
						},
						"creation_time": dschema.StringAttribute{
							Computed:    true,
							Description: "Machine deployment creation timestamp (RFC3339).",
						},
						"replicas": dschema.Int64Attribute{
							Computed:    true,
							Description: "Desired number of replicas.",
						},
						"ready_replicas": dschema.Int64Attribute{
							Computed:    true,
							Description: "Number of ready replicas.",
						},
						"labels": dschema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Machine deployment labels as key-value pairs.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceMachineDeployments) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceMachineDeployments) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.ValidateDataSourceBase(resp) {
		return
	}

	var config machineDeploymentsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := d.validateClusterID(&config, resp)
	if clusterID == "" {
		return
	}

	machineDeployments, err := d.fetchMachineDeployments(clusterID)
	if err != nil {
		resp.Diagnostics.AddError("Failed to list machine deployments", err.Error())
		return
	}

	summaries := d.convertMachineDeploymentsToSummaries(machineDeployments, clusterID)
	
	state := machineDeploymentsDataSourceModel{
		ID:                 types.StringValue("machine-deployments-" + clusterID),
		ClusterID:          types.StringValue(clusterID),
		MachineDeployments: summaries,
	}

	tflog.Info(ctx, "successfully listed machine deployments", map[string]any{
		"project_id":               d.DefaultProjectID,
		"cluster_id":               clusterID,
		"machine_deployment_count": len(summaries),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// validateClusterID validates the cluster ID from the configuration and returns the trimmed value.
func (d *dataSourceMachineDeployments) validateClusterID(config *machineDeploymentsDataSourceModel, resp *datasource.ReadResponse) string {
	clusterID := kkp.TrimmedStringValue(config.ClusterID)
	if clusterID == "" {
		resp.Diagnostics.AddError("Missing cluster_id", "cluster_id is required to list machine deployments")
		return ""
	}
	return clusterID
}

// fetchMachineDeployments retrieves machine deployments from the API for the given cluster.
func (d *dataSourceMachineDeployments) fetchMachineDeployments(clusterID string) ([]*models.NodeDeployment, error) {
	pclient := pcli.New(d.Client.Transport, nil)
	params := pcli.NewListMachineDeploymentsParams().
		WithProjectID(d.DefaultProjectID).
		WithClusterID(clusterID)

	machineDeploymentsResp, err := pclient.ListMachineDeployments(params, nil)
	if err != nil {
		return nil, err
	}

	if machineDeploymentsResp.Payload == nil {
		return nil, errors.New("machine deployments list response was empty")
	}

	return machineDeploymentsResp.Payload, nil
}

// convertMachineDeploymentsToSummaries converts API model objects to summary objects.
func (d *dataSourceMachineDeployments) convertMachineDeploymentsToSummaries(machineDeployments []*models.NodeDeployment, clusterID string) []machineDeploymentSummary {
	summaries := make([]machineDeploymentSummary, 0, len(machineDeployments))
	
	for _, md := range machineDeployments {
		if md == nil {
			continue
		}
		summaries = append(summaries, d.convertMachineDeploymentToSummary(md, clusterID))
	}

	// Sort machine deployments alphabetically by name for consistent ordering
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name.ValueString() < summaries[j].Name.ValueString()
	})

	return summaries
}

// convertMachineDeploymentToSummary converts a single machine deployment to a summary object.
func (d *dataSourceMachineDeployments) convertMachineDeploymentToSummary(md *models.NodeDeployment, clusterID string) machineDeploymentSummary {
	summary := machineDeploymentSummary{
		ID:        types.StringValue(md.ID),
		Name:      types.StringValue(md.Name),
		ClusterID: types.StringValue(clusterID),
	}

	// Set creation time if available
	if !md.CreationTimestamp.IsZero() {
		summary.CreationTime = types.StringValue(md.CreationTimestamp.String())
	}

	// Set replica counts from spec and status
	if md.Spec != nil && md.Spec.Replicas != nil {
		summary.Replicas = types.Int64Value(int64(*md.Spec.Replicas))
	}

	// Set status replica counts - simplified
	if md.Status != nil {
		summary.ReadyReplicas = types.Int64Value(int64(md.Status.ReadyReplicas))
	}

	// Convert machine deployment labels - check both Spec and root level
	var labels map[string]string
	if md.Spec != nil && md.Spec.Template != nil && md.Spec.Template.Labels != nil {
		labels = md.Spec.Template.Labels
	}
	summary.Labels = kkp.ConvertLabelsToTerraform(labels)

	return summary
}
