package addon_v2

import (
	"context"
	"encoding/json"
	"fmt"
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

	acli "github.com/kubermatic/go-kubermatic/client/addon"
	"github.com/kubermatic/go-kubermatic/models"
)

var (
	_ resource.Resource                = &resourceAddon{}
	_ resource.ResourceWithConfigure   = &resourceAddon{}
	_ resource.ResourceWithImportState = &resourceAddon{}
)

// New creates a new addon v2 resource.
func New() resource.Resource { return &resourceAddon{} }

func (r *resourceAddon) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_addon_v2"
}

func (r *resourceAddon) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Create a KKP addon for a cluster using V2 API.",
		Attributes:  r.buildSchemaAttributes(),
	}
}

// buildSchemaAttributes builds the attributes for the addon resource schema.
func (r *resourceAddon) buildSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id":                     r.buildIDAttribute(),
		"cluster_id":             r.buildClusterIDAttribute(),
		"name":                   r.buildNameAttribute(),
		"continuously_reconcile": r.buildContinuouslyReconcileAttribute(),
		"is_default":             r.buildIsDefaultAttribute(),
		"variables":              r.buildVariablesAttribute(),
		"wait_for_ready":         r.buildWaitForReadyAttribute(),
		"timeout_minutes":        r.buildTimeoutMinutesAttribute(),
		"status":                 r.buildStatusAttribute(),
		"last_checked":           r.buildLastCheckedAttribute(),
		"created_at":             r.buildCreatedAtAttribute(),
		"status_message":         r.buildStatusMessageAttribute(),
	}
}

// buildIDAttribute builds the id attribute.
func (r *resourceAddon) buildIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Addon ID.",
	}
}

// buildClusterIDAttribute builds the cluster_id attribute.
func (r *resourceAddon) buildClusterIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Cluster ID to install the addon to.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildNameAttribute builds the name attribute.
func (r *resourceAddon) buildNameAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Addon name (e.g., 'prometheus', 'grafana', 'logging').",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
}

// buildContinuouslyReconcileAttribute builds the continuously_reconcile attribute.
func (r *resourceAddon) buildContinuouslyReconcileAttribute() rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Indicates that the addon cannot be deleted or modified outside of the UI after installation.",
	}
}

// buildIsDefaultAttribute builds the is_default attribute.
func (r *resourceAddon) buildIsDefaultAttribute() rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Indicates whether the addon is default.",
	}
}

// buildVariablesAttribute builds the variables attribute.
func (r *resourceAddon) buildVariablesAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:    true,
		Description: "Free form JSON data to use for parsing the manifest templates.",
	}
}

// buildWaitForReadyAttribute builds the wait_for_ready attribute.
func (r *resourceAddon) buildWaitForReadyAttribute() rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Wait for addon to be ready during creation. Defaults to true.",
	}
}

// buildTimeoutMinutesAttribute builds the timeout_minutes attribute.
func (r *resourceAddon) buildTimeoutMinutesAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Timeout in minutes for waiting for addon to be ready. Defaults to 2 minutes.",
	}
}

// buildStatusAttribute builds the status attribute.
func (r *resourceAddon) buildStatusAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Installation status: 'installing', 'ready', 'failed', 'deleting'.",
	}
}

// buildLastCheckedAttribute builds the last_checked attribute.
func (r *resourceAddon) buildLastCheckedAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Last time status was checked (RFC3339 timestamp).",
	}
}

// buildCreatedAtAttribute builds the created_at attribute.
func (r *resourceAddon) buildCreatedAtAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "When the addon was created (RFC3339 timestamp).",
	}
}

// buildStatusMessageAttribute builds the status_message attribute.
func (r *resourceAddon) buildStatusMessageAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Additional status details or error messages.",
	}
}

func (r *resourceAddon) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

func (r *resourceAddon) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ValidateResourceBase(resp) {
		return
	}

	plan, ok := kkp.ExtractPlan[addonState](ctx, req, resp)
	if !ok {
		return
	}

	cp := Plan{
		Name:                  plan.Name.ValueString(),
		ClusterID:             plan.ClusterID.ValueString(),
		ContinuouslyReconcile: plan.ContinuouslyReconcile.ValueBool(),
		IsDefault:             plan.IsDefault.ValueBool(),
	}

	// Parse variables JSON if provided
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		variablesStr := strings.TrimSpace(plan.Variables.ValueString())
		if variablesStr != "" {
			variables, err := kkp.JSONToVariables(variablesStr)
			if err != nil {
				resp.Diagnostics.AddError("Invalid variables JSON", err.Error())
				return
			}
			cp.Variables = variables
		}
	}

	addon, err := cp.ToAddon()
	if err != nil {
		resp.Diagnostics.AddError("Addon spec invalid", err.Error())
		return
	}

	aclient := acli.New(r.Client.Transport, nil)
	params := acli.NewCreateAddonV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(cp.ClusterID).
		WithBody(addon)

	out, err := aclient.CreateAddonV2(params, nil)
	if err != nil {
		if e, ok := err.(*acli.CreateAddonV2Default); ok && e.Payload != nil {
			b, _ := json.MarshalIndent(e.Payload, "", "  ")
			resp.Diagnostics.AddError("Create addon failed", string(b))
		} else if _, ok := err.(*acli.CreateAddonV2Unauthorized); ok {
			resp.Diagnostics.AddError(
				"Create addon failed - Unauthorized",
				fmt.Sprintf("Failed to create addon '%s'. This could be due to: 1) Invalid API token, 2) Insufficient permissions, 3) Addon name '%s' does not exist. Check available addons with 'data.kkp_addons_v2' and verify your API token has addon management permissions.", cp.Name, cp.Name),
			)
		} else if _, ok := err.(*acli.CreateAddonV2Forbidden); ok {
			resp.Diagnostics.AddError(
				"Create addon failed - Forbidden",
				fmt.Sprintf("Access denied when creating addon '%s'. Your API token may not have sufficient permissions for addon management on this cluster.", cp.Name),
			)
		} else {
			resp.Diagnostics.AddError("Create addon failed", err.Error())
		}
		return
	}

	addonID := out.Payload.ID

	tflog.Info(ctx, "addon created successfully", map[string]any{
		"cluster_id": cp.ClusterID,
		"addon_id":   addonID,
		"name":       cp.Name,
	})

	// Build initial state from response
	state := plan
	state.ID = tftypes.StringValue(addonID)
	state.Name = tftypes.StringValue(out.Payload.Name)

	// Set defaults for optional fields
	waitForReady := true
	if kkp.IsAttributeSet(plan.WaitForReady) {
		waitForReady = plan.WaitForReady.ValueBool()
	}
	state.WaitForReady = tftypes.BoolValue(waitForReady)

	timeoutMinutes := int64(2)
	if kkp.IsAttributeSet(plan.TimeoutMinutes) {
		timeoutMinutes = plan.TimeoutMinutes.ValueInt64()
	}
	state.TimeoutMinutes = tftypes.Int64Value(timeoutMinutes)

	// Set creation timestamp
	if !out.Payload.CreationTimestamp.IsZero() {
		state.CreatedAt = tftypes.StringValue(out.Payload.CreationTimestamp.String())
	} else {
		state.CreatedAt = tftypes.StringValue(time.Now().Format(time.RFC3339))
	}

	if out.Payload.Spec != nil {
		state.ContinuouslyReconcile = tftypes.BoolValue(out.Payload.Spec.ContinuouslyReconcile)
		state.IsDefault = tftypes.BoolValue(out.Payload.Spec.IsDefault)

		// Convert variables back to JSON string for storage
		if out.Payload.Spec.Variables != nil {
			variablesJSON, err := kkp.VariablesToJSON(out.Payload.Spec.Variables)
			if err == nil {
				state.Variables = tftypes.StringValue(variablesJSON)
			}
		}
	}

	// Wait for addon to be ready based on user configuration
	if waitForReady {
		timeout := time.Duration(timeoutMinutes) * time.Minute
		tflog.Info(ctx, "Waiting for addon installation to complete", map[string]any{
			"cluster_id":      cp.ClusterID,
			"addon_id":        addonID,
			"timeout_minutes": timeoutMinutes,
			"wait_enabled":    waitForReady,
		})

		status, message := r.waitForAddonReady(ctx, cp.ClusterID, addonID, timeout)
		r.updateStatusFields(state, status, message)
	} else {
		tflog.Info(ctx, "Skipping addon readiness check (wait_for_ready=false)", map[string]any{
			"cluster_id": cp.ClusterID,
			"addon_id":   addonID,
		})

		// Just do a quick status check without waiting
		status, message := r.checkAddonStatus(ctx, cp.ClusterID, addonID)
		r.updateStatusFields(state, status, message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceAddon) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ValidateResourceBaseRead(resp) {
		return
	}

	state, ok := kkp.ExtractState[addonState](ctx, req, resp)
	if !ok {
		return
	}

	id, clusterID := r.extractIdentifiers(state, resp)
	if id == "" || clusterID == "" {
		return
	}

	addon, err := r.fetchAddon(clusterID, id)
	if err != nil {
		r.handleFetchAddonError(ctx, err, resp)
		return
	}

	r.updateStateFromAddon(state, addon)
	r.refreshAddonStatus(ctx, state, clusterID, id)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// extractIdentifiers extracts and validates addon and cluster IDs from state.
func (r *resourceAddon) extractIdentifiers(state *addonState, resp *resource.ReadResponse) (id, clusterID string) {
	id = strings.TrimSpace(state.ID.ValueString())
	clusterID = strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing addon or cluster ID.")
		return "", ""
	}
	return id, clusterID
}

// fetchAddon retrieves addon details from the API.
func (r *resourceAddon) fetchAddon(clusterID, addonID string) (*acli.GetAddonV2OK, error) {
	aclient := acli.New(r.Client.Transport, nil)
	get := acli.NewGetAddonV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithAddonID(addonID)

	return aclient.GetAddonV2(get, nil)
}

// handleFetchAddonError handles errors from fetching addon.
func (r *resourceAddon) handleFetchAddonError(ctx context.Context, err error, resp *resource.ReadResponse) {
	if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.AddError("Read addon failed", err.Error())
}

// updateStateFromAddon updates state with addon details from API response.
func (r *resourceAddon) updateStateFromAddon(state *addonState, addon *acli.GetAddonV2OK) {
	state.ID = tftypes.StringValue(addon.Payload.ID)
	state.Name = tftypes.StringValue(addon.Payload.Name)

	// Update creation timestamp if available
	if !addon.Payload.CreationTimestamp.IsZero() {
		state.CreatedAt = tftypes.StringValue(addon.Payload.CreationTimestamp.String())
	}

	if addon.Payload.Spec != nil {
		state.ContinuouslyReconcile = tftypes.BoolValue(addon.Payload.Spec.ContinuouslyReconcile)
		state.IsDefault = tftypes.BoolValue(addon.Payload.Spec.IsDefault)

		// Convert variables to JSON string
		if addon.Payload.Spec.Variables != nil {
			variablesJSON, err := kkp.VariablesToJSON(addon.Payload.Spec.Variables)
			if err == nil {
				state.Variables = tftypes.StringValue(variablesJSON)
			}
		}
	}
}

// refreshAddonStatus checks and updates current addon status.
func (r *resourceAddon) refreshAddonStatus(ctx context.Context, state *addonState, clusterID, addonID string) {
	status, message := r.checkAddonStatus(ctx, clusterID, addonID)
	r.updateStatusFields(state, status, message)

	tflog.Debug(ctx, "addon status refreshed", map[string]any{
		"cluster_id": clusterID,
		"addon_id":   addonID,
		"status":     status,
		"message":    message,
	})
}

func (r *resourceAddon) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.ValidateResourceBaseUpdate(resp) {
		return
	}

	plan, ok := kkp.ExtractStateForUpdate[addonState](ctx, req, resp)
	if !ok {
		return
	}

	var state addonState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	clusterID := strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing addon or cluster ID.")
		return
	}

	// Build patch body for changes
	patch := map[string]any{}

	// Check what changed and build minimal patch
	needsUpdate := false
	if plan.ContinuouslyReconcile.ValueBool() != state.ContinuouslyReconcile.ValueBool() {
		patch["continuouslyReconcile"] = plan.ContinuouslyReconcile.ValueBool()
		needsUpdate = true
	}

	if plan.IsDefault.ValueBool() != state.IsDefault.ValueBool() {
		patch["isDefault"] = plan.IsDefault.ValueBool()
		needsUpdate = true
	}

	// Check variables changes
	planVars := strings.TrimSpace(plan.Variables.ValueString())
	stateVars := strings.TrimSpace(state.Variables.ValueString())
	if planVars != stateVars {
		if planVars != "" {
			variables, err := kkp.JSONToVariables(planVars)
			if err != nil {
				resp.Diagnostics.AddError("Invalid variables JSON", err.Error())
				return
			}
			patch["variables"] = variables
		} else {
			patch["variables"] = nil
		}
		needsUpdate = true
	}

	// Nothing to change -> just keep state
	if !needsUpdate {
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	// Build the addon object for patch
	patchAddon := &models.Addon{
		Name: plan.Name.ValueString(),
		Spec: &models.AddonSpec{
			ContinuouslyReconcile: plan.ContinuouslyReconcile.ValueBool(),
			IsDefault:             plan.IsDefault.ValueBool(),
		},
	}

	// Add variables if provided
	if !plan.Variables.IsNull() && !plan.Variables.IsUnknown() {
		variablesStr := strings.TrimSpace(plan.Variables.ValueString())
		if variablesStr != "" {
			variables, err := kkp.JSONToVariables(variablesStr)
			if err != nil {
				resp.Diagnostics.AddError("Invalid variables JSON", err.Error())
				return
			}
			if m, ok := variables.(map[string]interface{}); ok {
				patchAddon.Spec.Variables = m
			} else {
				resp.Diagnostics.AddError("Invalid variables JSON", "expected object at top level")
				return
			}
		}
	}

	aclient := acli.New(r.Client.Transport, nil)
	_, err := aclient.PatchAddonV2(
		acli.NewPatchAddonV2Params().
			WithProjectID(r.DefaultProjectID).
			WithClusterID(clusterID).
			WithAddonID(id).
			WithBody(patchAddon),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Patch addon failed", err.Error())
		return
	}

	tflog.Info(ctx, "addon patch sent", map[string]any{
		"cluster_id": clusterID,
		"addon_id":   id,
		"name":       plan.Name.ValueString(),
	})

	// Read updated addon to get current values
	get := acli.NewGetAddonV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithAddonID(id)

	got, err := aclient.GetAddonV2(get, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read addon after update failed", err.Error())
		return
	}

	// Build final state with known values from API response
	finalState := plan
	finalState.ID = state.ID
	finalState.ClusterID = state.ClusterID
	finalState.Name = tftypes.StringValue(got.Payload.Name)

	// Preserve creation timestamp from previous state
	finalState.CreatedAt = state.CreatedAt

	if got.Payload.Spec != nil {
		finalState.ContinuouslyReconcile = tftypes.BoolValue(got.Payload.Spec.ContinuouslyReconcile)
		finalState.IsDefault = tftypes.BoolValue(got.Payload.Spec.IsDefault)

		// Convert variables to JSON string
		if got.Payload.Spec.Variables != nil {
			variablesJSON, err := kkp.VariablesToJSON(got.Payload.Spec.Variables)
			if err == nil {
				finalState.Variables = tftypes.StringValue(variablesJSON)
			}
		} else {
			finalState.Variables = tftypes.StringValue("{}")
		}
	}

	// Check and update current status after update
	status, message := r.checkAddonStatus(ctx, clusterID, id)
	r.updateStatusFields(finalState, status, message)

	resp.Diagnostics.Append(resp.State.Set(ctx, &finalState)...)
}

func (r *resourceAddon) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ValidateResourceBaseDelete(resp) {
		return
	}

	state, ok := kkp.ExtractStateForDelete[addonState](ctx, req, resp)
	if !ok {
		return
	}

	id := strings.TrimSpace(state.ID.ValueString())
	clusterID := strings.TrimSpace(state.ClusterID.ValueString())
	if id == "" || clusterID == "" {
		return
	}

	aclient := acli.New(r.Client.Transport, nil)
	del := acli.NewDeleteAddonV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithAddonID(id)

	if _, err := aclient.DeleteAddonV2(del, nil); err != nil {
		resp.Diagnostics.AddWarning("Delete addon warning", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceAddon) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: cluster_id:addon_id
	parts := strings.Split(req.ID, ":")
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Unexpected import ID", "Expected 'cluster_id:addon_id'")
		return
	}

	clusterID := strings.TrimSpace(parts[0])
	addonID := strings.TrimSpace(parts[1])
	if clusterID == "" || addonID == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Both cluster_id and addon_id must be non-empty")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), addonID)...)
}

// Helper function to check addon status
func (r *resourceAddon) checkAddonStatus(_ context.Context, clusterID, addonID string) (status, message string) {
	aclient := acli.New(r.Client.Transport, nil)
	get := acli.NewGetAddonV2Params().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithAddonID(addonID)

	got, err := aclient.GetAddonV2(get, nil)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
			return kkp.StatusFailed, "Addon not found - installation may have failed"
		}
		return kkp.StatusFailed, fmt.Sprintf("Error checking status: %s", err.Error())
	}

	// Check if addon has deletion timestamp (being deleted)
	if !got.Payload.DeletionTimestamp.IsZero() {
		return "deleting", "Addon is being deleted"
	}

	// If we get here, addon exists and is not being deleted
	return kkp.StatusReady, "Addon is installed and ready"
}

// Helper function to wait for addon installation with polling
func (r *resourceAddon) waitForAddonReady(ctx context.Context, clusterID, addonID string, maxWaitTime time.Duration) (status, message string) {
	deadline := time.Now().Add(maxWaitTime)
	pollInterval := 10 * time.Second
	maxAttempts := int(maxWaitTime / pollInterval)

	r.logWaitStart(ctx, clusterID, addonID, maxWaitTime, pollInterval, maxAttempts)

	for attempt := 1; time.Now().Before(deadline); attempt++ {
		status, msg := r.checkAddonStatus(ctx, clusterID, addonID)

		r.logStatusCheck(ctx, clusterID, addonID, attempt, maxAttempts, status, msg, deadline.Add(-maxWaitTime))

		finalStatus, finalMessage, isDone := r.evaluateAddonStatus(ctx, clusterID, addonID, status, msg, attempt, deadline.Add(-maxWaitTime))
		if isDone {
			return finalStatus, finalMessage
		}

		if r.shouldStopPolling(ctx, clusterID, addonID, attempt, pollInterval) {
			return "failed", "Installation canceled"
		}
	}

	return r.handleTimeout(ctx, clusterID, addonID, maxWaitTime, maxAttempts)
}

// logWaitStart logs the start of addon installation monitoring.
func (r *resourceAddon) logWaitStart(ctx context.Context, clusterID, addonID string, maxWaitTime, pollInterval time.Duration, maxAttempts int) {
	tflog.Info(ctx, "Starting addon installation monitoring", map[string]any{
		"cluster_id":    clusterID,
		"addon_id":      addonID,
		"timeout":       maxWaitTime.String(),
		"poll_interval": pollInterval.String(),
		"max_attempts":  maxAttempts,
	})
}

// logStatusCheck logs status check progress.
func (r *resourceAddon) logStatusCheck(ctx context.Context, clusterID, addonID string, attempt, maxAttempts int, status, msg string, startTime time.Time) {
	// Log progress every 30 seconds (every 3rd attempt)
	if attempt == 1 || attempt%3 == 0 || status == kkp.StatusReady || status == kkp.StatusFailed {
		tflog.Info(ctx, "Addon installation status check", map[string]any{
			"cluster_id": clusterID,
			"addon_id":   addonID,
			"attempt":    fmt.Sprintf("%d/%d", attempt, maxAttempts),
			"status":     status,
			"message":    msg,
			"elapsed":    time.Since(startTime).String(),
		})
	}
}

// evaluateAddonStatus evaluates the addon status and returns final status if done.
func (r *resourceAddon) evaluateAddonStatus(ctx context.Context, clusterID, addonID, status, msg string, attempt int, startTime time.Time) (finalStatus, finalMessage string, isDone bool) {
	elapsed := time.Since(startTime).String()

	switch status {
	case "ready":
		tflog.Info(ctx, "Addon installation completed successfully", map[string]any{
			"cluster_id": clusterID,
			"addon_id":   addonID,
			"attempts":   attempt,
			"elapsed":    elapsed,
		})
		return status, msg, true
	case kkp.StatusFailed:
		tflog.Error(ctx, "Addon installation failed", map[string]any{
			"cluster_id": clusterID,
			"addon_id":   addonID,
			"attempts":   attempt,
			"elapsed":    elapsed,
			"error":      msg,
		})
		return status, msg, true
	case "deleting":
		tflog.Error(ctx, "Addon was marked for deletion during installation", map[string]any{
			"cluster_id": clusterID,
			"addon_id":   addonID,
			"attempts":   attempt,
		})
		return "failed", "Addon was marked for deletion during installation", true
	}

	return "", "", false
}

// shouldStopPolling checks if polling should be stopped due to context cancellation.
func (r *resourceAddon) shouldStopPolling(ctx context.Context, clusterID, addonID string, attempt int, pollInterval time.Duration) bool {
	select {
	case <-ctx.Done():
		tflog.Warn(ctx, "Addon installation monitoring canceled", map[string]any{
			"cluster_id": clusterID,
			"addon_id":   addonID,
			"attempts":   attempt,
		})
		return true
	case <-time.After(pollInterval):
		return false
	}
}

// handleTimeout handles the timeout scenario when max wait time is reached.
func (r *resourceAddon) handleTimeout(ctx context.Context, clusterID, addonID string, maxWaitTime time.Duration, maxAttempts int) (status, message string) {
	tflog.Warn(ctx, "Addon installation monitoring timed out", map[string]any{
		"cluster_id":   clusterID,
		"addon_id":     addonID,
		"timeout":      maxWaitTime.String(),
		"max_attempts": maxAttempts,
	})
	return "installing", fmt.Sprintf("Installation timeout after %v - addon may still be installing", maxWaitTime)
}

// Helper function to update status fields in state
func (r *resourceAddon) updateStatusFields(state *addonState, status, message string) {
	now := time.Now().Format(time.RFC3339)
	state.Status = tftypes.StringValue(status)
	state.StatusMessage = tftypes.StringValue(message)
	state.LastChecked = tftypes.StringValue(now)
}
