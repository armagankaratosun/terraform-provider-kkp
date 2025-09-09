// Package provider implements the main Terraform provider for KKP.
package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	pframework "github.com/hashicorp/terraform-plugin-framework/provider"
	pschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	tflog "github.com/hashicorp/terraform-plugin-log/tflog"

	data_source_addon_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/addon_v2"
	data_source_application_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/application_v2"
	data_source_cluster_template_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/cluster_template_v2"
	data_source_cluster_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/cluster_v2"
	data_source_machine_deployment_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/machine_deployment_v2"
	data_source_ssh_key_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/data/ssh_key_v2"
	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	resource_addon_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/resources/addon_v2"
	resource_application_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/resources/application_v2"
	resource_cluster_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/resources/cluster_v2"
	resource_machine_deployment_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/resources/machine_deployment_v2"
	resource_ssh_key_v2 "github.com/armagankaratosun/terraform-provider-kkp/internal/resources/ssh_key_v2"
)

var _ pframework.Provider = &KKPProvider{}

// Config represents the provider configuration
type Config struct {
	Endpoint           tftypes.String `tfsdk:"endpoint"`
	Token              tftypes.String `tfsdk:"token"`
	InsecureSkipVerify tftypes.Bool   `tfsdk:"insecure_skip_verify"`
	ProjectID          tftypes.String `tfsdk:"project_id"`
}

// KKPProvider implements the KKP Terraform provider.
type KKPProvider struct {
	version string
}

// New creates a new KKP provider instance.
func New(version string) func() pframework.Provider {
	return func() pframework.Provider { return &KKPProvider{version: version} }
}

// Metadata returns the provider metadata.
func (p *KKPProvider) Metadata(_ context.Context, _ pframework.MetadataRequest, resp *pframework.MetadataResponse) {
	resp.TypeName = "kkp"
	resp.Version = p.version
}

// Schema returns the provider schema.
func (p *KKPProvider) Schema(_ context.Context, _ pframework.SchemaRequest, resp *pframework.SchemaResponse) {
	resp.Schema = pschema.Schema{
		Attributes: map[string]pschema.Attribute{
			"endpoint": pschema.StringAttribute{
				Required:    true,
				Description: "KKP API base URL (with or without /api), e.g. https://kkp.example.com or https://kkp.example.com/api",
			},
			"token": pschema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Bearer token for KKP API (use a project Service Account in the 'Editor' group to manage SSH keys).",
			},
			"insecure_skip_verify": pschema.BoolAttribute{
				Optional:    true,
				Description: "Skip TLS cert verification (dev/test only).",
			},
			"project_id": pschema.StringAttribute{
				Required:    true,
				Description: "Project ID used by all resources.",
			},
		},
	}
}

// Configure configures the KKP provider.
func (p *KKPProvider) Configure(ctx context.Context, req pframework.ConfigureRequest, resp *pframework.ConfigureResponse) {
	var cfg Config
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	endpoint := cfg.Endpoint.ValueString()
	token := cfg.Token.ValueString()
	projectID := cfg.ProjectID.ValueString()
	insecure := false
	if !cfg.InsecureSkipVerify.IsNull() && !cfg.InsecureSkipVerify.IsUnknown() {
		insecure = cfg.InsecureSkipVerify.ValueBool()
	}

	if endpoint == "" || token == "" || projectID == "" {
		resp.Diagnostics.AddError("Missing required provider settings", "'endpoint', 'token', and 'project_id' must be configured.")
		return
	}

	client, err := kkp.NewHTTPClient(kkp.Config{
		Endpoint:           endpoint,
		Token:              token,
		InsecureSkipVerify: insecure,
		Timeout:            60 * time.Second,
		UserAgent:          "terraform-provider-kkp/" + p.version,
		// CAFile:          cfg.CAFile.ValueString(), // if you expose it
	})
	if err != nil {
		resp.Diagnostics.AddError("KKP client initialization failed", err.Error())
		return
	}

	// Optional reachability check; failure is non-fatal.
	if err := client.Ping(ctx); err != nil {
		tflog.Warn(ctx, "KKP API ping failed (continuing)", map[string]any{
			"error":    err.Error(),
			"endpoint": endpoint,
		})
	}

	meta := &kkp.ProviderMeta{
		Client:           client,
		DefaultProjectID: projectID,
	}

	resp.ResourceData = meta
	resp.DataSourceData = meta
}

// Resources returns the provider's resources.
func (p *KKPProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resource_ssh_key_v2.New,
		resource_cluster_v2.New,
		resource_machine_deployment_v2.New,
		resource_addon_v2.New,
		resource_application_v2.New,
	}
}

// DataSources returns the provider's data sources.
func (p *KKPProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		data_source_addon_v2.NewDataSource,
		data_source_application_v2.NewDataSource,
		data_source_cluster_v2.NewDataSource,
		data_source_machine_deployment_v2.NewDataSource,
		data_source_ssh_key_v2.NewDataSource,
		data_source_cluster_template_v2.NewDataSource,
	}
}
