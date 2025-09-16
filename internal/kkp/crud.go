package kkp

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"

	kapi "github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

// ValidateResourceBase performs standard validation checks for all resources.
// Returns true if validation passes, false if validation fails (errors added to diagnostics).
func (rb *ResourceBase) ValidateResourceBase(diagnostics *resource.CreateResponse) bool {
	if rb.Client == nil {
		diagnostics.Diagnostics.AddError("Provider not configured", "KKP client is nil")
		return false
	}
	if rb.DefaultProjectID == "" {
		diagnostics.Diagnostics.AddError("Missing provider project_id", "Set 'project_id' in the provider configuration.")
		return false
	}
	return true
}

// ValidateResourceBaseRead performs standard validation checks for read operations.
func (rb *ResourceBase) ValidateResourceBaseRead(diagnostics *resource.ReadResponse) bool {
	if rb.Client == nil {
		diagnostics.Diagnostics.AddError("Provider not configured", "KKP client is nil")
		return false
	}
	if rb.DefaultProjectID == "" {
		diagnostics.Diagnostics.AddError("Missing provider project_id", "Set 'project_id' in the provider configuration.")
		return false
	}
	return true
}

// ValidateResourceBaseUpdate performs standard validation checks for update operations.
func (rb *ResourceBase) ValidateResourceBaseUpdate(diagnostics *resource.UpdateResponse) bool {
	if rb.Client == nil {
		diagnostics.Diagnostics.AddError("Provider not configured", "KKP client is nil")
		return false
	}
	if rb.DefaultProjectID == "" {
		diagnostics.Diagnostics.AddError("Missing provider project_id", "Set 'project_id' in the provider configuration.")
		return false
	}
	return true
}

// ValidateResourceBaseDelete performs standard validation checks for delete operations.
func (rb *ResourceBase) ValidateResourceBaseDelete(diagnostics *resource.DeleteResponse) bool {
	if rb.Client == nil {
		diagnostics.Diagnostics.AddError("Provider not configured", "KKP client is nil")
		return false
	}
	if rb.DefaultProjectID == "" {
		diagnostics.Diagnostics.AddError("Missing provider project_id", "Set 'project_id' in the provider configuration.")
		return false
	}
	return true
}

// ExtractPlan is a generic helper to extract plan from request with error handling.
func ExtractPlan[T any](ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) (*T, bool) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return nil, false
	}
	return &plan, true
}

// ExtractState is a generic helper to extract state from request with error handling.
func ExtractState[T any](ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) (*T, bool) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return nil, false
	}
	return &state, true
}

// ExtractStateForUpdate is a generic helper to extract state from update request.
func ExtractStateForUpdate[T any](ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) (*T, bool) {
	var plan T
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return nil, false
	}
	return &plan, true
}

// ExtractStateForDelete is a generic helper to extract state from delete request.
func ExtractStateForDelete[T any](ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) (*T, bool) {
	var state T
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return nil, false
	}
	return &state, true
}

// ValidateDataSourceBase performs standard validation checks for all data sources.
// Returns true if validation passes, false if validation fails (errors added to diagnostics).
func (dsb *DataSourceBase) ValidateDataSourceBase(diagnostics *datasource.ReadResponse) bool {
	if dsb.Client == nil {
		diagnostics.Diagnostics.AddError("Provider not configured", "KKP client is nil")
		return false
	}
	if dsb.DefaultProjectID == "" {
		diagnostics.Diagnostics.AddError("Missing provider project_id", "Set 'project_id' in the provider configuration.")
		return false
	}
	return true
}

// ConfigureDataSource performs standard configuration for all data sources.
func (dsb *DataSourceBase) ConfigureDataSource(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	meta, ok := req.ProviderData.(*ProviderMeta)
	if !ok || meta.Client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Unexpected provider data type")
		return
	}
	dsb.Client = meta.Client
	dsb.DefaultProjectID = strings.TrimSpace(meta.DefaultProjectID)
}

// ConfigureResource performs standard configuration for all resources.
func (rb *ResourceBase) ConfigureResource(req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	meta, ok := req.ProviderData.(*ProviderMeta)
	if !ok || meta.Client == nil {
		resp.Diagnostics.AddError("Provider not configured", "Unexpected provider data type")
		return
	}
	rb.Client = meta.Client
	rb.DefaultProjectID = strings.TrimSpace(meta.DefaultProjectID)
}

// FetchClusters retrieves clusters from KKP API for a given project
func (dsb *DataSourceBase) FetchClusters() ([]*models.Cluster, error) {
	pcli := kapi.New(dsb.Client.Transport, nil)
	params := kapi.NewListClustersV2Params().WithProjectID(dsb.DefaultProjectID)
	resp, err := pcli.ListClustersV2(params, nil)
	if err != nil {
		return nil, err
	}
	if resp.Payload == nil {
		return nil, nil
	}
	return resp.Payload.Clusters, nil
}
