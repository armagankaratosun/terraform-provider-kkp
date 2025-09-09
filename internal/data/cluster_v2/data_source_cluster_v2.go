// Package cluster_v2 implements the Terraform data source for KKP clusters.
package cluster_v2

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"

	"github.com/kubermatic/go-kubermatic/models"
)

var _ datasource.DataSource = &dataSourceClusters{}
var _ datasource.DataSourceWithConfigure = &dataSourceClusters{}

// NewDataSource creates a new cluster v2 data source.
func NewDataSource() datasource.DataSource {
	return &dataSourceClusters{}
}

func (d *dataSourceClusters) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_clusters_v2"
}

func (d *dataSourceClusters) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Retrieves a list of clusters from KKP using V2 API.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"clusters": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "List of clusters available in the project.",
				NestedObject: dschema.NestedAttributeObject{
					Attributes: map[string]dschema.Attribute{
						"id": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster ID.",
						},
						"name": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster name.",
						},
						"creation_time": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster creation timestamp (RFC3339).",
						},
						"type": dschema.StringAttribute{
							Computed:    true,
							Description: "Cluster type (e.g., kubernetes).",
						},
						"cloud": dschema.StringAttribute{
							Computed:    true,
							Description: "Cloud provider (aws, gcp, azure, etc.).",
						},
						"datacenter_name": dschema.StringAttribute{
							Computed:    true,
							Description: "Datacenter where cluster is deployed.",
						},
						"version": dschema.StringAttribute{
							Computed:    true,
							Description: "Kubernetes version.",
						},
						"labels": dschema.MapAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "Cluster labels as key-value pairs.",
						},
					},
				},
			},
		},
	}
}

func (d *dataSourceClusters) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceClusters) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.ValidateDataSourceBase(resp) {
		return
	}

	var config clustersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clustersPayload, err := d.FetchClusters()
	if err != nil {
		resp.Diagnostics.AddError("Failed to list clusters", err.Error())
		return
	}

	if clustersPayload == nil {
		resp.Diagnostics.AddError("Invalid response", "Cluster list response was empty")
		return
	}

	clusters := convertClustersToSummaries(clustersPayload)

	state := clustersDataSourceModel{
		ID:       types.StringValue("clusters-" + d.DefaultProjectID),
		Clusters: clusters,
	}

	tflog.Info(ctx, "successfully listed clusters", map[string]any{
		"project_id":    d.DefaultProjectID,
		"cluster_count": len(clusters),
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func convertClustersToSummaries(clusters []*models.Cluster) []clusterSummary {
	summaries := make([]clusterSummary, 0, len(clusters))

	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}
		summaries = append(summaries, convertClusterToSummary(cluster))
	}

	// Sort clusters alphabetically by name for consistent ordering
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name.ValueString() < summaries[j].Name.ValueString()
	})

	return summaries
}

func convertClusterToSummary(cluster *models.Cluster) clusterSummary {
	summary := clusterSummary{
		ID:   types.StringValue(cluster.ID),
		Name: types.StringValue(cluster.Name),
	}

	// Set creation time if available
	if !cluster.CreationTimestamp.IsZero() {
		summary.CreationTime = types.StringValue(cluster.CreationTimestamp.String())
	}

	// Set type if available
	if cluster.Type != "" {
		summary.Type = types.StringValue(cluster.Type)
	}

	// Set cloud provider if available
	if cluster.Spec != nil && cluster.Spec.Cloud != nil {
		switch {
		case cluster.Spec.Cloud.Aws != nil:
			summary.Cloud = types.StringValue("aws")
		case cluster.Spec.Cloud.Azure != nil:
			summary.Cloud = types.StringValue("azure")
		case cluster.Spec.Cloud.Gcp != nil:
			summary.Cloud = types.StringValue("gcp")
		case cluster.Spec.Cloud.Openstack != nil:
			summary.Cloud = types.StringValue("openstack")
		case cluster.Spec.Cloud.Vsphere != nil:
			summary.Cloud = types.StringValue("vsphere")
		}
	}

	// Set datacenter name if available
	if cluster.Spec != nil && cluster.Spec.Cloud != nil && cluster.Spec.Cloud.DatacenterName != "" {
		summary.DatacenterName = types.StringValue(cluster.Spec.Cloud.DatacenterName)
	}

	// Set Kubernetes version if available
	if cluster.Spec != nil && string(cluster.Spec.Version) != "" {
		summary.Version = types.StringValue(string(cluster.Spec.Version))
	}

	// Convert cluster labels using centralized utility
	summary.Labels = kkp.ConvertLabelsToTerraform(cluster.Labels)

	return summary
}
