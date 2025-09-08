// Package cluster_template_v2 implements a Terraform data source for KKP cluster templates (V2 API).
package cluster_template_v2

import (
    "context"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"

    kapi "github.com/kubermatic/go-kubermatic/client/project"
)

var _ datasource.DataSource = &dataSourceClusterTemplates{}
var _ datasource.DataSourceWithConfigure = &dataSourceClusterTemplates{}

// NewDataSource creates a new cluster templates V2 data source.
func NewDataSource() datasource.DataSource { return &dataSourceClusterTemplates{} }

func (d *dataSourceClusterTemplates) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cluster_templates_v2"
}

func (d *dataSourceClusterTemplates) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = dschema.Schema{
        Description: "Retrieves cluster templates for the configured project using the KKP V2 API.",
        Attributes: map[string]dschema.Attribute{
            "id": dschema.StringAttribute{
                Computed:    true,
                Description: "Data source identifier.",
            },
            "templates": dschema.ListNestedAttribute{
                Computed:    true,
                Description: "List of cluster templates in the project.",
                NestedObject: dschema.NestedAttributeObject{
                    Attributes: map[string]dschema.Attribute{
                        "id": dschema.StringAttribute{
                            Computed:    true,
                            Description: "Cluster template ID.",
                        },
                        "name": dschema.StringAttribute{
                            Computed:    true,
                            Description: "Cluster template name.",
                        },
                        "scope": dschema.StringAttribute{
                            Computed:    true,
                            Description: "Template scope (e.g., project, global).",
                        },
                    },
                },
            },
        },
    }
}

func (d *dataSourceClusterTemplates) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    d.ConfigureDataSource(req, resp)
}

func (d *dataSourceClusterTemplates) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    if !d.ValidateDataSourceBase(resp) {
        return
    }

    var _cfg clusterTemplatesDataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &_cfg)...) // no inputs yet; keeps symmetry
    if resp.Diagnostics.HasError() {
        return
    }

    pcli := kapi.New(d.Client.Transport, nil)
    out, err := pcli.ListClusterTemplates(
        kapi.NewListClusterTemplatesParams().WithProjectID(d.DefaultProjectID),
        nil,
    )
    if err != nil {
        resp.Diagnostics.AddError("Failed to list cluster templates", err.Error())
        return
    }

    var items []clusterTemplateEl
    if out != nil {
        for _, tpl := range out.Payload {
            if tpl == nil {
                continue
            }
            el := clusterTemplateEl{
                ID:   types.StringValue(tpl.ID),
                Name: types.StringValue(tpl.Name),
            }
            if tpl.Scope != "" {
                el.Scope = types.StringValue(tpl.Scope)
            } else {
                el.Scope = types.StringNull()
            }
            items = append(items, el)
        }
    }

    state := clusterTemplatesDataSourceModel{
        ID:        types.StringValue("cluster-templates-" + d.DefaultProjectID),
        Templates: items,
    }

    tflog.Info(ctx, "listed cluster templates", map[string]any{
        "project_id":      d.DefaultProjectID,
        "template_count": len(items),
    })

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
