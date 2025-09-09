package kkp

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	kapi "github.com/kubermatic/go-kubermatic/client/project"
	"github.com/kubermatic/go-kubermatic/models"
)

// Accept KKP's stringy enums like "HealthStatusUp" or plain "Up".
func statusUp(s models.HealthStatus) bool {
	v := strings.ToLower(string(s))
	return v == "up" || strings.HasSuffix(v, "up") || v == "1" || v == "true"
}

// HealthReady performs lenient core readiness check: apiserver+controller+scheduler+etcd must be up.
func HealthReady(h *models.ClusterHealth) bool {
	if h == nil {
		return false
	}
	core := []bool{
		statusUp(h.Apiserver),
		statusUp(h.Controller),
		statusUp(h.Scheduler),
		statusUp(h.Etcd),
	}
	for _, ok := range core {
		if !ok {
			return false
		}
	}
	return true
}

// StatusURLReady returns (url, true) if a non-empty URL field is present.
func StatusURLReady(status any) (string, bool) {
	if status == nil {
		return "", false
	}
	v := reflect.ValueOf(status)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "", false
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return "", false
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		name := strings.ToLower(sf.Name)
		tag := strings.ToLower(strings.Split(sf.Tag.Get("json"), ",")[0])
		if name != "url" && tag != "url" {
			continue
		}
		f := v.Field(i)
		if f.Kind() == reflect.String {
			u := strings.TrimSpace(f.String())
			if u != "" {
				return u, true
			}
		}
	}
	return "", false
}

// Poll calls fn immediately and then every `interval` until fn returns done=true,
// an error, or the context is cancelled/expired.
// It returns fn's error (if any) or ctx.Err() when the context ends.
func Poll(ctx context.Context, interval time.Duration, fn func(context.Context) (done bool, err error)) error {
	// First attempt (no initial delay)
	done, err := fn(ctx)
	if done || err != nil {
		return err
	}

	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			done, err = fn(ctx)
			if done || err != nil {
				return err
			}
			// Reset timer safely
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			timer.Reset(interval)
		}
	}
}

// PollWithTimeout is a convenience wrapper that applies a timeout on top of ctx.
func PollWithTimeout(parent context.Context, interval, timeout time.Duration, fn func(context.Context) (bool, error)) error {
	ctx := parent
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(parent, timeout)
		defer cancel()
	}
	return Poll(ctx, interval, fn)
}

// WaitForClusterReady waits for cluster to become healthy (for create operations)
func (c *ClusterHealthChecker) WaitForClusterReady(ctx context.Context) error {
	return c.WaitForClusterReadyWithTimeout(ctx, 10*time.Second, 15*time.Minute)
}

// WaitForClusterReadyWithTimeout waits for cluster to become healthy with custom timeout
func (c *ClusterHealthChecker) WaitForClusterReadyWithTimeout(ctx context.Context, interval, timeout time.Duration) error {
	return PollWithTimeout(ctx, interval, timeout, func(pc context.Context) (bool, error) {
		pcli := kapi.New(c.Client.Transport, nil)

		// Check health first (authoritative for readiness)
		hres, herr := pcli.GetClusterHealthV2(
			kapi.NewGetClusterHealthV2Params().
				WithProjectID(c.ProjectID).
				WithClusterID(c.ClusterID),
			nil,
		)
		if herr == nil && hres != nil && hres.Payload != nil {
			h := hres.Payload
			tflog.Debug(pc, "cluster health check", map[string]any{
				"cluster_id": c.ClusterID,
				"apiserver":  string(h.Apiserver),
				"controller": string(h.Controller),
				"scheduler":  string(h.Scheduler),
				"etcd":       string(h.Etcd),
			})
			if HealthReady(h) {
				tflog.Info(pc, "cluster is healthy", map[string]any{"cluster_id": c.ClusterID})
				return true, nil
			}
		} else if herr != nil {
			tflog.Debug(pc, "health check error", map[string]any{
				"cluster_id": c.ClusterID,
				"error":      herr.Error(),
			})
		}

		// Optional: get cluster status for breadcrumbs
		if g, gerr := pcli.GetClusterV2(
			kapi.NewGetClusterV2Params().
				WithProjectID(c.ProjectID).
				WithClusterID(c.ClusterID),
			nil,
		); gerr == nil && g != nil && g.Payload != nil && g.Payload.Status != nil {
			if u := strings.TrimSpace(g.Payload.Status.URL); u != "" {
				tflog.Debug(pc, "cluster API URL available", map[string]any{
					"cluster_id": c.ClusterID,
					"api_url":    u,
				})
			}
		}

		return false, nil
	})
}

// WaitForClusterUpdated waits for cluster update to complete with spec verification
func (c *ClusterHealthChecker) WaitForClusterUpdated(ctx context.Context, expectedSpec ClusterUpdateSpec) error {
	return c.WaitForClusterUpdatedWithTimeout(ctx, 10*time.Second, 45*time.Minute, expectedSpec)
}

// WaitForClusterUpdatedWithTimeout waits for cluster update with custom timeout and spec verification
func (c *ClusterHealthChecker) WaitForClusterUpdatedWithTimeout(ctx context.Context, interval, timeout time.Duration, expectedSpec ClusterUpdateSpec) error {
	state := &clusterUpdateState{
		lastNote:       "",
		seenTransition: false,
	}

	return PollWithTimeout(ctx, interval, timeout, func(pc context.Context) (bool, error) {
		return c.checkClusterUpdateProgress(pc, expectedSpec, state)
	})
}

// WaitForClusterDeleted waits for cluster to be completely deleted
func (c *ClusterHealthChecker) WaitForClusterDeleted(ctx context.Context) error {
	return c.WaitForClusterDeletedWithTimeout(ctx, 10*time.Second, 20*time.Minute)
}

// WaitForClusterDeletedWithTimeout waits for cluster deletion with custom timeout
func (c *ClusterHealthChecker) WaitForClusterDeletedWithTimeout(ctx context.Context, interval, timeout time.Duration) error {
	return PollWithTimeout(ctx, interval, timeout, func(pc context.Context) (bool, error) {
		pcli := kapi.New(c.Client.Transport, nil)

		g, gerr := pcli.GetClusterV2(
			kapi.NewGetClusterV2Params().
				WithProjectID(c.ProjectID).
				WithClusterID(c.ClusterID),
			nil,
		)
		if gerr != nil {
			low := strings.ToLower(gerr.Error())
			if strings.Contains(low, "404") || strings.Contains(low, "not found") {
				tflog.Info(pc, "cluster deletion confirmed", map[string]any{"cluster_id": c.ClusterID})
				return true, nil
			}
			tflog.Warn(pc, "cluster deletion check failed", map[string]any{
				"cluster_id": c.ClusterID,
				"error":      gerr.Error(),
			})
			return false, nil
		}

		// Cluster still exists, check progress
		if g != nil && g.Payload != nil && g.Payload.Status != nil {
			if u := strings.TrimSpace(g.Payload.Status.URL); u == "" {
				tflog.Debug(pc, "cluster API URL cleared", map[string]any{"cluster_id": c.ClusterID})
			} else {
				tflog.Debug(pc, "cluster still has API URL", map[string]any{
					"cluster_id": c.ClusterID,
					"api_url":    u,
				})
			}
		}

		// Optional: check health for deletion progress (often 404s before cluster is gone)
		if hres, herr := pcli.GetClusterHealthV2(
			kapi.NewGetClusterHealthV2Params().
				WithProjectID(c.ProjectID).
				WithClusterID(c.ClusterID),
			nil,
		); herr == nil && hres != nil && hres.Payload != nil {
			h := hres.Payload
			tflog.Debug(pc, "deleting cluster health", map[string]any{
				"cluster_id": c.ClusterID,
				"apiserver":  string(h.Apiserver),
				"controller": string(h.Controller),
				"scheduler":  string(h.Scheduler),
				"etcd":       string(h.Etcd),
			})
			tflog.Debug(pc, "cluster health during deletion", map[string]any{
				"cluster_id": c.ClusterID,
				"apiserver":  string(h.Apiserver),
				"controller": string(h.Controller),
				"scheduler":  string(h.Scheduler),
				"etcd":       string(h.Etcd),
			})
		} else if herr != nil && strings.Contains(strings.ToLower(herr.Error()), "404") {
			tflog.Debug(pc, "cluster health 404 during deletion (control-plane likely gone)", map[string]any{"cluster_id": c.ClusterID})
		}

		return false, nil
	})
}

// WaitForMachineDeploymentReady waits for machine deployment to become ready
func (c *MachineDeploymentHealthChecker) WaitForMachineDeploymentReady(ctx context.Context) error {
	return c.WaitForMachineDeploymentReadyWithTimeout(ctx, 10*time.Second, 15*time.Minute)
}

// WaitForMachineDeploymentReadyWithTimeout waits for machine deployment with custom timeout
func (c *MachineDeploymentHealthChecker) WaitForMachineDeploymentReadyWithTimeout(ctx context.Context, interval, timeout time.Duration) error {
	return PollWithTimeout(ctx, interval, timeout, func(pc context.Context) (bool, error) {
		deployment, err := c.fetchMachineDeploymentFromList(pc)
		if err != nil {
			return c.handleFetchError(pc, err)
		}

		if deployment == nil {
			return false, nil // Continue polling
		}

		return c.checkMachineDeploymentReadiness(pc, deployment)
	})
}

// fetchMachineDeploymentFromList retrieves the target machine deployment from the list API.
func (c *MachineDeploymentHealthChecker) fetchMachineDeploymentFromList(ctx context.Context) (*models.NodeDeployment, error) {
	pcli := kapi.New(c.Client.Transport, nil)

	// Use ListMachineDeployments instead of GetMachineDeployment to avoid TextConsumer issues
	listResp, listErr := pcli.ListMachineDeployments(
		kapi.NewListMachineDeploymentsParams().
			WithProjectID(c.ProjectID).
			WithClusterID(c.ClusterID),
		nil,
	)
	if listErr != nil {
		return nil, listErr
	}

	if listResp == nil || listResp.Payload == nil {
		tflog.Debug(ctx, "machine deployment list response not available", map[string]any{
			"cluster_id":            c.ClusterID,
			"machine_deployment_id": c.MachineDeploymentID,
		})
		return nil, nil
	}

	return c.findTargetDeploymentInList(ctx, listResp.Payload)
}

// handleFetchError handles errors from fetching machine deployment list.
func (c *MachineDeploymentHealthChecker) handleFetchError(ctx context.Context, err error) (bool, error) {
	errorMsg := err.Error()
	tflog.Debug(ctx, "list machine deployments failed during ready wait", map[string]any{
		"cluster_id":            c.ClusterID,
		"machine_deployment_id": c.MachineDeploymentID,
		"error":                 errorMsg,
	})

	// If it's a parsing error, continue polling
	if strings.Contains(errorMsg, "TextConsumer") || strings.Contains(errorMsg, "TextUnmarshaler") {
		tflog.Debug(ctx, "API parsing error during machine deployment listing", map[string]any{
			"cluster_id":            c.ClusterID,
			"machine_deployment_id": c.MachineDeploymentID,
			"parsing_error":         errorMsg,
		})
		return false, nil
	}

	// For other errors, continue polling
	tflog.Debug(ctx, "continuing to poll despite list error", map[string]any{
		"cluster_id":            c.ClusterID,
		"machine_deployment_id": c.MachineDeploymentID,
		"error":                 errorMsg,
	})
	return false, nil
}

// findTargetDeploymentInList finds the specific machine deployment in the API response.
func (c *MachineDeploymentHealthChecker) findTargetDeploymentInList(ctx context.Context, deployments []*models.NodeDeployment) (*models.NodeDeployment, error) {
	var targetDeployment *models.NodeDeployment
	var availableIDs []string

	for _, deployment := range deployments {
		if deployment != nil {
			availableIDs = append(availableIDs, deployment.ID)
			if deployment.ID == c.MachineDeploymentID {
				targetDeployment = deployment
				break
			}
		}
	}

	if targetDeployment == nil {
		tflog.Debug(ctx, "machine deployment not found in list - still being created", map[string]any{
			"cluster_id":               c.ClusterID,
			"machine_deployment_id":    c.MachineDeploymentID,
			"total_deployments":        len(deployments),
			"available_deployment_ids": availableIDs,
		})
		return nil, nil
	}

	return targetDeployment, nil
}

// checkMachineDeploymentReadiness evaluates if the machine deployment is ready.
func (c *MachineDeploymentHealthChecker) checkMachineDeploymentReadiness(ctx context.Context, deployment *models.NodeDeployment) (bool, error) {
	if deployment.Status == nil {
		tflog.Debug(ctx, "machine deployment status not available", map[string]any{
			"cluster_id":            c.ClusterID,
			"machine_deployment_id": c.MachineDeploymentID,
		})
		return false, nil
	}

	status := deployment.Status

	// Log current status for debugging
	tflog.Debug(ctx, "machine deployment current status", map[string]any{
		"cluster_id":            c.ClusterID,
		"machine_deployment_id": c.MachineDeploymentID,
		"available_replicas":    status.AvailableReplicas,
		"desired_replicas":      status.Replicas,
		"ready_replicas":        status.ReadyReplicas,
		"updated_replicas":      status.UpdatedReplicas,
	})

	expectedReplicas, err := c.calculateExpectedReplicas(status)
	if err != nil {
		return false, err
	}

	return c.evaluateReadinessCondition(ctx, status, expectedReplicas)
}

// calculateExpectedReplicas determines the expected number of replicas.
func (c *MachineDeploymentHealthChecker) calculateExpectedReplicas(status *models.MachineDeploymentStatus) (int32, error) {
	// Default: use desired replicas from API
	expectedReplicas := status.Replicas

	if c.ExpectedReplicas > 0 {
		safeReplicas, err := SafeInt32(c.ExpectedReplicas)
		if err != nil {
			return 0, fmt.Errorf("invalid expected replicas value: %w", err)
		}
		expectedReplicas = safeReplicas // Use expected replicas if specified
	}

	return expectedReplicas, nil
}

// evaluateReadinessCondition checks if the machine deployment meets readiness criteria.
func (c *MachineDeploymentHealthChecker) evaluateReadinessCondition(ctx context.Context, status *models.MachineDeploymentStatus, expectedReplicas int32) (bool, error) {
	// Available replicas should match expected replicas
	if status.AvailableReplicas >= expectedReplicas && expectedReplicas > 0 {
		tflog.Info(ctx, "machine deployment is ready", map[string]any{
			"cluster_id":            c.ClusterID,
			"machine_deployment_id": c.MachineDeploymentID,
			"available_replicas":    status.AvailableReplicas,
			"desired_replicas":      status.Replicas,
			"expected_replicas":     expectedReplicas,
			"ready_replicas":        status.ReadyReplicas,
			"updated_replicas":      status.UpdatedReplicas,
		})
		return true, nil
	}

	tflog.Debug(ctx, "machine deployment not ready yet", map[string]any{
		"cluster_id":            c.ClusterID,
		"machine_deployment_id": c.MachineDeploymentID,
		"available_replicas":    status.AvailableReplicas,
		"desired_replicas":      status.Replicas,
		"expected_replicas":     expectedReplicas,
		"reason":                "available_replicas < expected_replicas or expected_replicas <= 0",
	})

	return false, nil
}

// WaitForMachineDeploymentDeleted waits for machine deployment to be deleted
func (c *MachineDeploymentHealthChecker) WaitForMachineDeploymentDeleted(ctx context.Context) error {
	return c.WaitForMachineDeploymentDeletedWithTimeout(ctx, 10*time.Second, 10*time.Minute)
}

// WaitForMachineDeploymentDeletedWithTimeout waits for machine deployment deletion with custom timeout
func (c *MachineDeploymentHealthChecker) WaitForMachineDeploymentDeletedWithTimeout(ctx context.Context, interval, timeout time.Duration) error {
	return PollWithTimeout(ctx, interval, timeout, func(pc context.Context) (bool, error) {
		pcli := kapi.New(c.Client.Transport, nil)

		// Use ListMachineDeployments instead of GetMachineDeployment to avoid TextConsumer issues
		listResp, listErr := pcli.ListMachineDeployments(
			kapi.NewListMachineDeploymentsParams().
				WithProjectID(c.ProjectID).
				WithClusterID(c.ClusterID),
			nil,
		)
		if listErr != nil {
			low := strings.ToLower(listErr.Error())
			if strings.Contains(low, "404") || strings.Contains(low, "not found") {
				tflog.Info(pc, "cluster not found during machine deployment deletion check - cluster likely deleted", map[string]any{
					"cluster_id":            c.ClusterID,
					"machine_deployment_id": c.MachineDeploymentID,
				})
				return true, nil
			}

			tflog.Debug(pc, "list machine deployments failed during deletion wait", map[string]any{
				"cluster_id":            c.ClusterID,
				"machine_deployment_id": c.MachineDeploymentID,
				"error":                 listErr.Error(),
			})
			return false, nil
		}

		if listResp == nil || listResp.Payload == nil {
			tflog.Debug(pc, "machine deployment list response not available during deletion", map[string]any{
				"cluster_id":            c.ClusterID,
				"machine_deployment_id": c.MachineDeploymentID,
			})
			return false, nil
		}

		// Check if our specific machine deployment is in the list
		for _, deployment := range listResp.Payload {
			if deployment != nil && deployment.ID == c.MachineDeploymentID {
				tflog.Debug(pc, "machine deployment still exists", map[string]any{
					"cluster_id":            c.ClusterID,
					"machine_deployment_id": c.MachineDeploymentID,
				})
				return false, nil
			}
		}

		// Machine deployment not found in list - deletion confirmed
		tflog.Info(pc, "machine deployment deletion confirmed", map[string]any{
			"cluster_id":            c.ClusterID,
			"machine_deployment_id": c.MachineDeploymentID,
		})
		return true, nil
	})
}

// clusterUpdateState tracks the state during cluster update monitoring.
type clusterUpdateState struct {
	lastNote       string
	seenTransition bool
}

// checkClusterUpdateProgress checks the progress of a cluster update operation.
func (c *ClusterHealthChecker) checkClusterUpdateProgress(ctx context.Context, expectedSpec ClusterUpdateSpec, state *clusterUpdateState) (bool, error) {
	// Get current cluster state
	cluster, err := c.getClusterForUpdate(ctx)
	if err != nil {
		return false, nil // Continue polling on errors
	}
	if cluster == nil {
		state.lastNote = "cluster spec not available"
		return false, nil
	}

	// Check if cluster spec is up to date
	specUpToDate, note := c.isClusterSpecUpToDate(cluster.Spec, expectedSpec)
	if !specUpToDate {
		state.lastNote = note
		tflog.Debug(ctx, "cluster spec not yet updated", map[string]any{
			"cluster_id": c.ClusterID,
			"status":     note,
		})
		return false, nil
	}

	// Spec is updated, now check health and transitions
	return c.checkClusterHealthAndTransitions(ctx, cluster, expectedSpec, state)
}

// getClusterForUpdate fetches the cluster information for update monitoring.
func (c *ClusterHealthChecker) getClusterForUpdate(ctx context.Context) (*models.Cluster, error) {
	pcli := kapi.New(c.Client.Transport, nil)
	g, gerr := pcli.GetClusterV2(
		kapi.NewGetClusterV2Params().
			WithProjectID(c.ProjectID).
			WithClusterID(c.ClusterID),
		nil,
	)
	if gerr != nil {
		tflog.Debug(ctx, "get cluster failed during update wait", map[string]any{
			"cluster_id": c.ClusterID,
			"error":      gerr.Error(),
		})
		return nil, gerr
	}

	if g == nil || g.Payload == nil || g.Payload.Spec == nil {
		return nil, nil
	}

	return g.Payload, nil
}

// isClusterSpecUpToDate checks if the cluster spec reflects the expected changes.
func (c *ClusterHealthChecker) isClusterSpecUpToDate(spec *models.ClusterSpec, expectedSpec ClusterUpdateSpec) (bool, string) {
	if expectedSpec.K8sVersion != "" && string(spec.Version) != expectedSpec.K8sVersion {
		return false, fmt.Sprintf("k8s version: want=%s got=%s", expectedSpec.K8sVersion, string(spec.Version))
	}

	if spec.CniPlugin != nil {
		if expectedSpec.CNIType != "" && string(spec.CniPlugin.Type) != expectedSpec.CNIType {
			return false, fmt.Sprintf("CNI type: want=%s got=%s", expectedSpec.CNIType, string(spec.CniPlugin.Type))
		}
		if expectedSpec.CNIVersion != "" && spec.CniPlugin.Version != expectedSpec.CNIVersion {
			return false, fmt.Sprintf("CNI version: want=%s got=%s", expectedSpec.CNIVersion, spec.CniPlugin.Version)
		}
	}

	return true, ""
}

// checkClusterHealthAndTransitions monitors cluster health and API server transitions during update.
func (c *ClusterHealthChecker) checkClusterHealthAndTransitions(ctx context.Context, cluster *models.Cluster, expectedSpec ClusterUpdateSpec, state *clusterUpdateState) (bool, error) {
	// Get cluster health
	health, err := c.getClusterHealth(ctx)
	if err != nil {
		return c.handleHealthCheckError(ctx, cluster, err, state)
	}
	if health == nil {
		return false, nil
	}

	// Check for API server transitions
	if c.checkAPIServerTransition(ctx, health, state) {
		return false, nil // Continue monitoring transition
	}

	// Check if cluster is healthy and update is complete
	return c.evaluateUpdateCompletion(ctx, cluster, health, expectedSpec, state)
}

// getClusterHealth fetches the cluster health information.
func (c *ClusterHealthChecker) getClusterHealth(_ context.Context) (*models.ClusterHealth, error) {
	pcli := kapi.New(c.Client.Transport, nil)
	hres, herr := pcli.GetClusterHealthV2(
		kapi.NewGetClusterHealthV2Params().
			WithProjectID(c.ProjectID).
			WithClusterID(c.ClusterID),
		nil,
	)
	if herr != nil {
		return nil, herr
	}
	if hres == nil || hres.Payload == nil {
		return nil, nil
	}
	return hres.Payload, nil
}

// handleHealthCheckError handles health check errors during update monitoring.
func (c *ClusterHealthChecker) handleHealthCheckError(ctx context.Context, cluster *models.Cluster, err error, state *clusterUpdateState) (bool, error) {
	// Health check failures during update might indicate transition
	if !state.seenTransition {
		state.seenTransition = true
		tflog.Info(ctx, "detected health check failure during update (likely transition)", map[string]any{
			"cluster_id": c.ClusterID,
			"error":      err.Error(),
		})
	}
	state.lastNote = "spec updated but health check failed: " + err.Error()
	tflog.Debug(ctx, "post-update health error", map[string]any{
		"cluster_id": c.ClusterID,
		"error":      err.Error(),
	})

	// Add API URL for debugging
	if cluster.Status != nil {
		if u := strings.TrimSpace(cluster.Status.URL); u != "" {
			state.lastNote += " (api=" + u + ")"
		}
	}

	return false, nil
}

// checkAPIServerTransition monitors API server transitions during update.
func (c *ClusterHealthChecker) checkAPIServerTransition(ctx context.Context, health *models.ClusterHealth, state *clusterUpdateState) bool {
	apiServerDown := !statusUp(health.Apiserver)
	if apiServerDown {
		state.seenTransition = true
		tflog.Info(ctx, "detected API server transition during update", map[string]any{
			"cluster_id": c.ClusterID,
			"apiserver":  string(health.Apiserver),
		})
		state.lastNote = fmt.Sprintf("API server transitioning: %s", health.Apiserver)
		return true
	}
	return false
}

// evaluateUpdateCompletion determines if the cluster update is complete.
func (c *ClusterHealthChecker) evaluateUpdateCompletion(ctx context.Context, cluster *models.Cluster, health *models.ClusterHealth, expectedSpec ClusterUpdateSpec, state *clusterUpdateState) (bool, error) {
	tflog.Debug(ctx, "post-update health", map[string]any{
		"cluster_id":      c.ClusterID,
		"apiserver":       string(health.Apiserver),
		"controller":      string(health.Controller),
		"scheduler":       string(health.Scheduler),
		"etcd":            string(health.Etcd),
		"seen_transition": state.seenTransition,
	})

	if HealthReady(health) {
		// For K8s version updates, we should see the API server go down and come back up
		// If we haven't seen a transition yet for K8s updates, keep waiting a bit longer
		if expectedSpec.K8sVersion != "" && !state.seenTransition {
			tflog.Info(ctx, "cluster healthy but no API server transition detected yet", map[string]any{
				"cluster_id":       c.ClusterID,
				"expected_version": expectedSpec.K8sVersion,
			})
			state.lastNote = "healthy but no API server restart observed yet"
			return false, nil
		}

		tflog.Info(ctx, "cluster update complete", map[string]any{
			"cluster_id":      c.ClusterID,
			"version":         cluster.Spec.Version,
			"cni_type":        cluster.Spec.CniPlugin.Type,
			"cni_version":     cluster.Spec.CniPlugin.Version,
			"seen_transition": state.seenTransition,
		})
		return true, nil
	}

	state.lastNote = fmt.Sprintf("spec updated but health: apiserver=%s controller=%s scheduler=%s etcd=%s",
		health.Apiserver, health.Controller, health.Scheduler, health.Etcd)

	// Add API URL for debugging
	if cluster.Status != nil {
		if u := strings.TrimSpace(cluster.Status.URL); u != "" {
			state.lastNote += " (api=" + u + ")"
		}
	}

	return false, nil
}
