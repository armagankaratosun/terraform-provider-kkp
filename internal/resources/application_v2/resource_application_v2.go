package application_v2

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

	acli "github.com/kubermatic/go-kubermatic/client/applications"
	"github.com/kubermatic/go-kubermatic/models"
)

var (
	_ resource.Resource                = &resourceApplication{}
	_ resource.ResourceWithConfigure   = &resourceApplication{}
	_ resource.ResourceWithImportState = &resourceApplication{}
)

// New creates a new application v2 resource.
func New() resource.Resource { return &resourceApplication{} }

func (r *resourceApplication) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application_v2"
}

func (r *resourceApplication) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = rschema.Schema{
		Description: "Create a KKP application installation for a cluster using V2 API.",
		Attributes:  r.buildSchemaAttributes(),
	}
}

// buildSchemaAttributes builds the attributes for the application resource schema.
func (r *resourceApplication) buildSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"id":                  r.buildIDAttribute(),
		"cluster_id":          r.buildClusterIDAttribute(),
		"name":                r.buildNameAttribute(),
		"namespace":           r.buildNamespaceAttribute(),
		"application_name":    r.buildApplicationNameAttribute(),
		"application_version": r.buildApplicationVersionAttribute(),
		"values":              r.buildValuesAttribute(),
		"wait_for_ready":      r.buildWaitForReadyAttribute(),
		"timeout_minutes":     r.buildTimeoutMinutesAttribute(),
		"status":              r.buildStatusAttribute(),
		"last_checked":        r.buildLastCheckedAttribute(),
		"created_at":          r.buildCreatedAtAttribute(),
		"status_message":      r.buildStatusMessageAttribute(),
	}
}

// buildIDAttribute builds the id attribute.
func (r *resourceApplication) buildIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Application installation ID.",
	}
}

// buildClusterIDAttribute builds the cluster_id attribute.
func (r *resourceApplication) buildClusterIDAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Cluster ID to install the application to.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildNameAttribute builds the name attribute.
func (r *resourceApplication) buildNameAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Application installation name.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
}

// buildNamespaceAttribute builds the namespace attribute.
func (r *resourceApplication) buildNamespaceAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Kubernetes namespace for the application installation. Defaults to 'default'.",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
	}
}

// buildApplicationNameAttribute builds the application_name attribute.
func (r *resourceApplication) buildApplicationNameAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Name of the application to install (from ApplicationDefinitions).",
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.RequiresReplace(),
		},
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
}

// buildApplicationVersionAttribute builds the application_version attribute.
func (r *resourceApplication) buildApplicationVersionAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Required:    true,
		Description: "Version of the application to install.",
		Validators: []validator.String{
			stringvalidator.LengthAtLeast(1),
		},
	}
}

// buildValuesAttribute builds the values attribute.
func (r *resourceApplication) buildValuesAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Application configuration values as JSON string (e.g., Helm values).",
	}
}

// buildWaitForReadyAttribute builds the wait_for_ready attribute.
func (r *resourceApplication) buildWaitForReadyAttribute() rschema.BoolAttribute {
	return rschema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Wait for application to be ready during creation. Defaults to true.",
	}
}

// buildTimeoutMinutesAttribute builds the timeout_minutes attribute.
func (r *resourceApplication) buildTimeoutMinutesAttribute() rschema.Int64Attribute {
	return rschema.Int64Attribute{
		Optional:    true,
		Computed:    true,
		Description: "Timeout in minutes for waiting for application to be ready. Defaults to 5 minutes.",
	}
}

// buildStatusAttribute builds the status attribute.
func (r *resourceApplication) buildStatusAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Installation status: 'installing', 'ready', 'failed', 'deleting'.",
	}
}

// buildLastCheckedAttribute builds the last_checked attribute.
func (r *resourceApplication) buildLastCheckedAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Last time status was checked (RFC3339 timestamp).",
	}
}

// buildCreatedAtAttribute builds the created_at attribute.
func (r *resourceApplication) buildCreatedAtAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "When the application was created (RFC3339 timestamp).",
	}
}

// buildStatusMessageAttribute builds the status_message attribute.
func (r *resourceApplication) buildStatusMessageAttribute() rschema.StringAttribute {
	return rschema.StringAttribute{
		Computed:    true,
		Description: "Additional status details or error messages.",
	}
}

func (r *resourceApplication) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.ConfigureResource(req, resp)
}

func (r *resourceApplication) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	if !r.ValidateResourceBase(resp) {
		return
	}

	plan, ok := kkp.ExtractPlan[applicationState](ctx, req, resp)
	if !ok {
		return
	}

	cp := Plan{
		Name:               plan.Name.ValueString(),
		ClusterID:          plan.ClusterID.ValueString(),
		Namespace:          plan.Namespace.ValueString(),
		ApplicationName:    plan.ApplicationName.ValueString(),
		ApplicationVersion: plan.ApplicationVersion.ValueString(),
	}

	// Parse values JSON if provided
	if !plan.Values.IsNull() && !plan.Values.IsUnknown() {
		valuesStr := strings.TrimSpace(plan.Values.ValueString())
		if valuesStr != "" {
			values, err := kkp.JSONToVariables(valuesStr)
			if err != nil {
				resp.Diagnostics.AddError("Invalid values JSON", err.Error())
				return
			}
			cp.Values = values
		}
	}

	appInstallation, err := cp.ToApplicationInstallation()
	if err != nil {
		resp.Diagnostics.AddError("Application spec invalid", err.Error())
		return
	}

	aclient := acli.New(r.Client.Transport, nil)
	params := acli.NewCreateApplicationInstallationParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(cp.ClusterID).
		WithBody(&models.ApplicationInstallationBody{
			Name:      appInstallation.Name,
			Namespace: appInstallation.Namespace,
			Spec:      appInstallation.Spec,
		})

	out, err := aclient.CreateApplicationInstallation(params, nil)
	if err != nil {
		if e, ok := err.(*acli.CreateApplicationInstallationDefault); ok && e.Payload != nil {
			b, _ := json.MarshalIndent(e.Payload, "", "  ")
			resp.Diagnostics.AddError("Create application failed", string(b))
		} else {
			resp.Diagnostics.AddError("Create application failed", err.Error())
		}
		return
	}

	appID := strings.TrimSpace(out.Payload.ID)
	if appID == "" {
		appID = r.fallbackApplicationID(cp.ClusterID, cp.Namespace, cp.Name)
	}

	tflog.Info(ctx, "application created successfully", map[string]any{
		"cluster_id":  cp.ClusterID,
		"app_id":      appID,
		"name":        cp.Name,
		"app_name":    cp.ApplicationName,
		"app_version": cp.ApplicationVersion,
	})

	// Build initial state from response
	state := plan
	state.ID = tftypes.StringValue(appID)
	state.Name = tftypes.StringValue(out.Payload.Name)
	state.Namespace = tftypes.StringValue(out.Payload.Namespace)

	// Set defaults for optional fields
	waitForReady := true
	if kkp.IsAttributeSet(plan.WaitForReady) {
		waitForReady = plan.WaitForReady.ValueBool()
	}
	state.WaitForReady = tftypes.BoolValue(waitForReady)

	timeoutMinutes := int64(5)
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
		// Set application reference info from spec
		if out.Payload.Spec.ApplicationRef != nil {
			state.ApplicationName = tftypes.StringValue(out.Payload.Spec.ApplicationRef.Name)
			state.ApplicationVersion = tftypes.StringValue(out.Payload.Spec.ApplicationRef.Version)
		}

		// Convert values back to JSON string for storage
		if out.Payload.Spec.Values != nil {
			if valuesStr := fmt.Sprintf("%v", out.Payload.Spec.Values); valuesStr != "" && valuesStr != kkp.NullValue {
				state.Values = tftypes.StringValue(valuesStr)
			} else if !plan.Values.IsNull() {
				// Preserve the planned values if API returned empty/null
				state.Values = plan.Values
			}
		} else if !plan.Values.IsNull() {
			// Preserve the planned values if API didn't return values
			state.Values = plan.Values
		}
	}

	// Wait for application to be ready based on user configuration
	if waitForReady {
		timeout := time.Duration(timeoutMinutes) * time.Minute
		tflog.Info(ctx, "Waiting for application installation to complete", map[string]any{
			"cluster_id":      cp.ClusterID,
			"app_id":          appID,
			"timeout_minutes": timeoutMinutes,
			"wait_enabled":    waitForReady,
		})

		status, message := r.waitForApplicationReady(ctx, cp.ClusterID, cp.Namespace, cp.Name, timeout)
		r.updateStatusFields(state, status, message)
	} else {
		tflog.Info(ctx, "Skipping application readiness check (wait_for_ready=false)", map[string]any{
			"cluster_id": cp.ClusterID,
			"app_id":     appID,
		})

		// Just do a quick status check without waiting
		status, message := r.checkApplicationStatus(ctx, cp.ClusterID, cp.Namespace, cp.Name)
		r.updateStatusFields(state, status, message)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourceApplication) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	if !r.ValidateResourceBaseRead(resp) {
		return
	}

	state, ok := kkp.ExtractState[applicationState](ctx, req, resp)
	if !ok {
		return
	}

	id, clusterID, namespace, name := r.extractApplicationIdentifiers(state, resp)
	if id == "" || clusterID == "" || namespace == "" || name == "" {
		return
	}

	application, err := r.fetchApplication(clusterID, namespace, name)
	if err != nil {
		r.handleFetchApplicationError(ctx, err, resp)
		return
	}

	r.updateStateFromApplication(state, application, clusterID, namespace)
	r.refreshApplicationStatus(ctx, state, clusterID, namespace, name)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

// extractApplicationIdentifiers extracts and validates application identifiers from state.
func (r *resourceApplication) extractApplicationIdentifiers(state *applicationState, resp *resource.ReadResponse) (id, clusterID, namespace, name string) {
	clusterID = kkp.TrimmedStringValue(state.ClusterID)
	namespace = kkp.TrimmedStringValue(state.Namespace)
	name = kkp.TrimmedStringValue(state.Name)
	if clusterID == "" || namespace == "" || name == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing required identifiers (cluster_id, namespace, name).")
		return "", "", "", ""
	}
	id = kkp.TrimmedStringValue(state.ID)
	if id == "" {
		id = r.fallbackApplicationID(clusterID, namespace, name)
		state.ID = tftypes.StringValue(id)
	}
	return id, clusterID, namespace, name
}

// fetchApplication retrieves application installation details from the API.
func (r *resourceApplication) fetchApplication(clusterID, namespace, name string) (*acli.GetApplicationInstallationOK, error) {
	aclient := acli.New(r.Client.Transport, nil)
	get := acli.NewGetApplicationInstallationParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithNamespace(namespace).
		WithApplicationInstallationName(name)

	return aclient.GetApplicationInstallation(get, nil)
}

// handleFetchApplicationError handles errors from fetching application.
func (r *resourceApplication) handleFetchApplicationError(ctx context.Context, err error, resp *resource.ReadResponse) {
	if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.AddError("Read application failed", err.Error())
}

// updateStateFromApplication updates state with application details from API response.
func (r *resourceApplication) updateStateFromApplication(state *applicationState, application *acli.GetApplicationInstallationOK, clusterID, namespace string) {
	apiID := strings.TrimSpace(application.Payload.ID)
	if apiID == "" {
		apiID = r.fallbackApplicationID(clusterID, namespace, application.Payload.Name)
	}
	state.ID = tftypes.StringValue(apiID)
	state.Name = tftypes.StringValue(application.Payload.Name)
	state.Namespace = tftypes.StringValue(application.Payload.Namespace)

	// Update creation timestamp if available
	if !application.Payload.CreationTimestamp.IsZero() {
		state.CreatedAt = tftypes.StringValue(application.Payload.CreationTimestamp.String())
	}

	if application.Payload.Spec != nil {
		// Update application reference info
		if application.Payload.Spec.ApplicationRef != nil {
			state.ApplicationName = tftypes.StringValue(application.Payload.Spec.ApplicationRef.Name)
			state.ApplicationVersion = tftypes.StringValue(application.Payload.Spec.ApplicationRef.Version)
		}

		// Convert values to JSON string, preserving null state if values are empty
		if application.Payload.Spec.Values != nil {
			if valuesStr := fmt.Sprintf("%v", application.Payload.Spec.Values); valuesStr != "" && valuesStr != "null" && valuesStr != "map[]" {
				state.Values = tftypes.StringValue(valuesStr)
			}
			// Don't modify state.Values if API returned empty/null - preserve existing state
		}
	}
}

// refreshApplicationStatus checks and updates current application status.
func (r *resourceApplication) refreshApplicationStatus(ctx context.Context, state *applicationState, clusterID, namespace, name string) {
	status, message := r.checkApplicationStatus(ctx, clusterID, namespace, name)
	r.updateStatusFields(state, status, message)

	tflog.Debug(ctx, "application status refreshed", map[string]any{
		"cluster_id": clusterID,
		"app_id":     state.ID.ValueString(),
		"namespace":  namespace,
		"name":       name,
		"status":     status,
		"message":    message,
	})
}

func (r *resourceApplication) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	if !r.ValidateResourceBaseUpdate(resp) {
		return
	}

	plan, ok := kkp.ExtractStateForUpdate[applicationState](ctx, req, resp)
	if !ok {
		return
	}

	var state applicationState
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := kkp.TrimmedStringValue(state.ID)
	clusterID := kkp.TrimmedStringValue(state.ClusterID)
	namespace := kkp.TrimmedStringValue(state.Namespace)
	name := kkp.TrimmedStringValue(state.Name)
	if clusterID == "" || namespace == "" || name == "" {
		resp.Diagnostics.AddError("Missing identifiers", "State missing required identifiers.")
		return
	}
	if id == "" {
		id = r.fallbackApplicationID(clusterID, namespace, name)
	}

	// Build update body from plan
	updateBody := &models.ApplicationInstallationBody{
		Name:      plan.Name.ValueString(),
		Namespace: namespace, // Cannot change namespace
	}

	// Create application reference from plan
	appRef := &models.ApplicationRef{
		Name:    plan.ApplicationName.ValueString(),
		Version: plan.ApplicationVersion.ValueString(),
	}

	// Create namespace spec
	namespaceSpec := &models.NamespaceSpec{
		Name: namespace,
	}

	// Create spec
	spec := &models.ApplicationInstallationSpec{
		ApplicationRef: appRef,
		Namespace:      namespaceSpec,
	}

	// Add values if provided
	if !plan.Values.IsNull() && !plan.Values.IsUnknown() {
		valuesStr := strings.TrimSpace(plan.Values.ValueString())
		if valuesStr != "" {
			spec.Values = models.RawExtension(valuesStr)
		}
	}

	updateBody.Spec = spec

	aclient := acli.New(r.Client.Transport, nil)
	_, err := aclient.UpdateApplicationInstallation(
		acli.NewUpdateApplicationInstallationParams().
			WithProjectID(r.DefaultProjectID).
			WithClusterID(clusterID).
			WithNamespace(namespace).
			WithApplicationInstallationName(name).
			WithBody(updateBody),
		nil,
	)
	if err != nil {
		resp.Diagnostics.AddError("Update application failed", err.Error())
		return
	}

	tflog.Info(ctx, "application updated successfully", map[string]any{
		"cluster_id": clusterID,
		"app_id":     id,
		"namespace":  namespace,
		"name":       name,
	})

	// Read updated application to get current values
	get := acli.NewGetApplicationInstallationParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithNamespace(namespace).
		WithApplicationInstallationName(name)

	got, err := aclient.GetApplicationInstallation(get, nil)
	if err != nil {
		resp.Diagnostics.AddError("Read application after update failed", err.Error())
		return
	}

	// Build final state with known values from API response
	finalState := plan
	currentID := kkp.TrimmedStringValue(state.ID)
	if currentID == "" {
		currentID = r.fallbackApplicationID(clusterID, namespace, got.Payload.Name)
	}
	finalState.ID = tftypes.StringValue(currentID)
	finalState.ClusterID = state.ClusterID
	finalState.Name = tftypes.StringValue(got.Payload.Name)
	finalState.Namespace = tftypes.StringValue(got.Payload.Namespace)

	// Preserve creation timestamp from previous state
	finalState.CreatedAt = state.CreatedAt
	// Preserve control settings from previous state
	finalState.WaitForReady = state.WaitForReady
	finalState.TimeoutMinutes = state.TimeoutMinutes

	if got.Payload.Spec != nil {
		if got.Payload.Spec.ApplicationRef != nil {
			finalState.ApplicationName = tftypes.StringValue(got.Payload.Spec.ApplicationRef.Name)
			finalState.ApplicationVersion = tftypes.StringValue(got.Payload.Spec.ApplicationRef.Version)
		}

		// Handle values field - preserve null state when appropriate
		if got.Payload.Spec.Values != nil {
			if valuesStr := fmt.Sprintf("%v", got.Payload.Spec.Values); valuesStr != "" && valuesStr != "null" && valuesStr != "map[]" {
				finalState.Values = tftypes.StringValue(valuesStr)
			} else if !plan.Values.IsNull() {
				// Preserve planned values if API returned empty but user provided values
				finalState.Values = plan.Values
			}
			// If both are empty/null, leave finalState.Values as is (from plan)
		} else if !plan.Values.IsNull() {
			// Preserve planned values if API didn't return values
			finalState.Values = plan.Values
		}
	}

	// Check and update current status after update
	status, message := r.checkApplicationStatus(ctx, clusterID, namespace, name)
	r.updateStatusFields(finalState, status, message)

	resp.Diagnostics.Append(resp.State.Set(ctx, &finalState)...)
}

func (r *resourceApplication) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	if !r.ValidateResourceBaseDelete(resp) {
		return
	}

	state, ok := kkp.ExtractStateForDelete[applicationState](ctx, req, resp)
	if !ok {
		return
	}

	id := kkp.TrimmedStringValue(state.ID)
	clusterID := kkp.TrimmedStringValue(state.ClusterID)
	namespace := kkp.TrimmedStringValue(state.Namespace)
	name := kkp.TrimmedStringValue(state.Name)
	if clusterID == "" || namespace == "" || name == "" {
		return
	}
	if id == "" {
		id = r.fallbackApplicationID(clusterID, namespace, name)
	}

	aclient := acli.New(r.Client.Transport, nil)
	del := acli.NewDeleteApplicationInstallationParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithNamespace(namespace).
		WithApplicationInstallationName(name)

	if _, err := aclient.DeleteApplicationInstallation(del, nil); err != nil {
		resp.Diagnostics.AddWarning("Delete application warning", err.Error())
	}

	resp.State.RemoveResource(ctx)
}

func (r *resourceApplication) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: cluster_id:namespace:name
	parts := strings.Split(req.ID, ":")
	if len(parts) != 3 {
		resp.Diagnostics.AddError("Unexpected import ID", "Expected 'cluster_id:namespace:name'")
		return
	}

	clusterID := strings.TrimSpace(parts[0])
	namespace := strings.TrimSpace(parts[1])
	name := strings.TrimSpace(parts[2])
	if clusterID == "" || namespace == "" || name == "" {
		resp.Diagnostics.AddError("Invalid import ID", "cluster_id, namespace, and name must be non-empty")
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_id"), clusterID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("namespace"), namespace)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), name)...)
}

// Helper function to check application status
func (r *resourceApplication) checkApplicationStatus(_ context.Context, clusterID, namespace, name string) (status, message string) {
	aclient := acli.New(r.Client.Transport, nil)
	get := acli.NewGetApplicationInstallationParams().
		WithProjectID(r.DefaultProjectID).
		WithClusterID(clusterID).
		WithNamespace(namespace).
		WithApplicationInstallationName(name)

	got, err := aclient.GetApplicationInstallation(get, nil)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") || strings.Contains(err.Error(), "404") {
			return kkp.StatusFailed, "Application not found - installation may have failed"
		}
		return kkp.StatusFailed, fmt.Sprintf("Error checking status: %s", err.Error())
	}

	// Check if application has deletion timestamp (being deleted)
	if !got.Payload.DeletionTimestamp.IsZero() {
		return "deleting", "Application is being deleted"
	}

	// Check application status
	if got.Payload.Status != nil && len(got.Payload.Status.Conditions) > 0 {
		// Find the Ready condition
		for _, condition := range got.Payload.Status.Conditions {
			if condition.Type == "Ready" {
				switch condition.Status {
				case "True":
					return kkp.StatusReady, "Application is ready"
				case "False":
					return kkp.StatusFailed, fmt.Sprintf("Application failed: %s", condition.Message)
				}
			}
		}
		return kkp.StatusInstalling, "Application is still installing"
	}

	// If we get here, application exists but no status yet
	return "installing", "Application installation in progress"
}

// Helper function to wait for application installation with polling
func (r *resourceApplication) waitForApplicationReady(ctx context.Context, clusterID, namespace, name string, maxWaitTime time.Duration) (status, message string) {
	deadline := time.Now().Add(maxWaitTime)
	pollInterval := 15 * time.Second // Longer interval for applications
	maxAttempts := int(maxWaitTime / pollInterval)

	tflog.Info(ctx, "Starting application installation monitoring", map[string]any{
		"cluster_id":    clusterID,
		"namespace":     namespace,
		"name":          name,
		"timeout":       maxWaitTime.String(),
		"poll_interval": pollInterval.String(),
		"max_attempts":  maxAttempts,
	})

	for attempt := 1; time.Now().Before(deadline); attempt++ {
		status, msg := r.checkApplicationStatus(ctx, clusterID, namespace, name)

		// Log progress every 60 seconds (every 4th attempt)
		if attempt == 1 || attempt%4 == 0 || status == kkp.StatusReady || status == kkp.StatusFailed {
			tflog.Info(ctx, "Application installation status check", map[string]any{
				"cluster_id": clusterID,
				"namespace":  namespace,
				"name":       name,
				"attempt":    fmt.Sprintf("%d/%d", attempt, maxAttempts),
				"status":     status,
				"message":    msg,
				"elapsed":    time.Since(deadline.Add(-maxWaitTime)).String(),
			})
		}

		switch status {
		case "ready":
			tflog.Info(ctx, "Application installation completed successfully", map[string]any{
				"cluster_id": clusterID,
				"namespace":  namespace,
				"name":       name,
				"attempts":   attempt,
				"elapsed":    time.Since(deadline.Add(-maxWaitTime)).String(),
			})
			return status, msg
		case kkp.StatusFailed:
			tflog.Error(ctx, "Application installation failed", map[string]any{
				"cluster_id": clusterID,
				"namespace":  namespace,
				"name":       name,
				"attempts":   attempt,
				"elapsed":    time.Since(deadline.Add(-maxWaitTime)).String(),
				"error":      msg,
			})
			return status, msg
		case "deleting":
			tflog.Error(ctx, "Application was marked for deletion during installation", map[string]any{
				"cluster_id": clusterID,
				"namespace":  namespace,
				"name":       name,
				"attempts":   attempt,
			})
			return "failed", "Application was marked for deletion during installation"
		}

		// Wait before next poll, but respect context cancellation
		select {
		case <-ctx.Done():
			tflog.Warn(ctx, "Application installation monitoring canceled", map[string]any{
				"cluster_id": clusterID,
				"namespace":  namespace,
				"name":       name,
				"attempts":   attempt,
			})
			return "failed", "Installation canceled"
		case <-time.After(pollInterval):
			continue
		}
	}

	// Timeout reached
	tflog.Warn(ctx, "Application installation monitoring timed out", map[string]any{
		"cluster_id":   clusterID,
		"namespace":    namespace,
		"name":         name,
		"timeout":      maxWaitTime.String(),
		"max_attempts": maxAttempts,
	})
	return "installing", fmt.Sprintf("Installation timeout after %v - application may still be installing", maxWaitTime)
}

func (r *resourceApplication) fallbackApplicationID(clusterID, namespace, name string) string {
	return fmt.Sprintf("%s/%s/%s", strings.TrimSpace(clusterID), strings.TrimSpace(namespace), strings.TrimSpace(name))
}

// Helper function to update status fields in state
func (r *resourceApplication) updateStatusFields(state *applicationState, status, message string) {
	now := time.Now().Format(time.RFC3339)
	state.Status = tftypes.StringValue(status)
	state.StatusMessage = tftypes.StringValue(message)
	state.LastChecked = tftypes.StringValue(now)
}
