package machine_deployment_v2

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/armagankaratosun/terraform-provider-kkp/internal/kkp"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	tftypes "github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	kapi "github.com/kubermatic/go-kubermatic/client/project"
)

var (
	_ resource.Resource                     = &resourceMachineDeployment{}
	_ resource.ResourceWithConfigure        = &resourceMachineDeployment{}
	_ resource.ResourceWithImportState      = &resourceMachineDeployment{}
	_ resource.ResourceWithConfigValidators = &resourceMachineDeployment{}
)

// New creates a new machine deployment v2 resource.
func New() resource.Resource { return &resourceMachineDeployment{} }

func (r *resourceMachineDeployment) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_deployment_v2"
}

func (r *resourceMachineDeployment) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Create a KKP machine deployment for worker nodes in a cluster.",
		Attributes:  r.buildSchemaAttributes(),
		Blocks:      r.buildSchemaBlocks(),
	}
}

// buildSchemaAttributes builds the attributes for the machine deployment resource schema.
func (r *resourceMachineDeployment) buildSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id":                r.buildIDAttribute(),
		"cluster_id":        r.buildClusterIDAttribute(),
		"name":              r.buildNameAttribute(),
		"replicas":          r.buildReplicasAttribute(),
		"k8s_version":       r.buildK8sVersionAttribute(),
		"min_ready_seconds": r.buildMinReadySecondsAttribute(),
		"paused":            r.buildPausedAttribute(),
		"min_replicas":      r.buildMinReplicasAttribute(),
		"max_replicas":      r.buildMaxReplicasAttribute(),
		"cloud":             r.buildCloudAttribute(),
	}
}

// buildSchemaBlocks builds the blocks for the machine deployment resource schema.
func (r *resourceMachineDeployment) buildSchemaBlocks() map[string]rschema.Block {
	return map[string]rschema.Block{
		"openstack": r.buildOpenstackBlock(),
		"aws":       r.buildAWSBlock(),
		"vsphere":   r.buildVSphereBlock(),
		"azure":     r.buildAzureBlock(),
	}
}

// buildIDAttribute builds the id attribute.
func (r *resourceMachineDeployment) buildIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Machine deployment ID.",
	}
}

// buildClusterIDAttribute builds the cluster_id attribute.
func (r *resourceMachineDeployment) buildClusterIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Cluster ID to deploy machines to.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildNameAttribute builds the name attribute.
func (r *resourceMachineDeployment) buildNameAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Machine deployment name.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildReplicasAttribute builds the replicas attribute.
func (r *resourceMachineDeployment) buildReplicasAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Number of worker nodes (default: 1).",
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
			int64validator.AtMost(100),
		},
	}
}

// buildK8sVersionAttribute builds the k8s_version attribute.
func (r *resourceMachineDeployment) buildK8sVersionAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Kubernetes version for worker nodes (defaults to cluster version).",
	}
}

// buildMinReadySecondsAttribute builds the min_ready_seconds attribute.
func (r *resourceMachineDeployment) buildMinReadySecondsAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Minimum number of seconds for which a newly created machine should be ready.",
		Validators: []validator.Int64{
			int64validator.AtLeast(0),
		},
	}
}

// buildPausedAttribute builds the paused attribute.
func (r *resourceMachineDeployment) buildPausedAttribute() rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Whether the deployment is paused.",
	}
}

// buildMinReplicasAttribute builds the min_replicas attribute.
func (r *resourceMachineDeployment) buildMinReplicasAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Description: "Minimum number of replicas for autoscaling. When set, enables autoscaling.",
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
			int64validator.AtMost(1000),
		},
	}
}

// buildMaxReplicasAttribute builds the max_replicas attribute.
func (r *resourceMachineDeployment) buildMaxReplicasAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Description: "Maximum number of replicas for autoscaling. Must be >= min_replicas when autoscaling is enabled.",
		Validators: []validator.Int64{
			int64validator.AtLeast(1),
			int64validator.AtMost(1000),
		},
	}
}

// buildCloudAttribute builds the cloud attribute.
func (r *resourceMachineDeployment) buildCloudAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Target cloud: openstack | aws | vsphere | azure.",
		Validators: []validator.String{
			stringvalidator.OneOf("openstack", "aws", "vsphere", "azure"),
		},
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildOpenstackBlock builds the openstack configuration block.
func (r *resourceMachineDeployment) buildOpenstackBlock() rschema.SingleNestedBlock {
	return rschema.SingleNestedBlock{
		Attributes: map[string]rschema.Attribute{
			"flavor": rschema.StringAttribute{
				Required:    true,
				Description: "OpenStack flavor name (e.g. m1.small, standard.medium).",
			},
			"image": rschema.StringAttribute{
				Required:    true,
				Description: "OpenStack image name or UUID.",
			},
			"use_floating_ip": rschema.BoolAttribute{
				Optional:    true,
				Description: "Whether to assign floating IP to worker nodes.",
			},
			"disk_size": rschema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Root disk size in GB (default: 25).",
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
					int64validator.AtMost(1000),
				},
			},
			"availability_zone": rschema.StringAttribute{
				Optional:    true,
				Description: "OpenStack availability zone.",
			},
		},
	}
}

// buildAWSBlock builds the aws configuration block.
func (r *resourceMachineDeployment) buildAWSBlock() rschema.SingleNestedBlock {
	return rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}}
}

// buildVSphereBlock builds the vsphere configuration block.
func (r *resourceMachineDeployment) buildVSphereBlock() rschema.SingleNestedBlock {
	return rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}}
}

// buildAzureBlock builds the azure configuration block.
func (r *resourceMachineDeployment) buildAzureBlock() rschema.SingleNestedBlock {
	return rschema.SingleNestedBlock{Attributes: map[string]rschema.Attribute{}}
}

func (r *resourceMachineDeployment) ConfigValidators(context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{cloudBlockMatchesMachineDeploymentValidator{}}
}

func (r *resourceMachineDeployment) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

func (r *resourceMachineDeployment) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ValidateResourceBase(resp) {
		return
	}

	plan, ok := kkp.ExtractPlan[machineDeploymentState](ctx, req, resp)
	if !ok {
		return
	}

	// Convert plan to internal plan structure
	cp, err := r.convertPlanFromState(*plan, resp)
	if err != nil {
		return
	}

	// Create machine deployment
	machineDeploymentID, err := r.createMachineDeployment(ctx, cp, resp)
	if err != nil {
		return
	}

	// Build and persist final state
	r.buildAndPersistCreateState(ctx, *plan, machineDeploymentID, resp)
}

func (r *resourceMachineDeployment) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ValidateResourceBaseRead(resp) {
		return
	}

	state, ok := kkp.ExtractState[machineDeploymentState](ctx, req, resp)
	if !ok {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	clusterID := strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing machine deployment or cluster ID.")
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	get := kapi.NewGetMachineDeploymentParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithMachineDeploymentID(id)

	got, err := pcli.GetMachineDeployment(get, nil)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Read machine deployment failed", err.Error())
		return
	}

	// Update state from API response
	state.ID = tftypes.StringValue(got.Payload.ID)
	state.Name = tftypes.StringValue(got.Payload.Name)
	if got.Payload.Spec != nil {
		if got.Payload.Spec.Replicas != nil {
			state.Replicas = tftypes.Int64Value(int64(*got.Payload.Spec.Replicas))
		}
		if got.Payload.Spec.Template != nil && got.Payload.Spec.Template.Versions != nil {
			state.K8sVersion = tftypes.StringValue(got.Payload.Spec.Template.Versions.Kubelet)
		}
		state.Paused = tftypes.BoolValue(got.Payload.Spec.Paused)
	}
	// k8s_version, paused, and min_ready_seconds are preserved from existing state as they may not be reliably returned by the API

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceMachineDeployment) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.ValidateResourceBaseUpdate(resp) {
		return
	}

	plan, ok := kkp.ExtractStateForUpdate[machineDeploymentState](ctx, req, resp)
	if !ok {
		return
	}

	var state machineDeploymentState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	clusterID := strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing machine deployment or cluster ID.")
		return
	}

	// Detect changes that need to be applied
	changes := r.detectUpdateChanges(*plan, state)
	if !changes.hasChanges() {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Apply the changes
	if err := r.applyUpdateChanges(ctx, changes, clusterID, id, resp); err != nil {
		return
	}

	// Build and persist final state
	r.buildAndPersistUpdateState(ctx, *plan, state, clusterID, id, resp)
}

func (r *resourceMachineDeployment) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ValidateResourceBaseDelete(resp) {
		return
	}

	state, ok := kkp.ExtractStateForDelete[machineDeploymentState](ctx, req, resp)
	if !ok {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	clusterID := strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		return
	}

	pcli := kapi.New(r.Client.Transport, nil)
	del := kapi.NewDeleteMachineDeploymentParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithMachineDeploymentID(id)

	if _, err := pcli.DeleteMachineDeployment(del, nil); err != nil {
		resp.Diagnostics.AddWarning("Delete machine deployment warning", err.Error())
	}

	// Wait for machine deployment to be deleted
	checker := &kkp.MachineDeploymentHealthChecker{
		Client:              r.Client,
		ProjectID:           r.DefaultProjectID,
		ClusterID:           clusterID,
		MachineDeploymentID: id,
	}

	if err := checker.WaitForMachineDeploymentDeleted(ctx); err != nil {
		resp.Diagnostics.AddError("Machine deployment deletion timed out", err.Error())
		return
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceMachineDeployment) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: cluster_id:machine_deployment_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Unexpected import ID", "Expected 'cluster_id:machine_deployment_id'")
		return
	}

	clusterID := strings.TrimSpace(parts[0])
	mdID := strings.TrimSpace(parts[1])
	if clusterID == "" || mdID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Both cluster_id and machine_deployment_id must be non-empty")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), mdID)...)
}

// --- config validator: ensure cloud block matches `cloud` ---
type cloudBlockMatchesMachineDeploymentValidator struct{}

func (cloudBlockMatchesMachineDeploymentValidator) Description(context.Context) string {
	return "exactly one cloud block must be set and it must match `cloud`"
}

func (cloudBlockMatchesMachineDeploymentValidator) MarkdownDescription(context.Context) string {
	return "exactly one of the cloud blocks must be set and it must match `cloud`"
}

func (cloudBlockMatchesMachineDeploymentValidator) ValidateResource(
	ctx context.Context,
	req resource.ValidateConfigRequest,
	resp *resource.ValidateConfigResponse,
) {
	var cfg machineDeploymentState
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cloud := strings.ToLower(cfg.Cloud.ValueString())
	switch cloud {
	case "openstack":
		if cfg.OpenStack == nil {
			resp.Diagnostics.AddError(
				"Cloud mismatch",
				"`cloud = openstack` but no `openstack { ... }` block is set.",
			)
			return
		}
	default:
		if cloud != "" {
			resp.Diagnostics.AddWarning(
				"Limited validation",
				"Validation for cloud '"+cloud+"' isn't implemented in this version. "+
					"Only `openstack {}` is currently supported by the resource schema.",
			)
		}
	}
}

// convertPlanFromState converts the Terraform state plan to internal Plan structure.
func (r *resourceMachineDeployment) convertPlanFromState(plan machineDeploymentState, resp *resource.CreateResponse) (*Plan, error) {
	replicas, err := kkp.SafeInt32(plan.Replicas.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError("Invalid Replicas Value", err.Error())
		return nil, err
	}

	cp := &Plan{
		Name:      plan.Name.ValueString(),
		ClusterID: plan.ClusterID.ValueString(),
		Replicas:  replicas,
		Cloud:     plan.Cloud.ValueString(),
	}

	// Set optional fields
	if err := r.setPlanOptionalFields(plan, cp, resp); err != nil {
		return nil, err
	}

	// Configure cloud-specific settings
	if err := r.setPlanCloudConfig(plan, cp, resp); err != nil {
		return nil, err
	}

	return cp, nil
}

// setPlanOptionalFields sets optional fields in the plan from the state.
func (r *resourceMachineDeployment) setPlanOptionalFields(plan machineDeploymentState, cp *Plan, resp *resource.CreateResponse) error {
	if !plan.K8sVersion.IsNull() && !plan.K8sVersion.IsUnknown() {
		cp.K8sVersion = plan.K8sVersion.ValueString()
	}

	if !plan.MinReadySeconds.IsNull() && !plan.MinReadySeconds.IsUnknown() {
		minReadySeconds, err := kkp.SafeInt32(plan.MinReadySeconds.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Invalid MinReadySeconds Value", err.Error())
			return err
		}
		cp.MinReadySeconds = minReadySeconds
	}

	if !plan.Paused.IsNull() && !plan.Paused.IsUnknown() {
		cp.Paused = plan.Paused.ValueBool()
	}

	if !plan.MinReplicas.IsNull() && !plan.MinReplicas.IsUnknown() {
		minReplicas, err := kkp.SafeInt32(plan.MinReplicas.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Invalid MinReplicas Value", err.Error())
			return err
		}
		cp.MinReplicas = &minReplicas
	}

	if !plan.MaxReplicas.IsNull() && !plan.MaxReplicas.IsUnknown() {
		maxReplicas, err := kkp.SafeInt32(plan.MaxReplicas.ValueInt64())
		if err != nil {
			resp.Diagnostics.AddError("Invalid MaxReplicas Value", err.Error())
			return err
		}
		cp.MaxReplicas = &maxReplicas
	}

	return nil
}

// setPlanCloudConfig configures cloud-specific settings in the plan.
func (r *resourceMachineDeployment) setPlanCloudConfig(plan machineDeploymentState, cp *Plan, resp *resource.CreateResponse) error {
	switch cp.Cloud {
	case "openstack":
		return r.setOpenstackConfig(plan, cp, resp)
	case "aws":
		cp.AWS = &AWS{}
	case "vsphere":
		cp.VSphere = &VSphere{}
	case "azure":
		cp.Azure = &Azure{}
	}
	return nil
}

// setOpenstackConfig configures OpenStack-specific settings.
func (r *resourceMachineDeployment) setOpenstackConfig(plan machineDeploymentState, cp *Plan, resp *resource.CreateResponse) error {
	os := &OpenStack{}
	if plan.OpenStack != nil {
		os.Flavor = strings.TrimSpace(plan.OpenStack.Flavor.ValueString())
		os.Image = strings.TrimSpace(plan.OpenStack.Image.ValueString())

		if !plan.OpenStack.UseFloatingIP.IsNull() && !plan.OpenStack.UseFloatingIP.IsUnknown() {
			os.UseFloatingIP = plan.OpenStack.UseFloatingIP.ValueBool()
		}

		if !plan.OpenStack.DiskSize.IsNull() && !plan.OpenStack.DiskSize.IsUnknown() {
			diskSize, err := kkp.SafeInt32(plan.OpenStack.DiskSize.ValueInt64())
			if err != nil {
				resp.Diagnostics.AddError("Invalid DiskSize Value", err.Error())
				return err
			}
			os.DiskSize = diskSize
		}

		if !plan.OpenStack.AvailabilityZone.IsNull() && !plan.OpenStack.AvailabilityZone.IsUnknown() {
			os.AvailabilityZone = strings.TrimSpace(plan.OpenStack.AvailabilityZone.ValueString())
		}
	}
	cp.OpenStack = os
	return nil
}

// createMachineDeployment creates the machine deployment via API and waits for it to be ready.
func (r *resourceMachineDeployment) createMachineDeployment(ctx context.Context, cp *Plan, resp *resource.CreateResponse) (string, error) {
	spec, err := cp.ToMachineDeploymentSpec()
	if err != nil {
		resp.Diagnostics.AddError("Machine deployment spec invalid", err.Error())
		return "", err
	}

	pcli := kapi.New(r.Client.Transport, nil)
	params := kapi.NewCreateMachineDeploymentParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(cp.ClusterID).
		WithBody(spec)

	out, err := pcli.CreateMachineDeployment(params, nil)
	if err != nil {
		if e, ok := err.(*kapi.CreateMachineDeploymentDefault); ok && e.Payload != nil {
			b, _ := json.MarshalIndent(e.Payload, "", "  ")
			resp.Diagnostics.AddError("Create machine deployment failed", string(b))
			return "", err
		}
		resp.Diagnostics.AddError("Create machine deployment failed", err.Error())
		return "", err
	}

	machineDeploymentID := out.Payload.ID

	// Wait for machine deployment to become ready
	checker := &kkp.MachineDeploymentHealthChecker{
		Client:              r.Client,
		ProjectID:           r.DefaultProjectID,
		ClusterID:           cp.ClusterID,
		MachineDeploymentID: machineDeploymentID,
	}

	if err := checker.WaitForMachineDeploymentReady(ctx); err != nil {
		resp.Diagnostics.AddError("Machine deployment provisioning timed out", err.Error())
		return "", err
	}

	return machineDeploymentID, nil
}

// buildAndPersistCreateState builds the final state after creation and persists it.
func (r *resourceMachineDeployment) buildAndPersistCreateState(ctx context.Context, plan machineDeploymentState, machineDeploymentID string, resp *resource.CreateResponse) {
	// Get the created machine deployment to build accurate state
	pcli := kapi.New(r.Client.Transport, nil)
	getParams := kapi.NewGetMachineDeploymentParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(plan.ClusterID.ValueString()).
		WithMachineDeploymentID(machineDeploymentID)

	got, err := pcli.GetMachineDeployment(getParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read machine deployment after creation failed", err.Error())
		return
	}

	// Build final state
	state := plan
	state.ID = tftypes.StringValue(machineDeploymentID)
	state.Name = tftypes.StringValue(got.Payload.Name)

	if got.Payload.Spec != nil {
		if got.Payload.Spec.Replicas != nil {
			state.Replicas = tftypes.Int64Value(int64(*got.Payload.Spec.Replicas))
		}
		if got.Payload.Spec.Template != nil && got.Payload.Spec.Template.Versions != nil {
			state.K8sVersion = tftypes.StringValue(got.Payload.Spec.Template.Versions.Kubelet)
		}
		state.Paused = tftypes.BoolValue(got.Payload.Spec.Paused)
	}

	// Set default for min_ready_seconds if not provided by user
	if state.MinReadySeconds.IsNull() || state.MinReadySeconds.IsUnknown() {
		state.MinReadySeconds = tftypes.Int64Value(0)
	}

	// Set autoscaling values from plan (they don't come back from API response)
	if !plan.MinReplicas.IsNull() && !plan.MinReplicas.IsUnknown() {
		state.MinReplicas = plan.MinReplicas
	}
	if !plan.MaxReplicas.IsNull() && !plan.MaxReplicas.IsUnknown() {
		state.MaxReplicas = plan.MaxReplicas
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// updateChanges represents the changes detected in an update operation.
type updateChanges struct {
	needReplicas    bool
	needVersion     bool
	needMinReplicas bool
	needMaxReplicas bool
	
	wantReplicas    int64
	wantVersion     string
	wantMinReplicas int64
	wantMaxReplicas int64
}

// hasChanges returns true if any changes are needed.
func (c *updateChanges) hasChanges() bool {
	return c.needReplicas || c.needVersion || c.needMinReplicas || c.needMaxReplicas
}

// detectUpdateChanges detects what changes need to be applied during an update.
func (r *resourceMachineDeployment) detectUpdateChanges(plan, state machineDeploymentState) *updateChanges {
	wantReplicas := plan.Replicas.ValueInt64()
	curReplicas := state.Replicas.ValueInt64()
	
	wantVersion := strings.TrimSpace(plan.K8sVersion.ValueString())
	curVersion := strings.TrimSpace(state.K8sVersion.ValueString())
	
	wantMinReplicas := plan.MinReplicas.ValueInt64()
	curMinReplicas := state.MinReplicas.ValueInt64()
	
	wantMaxReplicas := plan.MaxReplicas.ValueInt64()
	curMaxReplicas := state.MaxReplicas.ValueInt64()
	
	return &updateChanges{
		needReplicas:    wantReplicas != curReplicas,
		needVersion:     wantVersion != "" && wantVersion != curVersion,
		needMinReplicas: wantMinReplicas != curMinReplicas,
		needMaxReplicas: wantMaxReplicas != curMaxReplicas,
		
		wantReplicas:    wantReplicas,
		wantVersion:     wantVersion,
		wantMinReplicas: wantMinReplicas,
		wantMaxReplicas: wantMaxReplicas,
	}
}

// applyUpdateChanges applies the detected changes by building a patch and executing it.
func (r *resourceMachineDeployment) applyUpdateChanges(ctx context.Context, changes *updateChanges, clusterID, id string, resp *resource.UpdateResponse) error {
	// Build patch specification
	patchBody, err := r.buildUpdatePatch(changes, resp)
	if err != nil {
		return err
	}
	
	// Execute the patch
	if err := r.executePatch(patchBody, clusterID, id, resp); err != nil {
		return err
	}
	
	// Wait for update to complete
	return r.waitForUpdateCompletion(ctx, changes, clusterID, id, resp)
}

// buildUpdatePatch builds the patch body for the update operation.
func (r *resourceMachineDeployment) buildUpdatePatch(changes *updateChanges, resp *resource.UpdateResponse) (map[string]any, error) {
	spec := map[string]any{}
	
	if changes.needReplicas {
		replicas, err := kkp.SafeInt32(changes.wantReplicas)
		if err != nil {
			resp.Diagnostics.AddError("Invalid Replicas Value", err.Error())
			return nil, err
		}
		spec["replicas"] = replicas
	}
	
	if changes.needVersion {
		spec["template"] = map[string]any{
			"versions": map[string]any{
				"kubelet": changes.wantVersion,
			},
		}
	}
	
	if changes.needMinReplicas || changes.needMaxReplicas {
		autoscaling, err := r.buildAutoscalingPatch(changes, resp)
		if err != nil {
			return nil, err
		}
		if len(autoscaling) > 0 {
			autoscaling["enabled"] = true
			spec["autoscaling"] = autoscaling
		}
	}
	
	return map[string]any{"spec": spec}, nil
}

// buildAutoscalingPatch builds the autoscaling configuration for the patch.
func (r *resourceMachineDeployment) buildAutoscalingPatch(changes *updateChanges, resp *resource.UpdateResponse) (map[string]any, error) {
	autoscaling := map[string]any{}
	
	if changes.needMinReplicas && changes.wantMinReplicas > 0 {
		minReplicas, err := kkp.SafeInt32(changes.wantMinReplicas)
		if err != nil {
			resp.Diagnostics.AddError("Invalid MinReplicas Value", err.Error())
			return nil, err
		}
		autoscaling["minNodeCount"] = minReplicas
	}
	
	if changes.needMaxReplicas && changes.wantMaxReplicas > 0 {
		maxReplicas, err := kkp.SafeInt32(changes.wantMaxReplicas)
		if err != nil {
			resp.Diagnostics.AddError("Invalid MaxReplicas Value", err.Error())
			return nil, err
		}
		autoscaling["maxNodeCount"] = maxReplicas
	}
	
	return autoscaling, nil
}

// executePatch executes the patch operation against the API.
func (r *resourceMachineDeployment) executePatch(patchBody map[string]any, clusterID, id string, resp *resource.UpdateResponse) error {
	pcli := kapi.New(r.Client.Transport, nil)
	_, err := pcli.PatchMachineDeployment(
		kapi.NewPatchMachineDeploymentParams().
			WithProjectID(r.DefaultProjectID).
			WithClusterID(clusterID).
			WithMachineDeploymentID(id).
			WithPatch(patchBody),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Patch machine deployment failed", err.Error())
		return err
	}
	return nil
}

// waitForUpdateCompletion waits for the update operation to complete.
func (r *resourceMachineDeployment) waitForUpdateCompletion(ctx context.Context, changes *updateChanges, clusterID, id string, resp *resource.UpdateResponse) error {
	tflog.Info(ctx, "machine deployment patch sent", map[string]any{
		"cluster_id":            clusterID,
		"machine_deployment_id": id,
		"replicas":              changes.wantReplicas,
		"version":               changes.wantVersion,
	})
	
	checker := &kkp.MachineDeploymentHealthChecker{
		Client:              r.Client,
		ProjectID:           r.DefaultProjectID,
		ClusterID:           clusterID,
		MachineDeploymentID: id,
		ExpectedReplicas:    changes.wantReplicas, // Wait for the expected replica count
	}
	
	if err := checker.WaitForMachineDeploymentReady(ctx); err != nil {
		resp.Diagnostics.AddError("Machine deployment update timed out", err.Error())
		return err
	}
	
	return nil
}

// buildAndPersistUpdateState builds the final state after update and persists it.
func (r *resourceMachineDeployment) buildAndPersistUpdateState(ctx context.Context, plan, state machineDeploymentState, clusterID, id string, resp *resource.UpdateResponse) {
	// Read updated machine deployment to get current values for computed fields
	pcli := kapi.New(r.Client.Transport, nil)
	getParams := kapi.NewGetMachineDeploymentParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithMachineDeploymentID(id)

	got, err := pcli.GetMachineDeployment(getParams, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read machine deployment after update failed", err.Error())
		return
	}

	// Build final state with known values from API response
	finalState := plan
	finalState.ID = state.ID
	finalState.ClusterID = state.ClusterID

	// Set computed fields from API response to ensure they have known values
	if got.Payload != nil && got.Payload.Spec != nil {
		if got.Payload.Spec.Replicas != nil {
			finalState.Replicas = tftypes.Int64Value(int64(*got.Payload.Spec.Replicas))
		}
		if got.Payload.Spec.Template != nil && got.Payload.Spec.Template.Versions != nil {
			finalState.K8sVersion = tftypes.StringValue(got.Payload.Spec.Template.Versions.Kubelet)
		}
		finalState.Paused = tftypes.BoolValue(got.Payload.Spec.Paused)
	}

	// Set default for min_ready_seconds if not provided by user (similar to Create method)
	if finalState.MinReadySeconds.IsNull() || finalState.MinReadySeconds.IsUnknown() {
		finalState.MinReadySeconds = tftypes.Int64Value(0)
	}

	// Set autoscaling values from plan (they don't come back from API response)
	if !plan.MinReplicas.IsNull() && !plan.MinReplicas.IsUnknown() {
		finalState.MinReplicas = plan.MinReplicas
	}
	if !plan.MaxReplicas.IsNull() && !plan.MaxReplicas.IsUnknown() {
		finalState.MaxReplicas = plan.MaxReplicas
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &finalState)...)
}
