// Package cluster_kubeconfig_v2 implements a data source to fetch a cluster kubeconfig.
package cluster_kubeconfig_v2

import (
	"context"
	"encoding/base64"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kapi "github.com/kubermatic/go-kubermatic/client/project"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
)

var _ datasource.DataSource = &dataSourceClusterKubeconfig{}
var _ datasource.DataSourceWithConfigure = &dataSourceClusterKubeconfig{}

// NewDataSource creates a new cluster kubeconfig v2 data source.
func NewDataSource() datasource.DataSource { return &dataSourceClusterKubeconfig{} }

func (d *dataSourceClusterKubeconfig) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_kubeconfig_v2"
}

func (d *dataSourceClusterKubeconfig) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Fetch the user cluster kubeconfig via KKP API V2.",
		Attributes: map[string]dschema.Attribute{
			"id": dschema.StringAttribute{
				Computed:    true,
				Description: "Data source identifier.",
			},
			"cluster_id": dschema.StringAttribute{
				Required:    true,
				Description: "Cluster ID to fetch kubeconfig for.",
			},
			"content": dschema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The kubeconfig content as UTF-8 string.",
			},
			"content_base64": dschema.StringAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "The kubeconfig content base64-encoded.",
			},
		},
	}
}

func (d *dataSourceClusterKubeconfig) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.ConfigureDataSource(req, resp)
}

func (d *dataSourceClusterKubeconfig) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	if !d.ValidateDataSourceBase(resp) {
		return
	}

	var config kubeconfigDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterID := kkp.TrimmedStringValue(config.ClusterID)
	if clusterID == "" {
		resp.Diagnostics.AddError("Missing cluster_id", "cluster_id is required to fetch kubeconfig")
		return
	}

	// Call API
	pcli := kapi.New(d.Client.Transport, nil)
	var payload []byte
	params := kapi.NewGetClusterKubeconfigV2Params().
		WithProjectID(d.DefaultProjectID).
		WithClusterID(clusterID)
	out, err := pcli.GetClusterKubeconfigV2(params, nil)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch kubeconfig", err.Error())
		return
	}
	payload = out.GetPayload()

	// Map to state
	content := string(payload)
	state := kubeconfigDataSourceModel{
		ID:            types.StringValue("kubeconfig-" + clusterID),
		ClusterID:     types.StringValue(clusterID),
		Content:       types.StringValue(content),
		ContentBase64: types.StringValue(base64.StdEncoding.EncodeToString(payload)),
	}

	tflog.Info(ctx, "fetched cluster kubeconfig", map[string]any{
		"project_id": d.DefaultProjectID,
		"cluster_id": clusterID,
	})

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
