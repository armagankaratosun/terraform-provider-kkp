package cluster_v2

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"

	kapi "github.com/kubermatic/go-kubermatic/client/project"
)

var (
	_ resource.Resource                     = &resourceCluster{}
	_ resource.ResourceWithConfigure        = &resourceCluster{}
	_ resource.ResourceWithImportState      = &resourceCluster{}
	_ resource.ResourceWithConfigValidators = &resourceCluster{}
)

// New creates a new cluster v2 resource.
func New() resource.Resource { return &resourceCluster{} }

func (r *resourceCluster) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_v2"
}

func (r *resourceCluster) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Create a KKP cluster in the provider-level project (modular cloud support).",
		Attributes: map[string]rschema.Attribute{
			"id": rschema.StringAttribute{
				Computed:    true,
				Description: "Cluster ID.",
			},
			"use_template": rschema.BoolAttribute{
				Optional:    true,
				Description: "When true with template_id set, create the cluster by instantiating a Cluster Template (V2).",
			},
			"template_id": rschema.StringAttribute{
				Optional:    true,
				Description: "Cluster Template ID to instantiate (used when use_template = true).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_name": rschema.StringAttribute{
				Optional:    true,
				Description: "Cluster Template name to instantiate (alternative to template_id). If both are set, template_id is used.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"template_replicas": rschema.Int64Attribute{
				Optional:    true,
				Description: "Number of template instances to create (default 1).",
			},
			"name": rschema.StringAttribute{
				Required:    true,
				Description: "Cluster name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"k8s_version": rschema.StringAttribute{
				Required:    true,
				Description: "Kubernetes version (e.g. 1.28.5).",
			},
			"datacenter": rschema.StringAttribute{
				Required:    true,
				Description: "KKP datacenter name (e.g. ewc-eumetsat).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"preset": rschema.StringAttribute{
				Optional:    true,
				Description: "KKP preset/credential name. Leave empty when using OpenStack application credentials.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud": rschema.StringAttribute{
				Required:    true,
				Description: "Target cloud: openstack | aws | vsphere | azure.",
				Validators: []validator.String{
					stringvalidator.OneOf(kkp.CloudOpenStack, kkp.CloudAWS, kkp.CloudVSphere, kkp.CloudAzure),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},

			"cni_type": rschema.StringAttribute{
				Optional:    true,
				Description: "CNI plugin type (default: cilium).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cni_version": rschema.StringAttribute{
				Optional:    true,
				Description: "CNI plugin version (default: v1.14).",
			},
		},
		Blocks: map[string]rschema.Block{
			"openstack": rschema.SingleNestedBlock{
				Attributes: map[string]rschema.Attribute{
					"use_token": rschema.BoolAttribute{
						Optional:    true,
						Description: "Use token-based auth from preset (default: true when preset is set). Ignored in app-credential flow.",
					},
					"application_credential_id": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OpenStack application credential ID (required when no preset).",
					},
					"application_credential_secret": rschema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "OpenStack application credential secret (required when no preset).",
					},
					"domain": rschema.StringAttribute{
						Optional:    true,
						Description: "OpenStack domain name (e.g. 'default'). Usually provided by preset.",
					},
					"network": rschema.StringAttribute{
						Optional:    true,
						Description: "Neutron network name or ID (required when no preset).",
					},
					"security_groups": rschema.StringAttribute{
						Optional:    true,
						Description: "Security group name (required when no preset).",
					},
					"subnet_id": rschema.StringAttribute{
						Optional:    true,
						Description: "IPv4 subnet ID (required when no preset).",
					},
					"floating_ip_pool": rschema.StringAttribute{
						Optional:    true,
						Description: "External network / Floating IP pool (required when no preset).",
					},
				},
			},
			"aws":     rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}},
			"vsphere": rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}},
			"azure":   rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}},
		},
	}
}

func (r *resourceCluster) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{cloudBlockMatchesProviderValidator{}}
}

func (r *resourceCluster) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

// nolint:gocyclo // Create handles multiple creation paths (direct and template-based) for clarity.
func (r *resourceCluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ValidateResourceBase(resp) {
		return
	}

	plan, ok := kkp.ExtractPlan[clusterState](ctx, req, resp)
	if !ok {
		return
	}

	// --- Optional: Template-based creation path ---
	useTemplate := false
	if !plan.UseTemplate.IsNull() && !plan.UseTemplate.IsUnknown() {
		useTemplate = plan.UseTemplate.ValueBool()
	}
	templateID := strings.TrimSpace(plan.TemplateID.ValueString())
	templateName := strings.TrimSpace(plan.TemplateName.ValueString())
	if useTemplate && templateID != "" {
		replicas := int64(1)
		if !plan.TemplateReplicas.IsNull() && !plan.TemplateReplicas.IsUnknown() {
			replicas = plan.TemplateReplicas.ValueInt64()
		}

		clusterName := strings.TrimSpace(plan.Name.ValueString())
		if clusterName == "" {
			resp.Diagnostics.AddError("Missing name", "When use_template = true, 'name' must match the cluster name of the template to resolve the created cluster.")
			return
		}

		pcli := kapi.New(r.Client.Transport, nil)

		// If only template_name provided, resolve it to an ID via ListClusterTemplates
		if templateID == "" && templateName != "" {
			lres, lerr := pcli.ListClusterTemplates(kapi.NewListClusterTemplatesParams().WithProjectID(r.DefaultProjectID), nil)
			if lerr != nil || lres == nil {
				if lerr != nil {
					resp.Diagnostics.AddError("Failed to resolve template by name", lerr.Error())
				} else {
					resp.Diagnostics.AddError("Failed to resolve template by name", "empty response")
				}
				return
			}
			matches := make([]string, 0)
			for _, t := range lres.Payload {
				if t == nil {
					continue
				}
				if strings.TrimSpace(t.Name) == templateName {
					matches = append(matches, t.ID)
				}
			}
			if len(matches) == 0 {
				resp.Diagnostics.AddError("Template not found", "No cluster template with the given name found in the project. Please check the name or use template_id.")
				return
			}
			if len(matches) > 1 {
				resp.Diagnostics.AddError("Template name is ambiguous", "Multiple templates share this name. Please specify template_id instead.")
				return
			}
			templateID = matches[0]
		}

		if templateID == "" {
			resp.Diagnostics.AddError("Missing template reference", "Provide either template_id or template_name when use_template = true.")
			return
		}
		_, err := pcli.CreateClusterTemplateInstance(
			kapi.NewCreateClusterTemplateInstanceParams().
				WithProjectID(r.DefaultProjectID).
				WithClusterTemplateID(templateID).
				WithBody(kapi.CreateClusterTemplateInstanceBody{Replicas: replicas}),
			nil,
		)
		if err != nil {
			resp.Diagnostics.AddError("Create cluster from template failed", err.Error())
			return
		}

		// Resolve cluster ID by polling ListClustersV2 for appearance of clusterName
		var clusterID string
		findErr := kkp.PollWithTimeout(ctx, 5*time.Second, 10*time.Minute, func(_ context.Context) (bool, error) {
			list, lerr := pcli.ListClustersV2(
				kapi.NewListClustersV2Params().WithProjectID(r.DefaultProjectID),
				nil,
			)
			if lerr != nil || list == nil || list.Payload == nil {
				return false, nil
			}
			if list.Payload != nil && list.Payload.Clusters != nil {
				for _, c := range list.Payload.Clusters {
					if c == nil {
						continue
					}
					if strings.TrimSpace(c.Name) == clusterName {
						clusterID = c.ID
						return true, nil
					}
				}
			}
			return false, nil
		})
		if findErr != nil || clusterID == "" {
			if findErr != nil {
				resp.Diagnostics.AddError("Created cluster not found", findErr.Error())
			} else {
				resp.Diagnostics.AddError("Created cluster not found", "Timed out waiting for cluster to appear with the expected name. Confirm the template's cluster name.")
			}
			return
		}

		// Wait for ready using the existing checker
		checker := &kkp.ClusterHealthChecker{
			Client:    r.Client,
			ProjectID: r.DefaultProjectID,
			ClusterID: clusterID,
		}
		if err := checker.WaitForClusterReady(ctx); err != nil {
			resp.Diagnostics.AddError("Cluster provisioning timed out", err.Error())
			return
		}

		// Persist minimal state; keep planned fields untouched except ID and Name
		state := plan
		state.ID = tftypes.StringValue(clusterID)
		// Refresh name from API for accuracy
		if got, gerr := pcli.GetClusterV2(kapi.NewGetClusterV2Params().WithProjectID(r.DefaultProjectID).WithClusterID(clusterID), nil); gerr == nil && got != nil && got.Payload != nil {
			state.Name = tftypes.StringValue(got.Payload.Name)
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	cp := Plan{
		Name:       plan.Name.ValueString(),
		K8sVersion: plan.K8sVersion.ValueString(),
		Datacenter: plan.Datacenter.ValueString(),
		Preset:     plan.Preset.ValueString(), // may be empty -> app-credential path
		Cloud:      plan.Cloud.ValueString(),
		CNI: CNI{
			Type:    plan.CNIType.ValueString(),
			Version: plan.CNIVersion.ValueString(),
		},
	}

	switch cp.Cloud {
	case kkp.CloudOpenStack:
		os := &OpenStack{}
		usingPreset := strings.TrimSpace(cp.Preset) != ""
		if usingPreset {
			os.UseToken = true
		}
		if plan.OpenStack != nil {
			if !plan.OpenStack.UseToken.IsNull() && !plan.OpenStack.UseToken.IsUnknown() {
				os.UseToken = plan.OpenStack.UseToken.ValueBool()
			}
			os.ApplicationCredentialID = strings.TrimSpace(plan.OpenStack.ApplicationCredentialID.ValueString())
			os.ApplicationCredentialSecret = strings.TrimSpace(plan.OpenStack.ApplicationCredentialSecret.ValueString())
			os.Domain = strings.TrimSpace(plan.OpenStack.Domain.ValueString())
			os.Network = strings.TrimSpace(plan.OpenStack.Network.ValueString())
			os.SubnetID = strings.TrimSpace(plan.OpenStack.SubnetID.ValueString())
			os.FloatingIPPool = strings.TrimSpace(plan.OpenStack.FloatingIPPool.ValueString())
			os.SecurityGroups = strings.TrimSpace(plan.OpenStack.SecurityGroups.ValueString())
		}
		cp.OpenStack = os
	case kkp.CloudAWS:
		cp.AWS = &AWS{}
	case kkp.CloudVSphere:
		cp.VSphere = &VSphere{}
	case kkp.CloudAzure:
		cp.Azure = &Azure{}
	}

	spec, err := cp.ToCreateSpec(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Cluster spec invalid", err.Error())
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	params := kapi.NewCreateClusterV2Params().
		WithProjectID(r.DefaultProjectID).
		WithBody(spec)

	out, err := pcli.CreateClusterV2(params, nil)
	if err != nil {
		if e, ok := err.(*kapi.CreateClusterV2Default); ok && e.Payload != nil {
			b, _ := json.MarshalIndent(e.Payload, "", "  ")
			resp.Diagnostics.AddError("Create cluster failed", string(b))
			return
		}
		resp.Diagnostics.AddError("Create cluster failed", err.Error())
		return
	}

	clusterID := out.Payload.ID

	// --------- Wait for cluster to become ready ----------
	checker := &kkp.ClusterHealthChecker{
		Client:    r.Client,
		ProjectID: r.DefaultProjectID,
		ClusterID: clusterID,
	}

	if err := checker.WaitForClusterReady(ctx); err != nil {
		resp.Diagnostics.AddError("Cluster provisioning timed out", err.Error())
		return
	}

	// Ready: persist state
	state := plan
	state.ID = tftypes.StringValue(clusterID)
	state.Name = tftypes.StringValue(out.Payload.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ValidateResourceBaseRead(resp) {
		return
	}

	state, ok := kkp.ExtractState[clusterState](ctx, req, resp)
	if !ok {
		return
	}
	id := strings.TrimSpace(state.ID.ValueString())
	if id == "" {
		resp.Diagnostics.AddError("Missing id", "State did not contain cluster id.")
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	get := kapi.NewGetClusterV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(id)
	got, err := pcli.GetClusterV2(get, nil)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read cluster failed", err.Error())
		return
	}

	state.ID = tftypes.StringValue(got.Payload.ID)
	state.Name = tftypes.StringValue(got.Payload.Name)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceCluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.ValidateResourceBaseUpdate(resp) {
		return
	}

	plan, ok := kkp.ExtractStateForUpdate[clusterState](ctx, req, resp)
	if !ok {
		return
	}

	var state clusterState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	if id == "" {
		resp.Diagnostics.AddError("Missing id", "State did not contain cluster id.")
		return
	}

	// Decide what changed
	wantVersion := strings.TrimSpace(plan.K8sVersion.ValueString())
	curVersion := strings.TrimSpace(state.K8sVersion.ValueString())
	needVersion := wantVersion != "" && wantVersion != curVersion

	wantCNIType := strings.TrimSpace(plan.CNIType.ValueString())
	wantCNIVer := strings.TrimSpace(plan.CNIVersion.ValueString())
	curCNIType := strings.TrimSpace(state.CNIType.ValueString())
	curCNIVer := strings.TrimSpace(state.CNIVersion.ValueString())
	needCNI := (wantCNIType != "" && wantCNIType != curCNIType) ||
		(wantCNIVer != "" && wantCNIVer != curCNIVer)

	// Nothing to change -> just keep state
	if !needVersion && !needCNI {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// ---- build minimal patch (matches KKP spec shape) ----
	spec := map[string]any{}
	if needVersion {
		spec["version"] = wantVersion
	}
	if needCNI {
		// KKP expects spec.cniPlugin.{type,version}
		spec["cniPlugin"] = map[string]any{
			"type":    wantCNIType,
			"version": wantCNIVer,
		}
	}
	patchBody := map[string]any{"spec": spec}

	pcli := kapi.New(r.Client.Transport, nil)

	_, err := pcli.PatchClusterV2(
		kapi.NewPatchClusterV2Params().
			WithProjectID(r.DefaultProjectID).
			WithClusterID(id).
			WithPatch(patchBody),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Patch cluster failed", err.Error())
		return
	}
	tflog.Info(ctx, "patch sent", map[string]any{
		"cluster_id": id,
		"version":    wantVersion,
		"cni_type":   wantCNIType,
		"cni_ver":    wantCNIVer,
	})

	// ---- wait for update to complete ----
	checker := &kkp.ClusterHealthChecker{
		Client:    r.Client,
		ProjectID: r.DefaultProjectID,
		ClusterID: id,
	}

	expectedSpec := kkp.ClusterUpdateSpec{
		K8sVersion: wantVersion,
		CNIType:    wantCNIType,
		CNIVersion: wantCNIVer,
	}

	if err := checker.WaitForClusterUpdated(ctx, expectedSpec); err != nil {
		resp.Diagnostics.AddError("Cluster update timed out", err.Error())
		return
	}

	// Success: write new state (preserve id)
	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *resourceCluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ValidateResourceBaseDelete(resp) {
		return
	}

	state, ok := kkp.ExtractStateForDelete[clusterState](ctx, req, resp)
	if !ok {
		return
	}
	id := strings.TrimSpace(state.ID.ValueString())
	if id == "" {
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	del := kapi.NewDeleteClusterV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(id)

	if _, err := pcli.DeleteClusterV2(del, nil); err != nil {
		// Deletion might still be progressing server-side; warn and continue to poll.
		resp.Diagnostics.AddWarning("Delete cluster warning", err.Error())
	}

	// --- wait for cluster deletion to complete ---
	checker := &kkp.ClusterHealthChecker{
		Client:    r.Client,
		ProjectID: r.DefaultProjectID,
		ClusterID: id,
	}

	if err := checker.WaitForClusterDeleted(ctx); err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			resp.Diagnostics.AddError("Cluster deletion timed out", err.Error())
			return
		}
		resp.Diagnostics.AddError("Cluster delete polling failed", err.Error())
		return
	}

	// Confirmed gone; remove from state.
	resp.State.RemoveResource(ctx)
}

func (r *resourceCluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		resp.Diagnostics.AddError("Unexpected import ID", "Expected '<cluster_id>'")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
}

// --- config validator: ensure cloud block matches `cloud` ---
type cloudBlockMatchesProviderValidator struct{}

func (cloudBlockMatchesProviderValidator) Description(context.Context) string {
	return "exactly one cloud block must be set and it must match `cloud`"
}

func (cloudBlockMatchesProviderValidator) MarkdownDescription(context.Context) string {
	return "exactly one of the cloud blocks must be set and it must match `cloud`"
}

func (cloudBlockMatchesProviderValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var cfg clusterState
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip cloud block validation when creating from template
	if !cfg.UseTemplate.IsNull() && !cfg.UseTemplate.IsUnknown() && cfg.UseTemplate.ValueBool() {
		return
	}

	cloud := strings.ToLower(cfg.Cloud.ValueString())
	usingPreset := strings.TrimSpace(cfg.Preset.ValueString()) != ""

	// Generic validation: cloud blocks are optional with preset, required without
	switch cloud {
	case "openstack":
		if cfg.OpenStack == nil && !usingPreset {
			resp.Diagnostics.AddError(
				"Cloud mismatch",
				"`cloud = openstack` but no `openstack { ... }` block is set. Either provide an `openstack {}` block or use a preset.",
			)
			return
		}
	case "aws":
		if cfg.AWS == nil && !usingPreset {
			resp.Diagnostics.AddError(
				"Cloud mismatch",
				"`cloud = aws` but no `aws { ... }` block is set. Either provide an `aws {}` block or use a preset.",
			)
			return
		}
	case "vsphere":
		if cfg.VSphere == nil && !usingPreset {
			resp.Diagnostics.AddError(
				"Cloud mismatch",
				"`cloud = vsphere` but no `vsphere { ... }` block is set. Either provide a `vsphere {}` block or use a preset.",
			)
			return
		}
	case "azure":
		if cfg.Azure == nil && !usingPreset {
			resp.Diagnostics.AddError(
				"Cloud mismatch",
				"`cloud = azure` but no `azure { ... }` block is set. Either provide an `azure {}` block or use a preset.",
			)
			return
		}
	default:
		if cloud != "" {
			resp.Diagnostics.AddWarning(
				"Limited validation",
				"Validation for cloud '"+cloud+"' isn't implemented in this version.",
			)
		}
	}
}
