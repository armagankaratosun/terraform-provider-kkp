# kkp_applications_v2

Provides information about applications deployed on a KKP cluster.

## Example Usage

```terraform
# Get all applications for a cluster
data "kkp_applications_v2" "cluster_apps" {
  cluster_id = kkp_cluster_v2.example.id
}

# Output application information
output "deployed_applications" {
  value = [
    for app in data.kkp_applications_v2.cluster_apps.applications : {
      name      = app.name
      namespace = app.namespace
      status    = app.status
    }
  ]
}

# Filter applications by namespace
locals {
  database_apps = [
    for app in data.kkp_applications_v2.cluster_apps.applications : app
    if app.namespace == "database"
  ]
  
  monitoring_apps = [
    for app in data.kkp_applications_v2.cluster_apps.applications : app
    if app.namespace == "monitoring"
  ]
}

# Filter applications by status
locals {
  ready_apps = [
    for app in data.kkp_applications_v2.cluster_apps.applications : app
    if app.status == "Ready"
  ]
  
  failed_apps = [
    for app in data.kkp_applications_v2.cluster_apps.applications : app
    if app.status == "Failed"
  ]
}

# Check if specific applications are deployed
locals {
  has_redis = contains([
    for app in data.kkp_applications_v2.cluster_apps.applications : "${app.namespace}/${app.name}"
  ], "database/redis")
  
  has_postgresql = contains([
    for app in data.kkp_applications_v2.cluster_apps.applications : "${app.namespace}/${app.name}"
  ], "database/postgresql")
}

# Group applications by namespace
locals {
  apps_by_namespace = {
    for app in data.kkp_applications_v2.cluster_apps.applications :
    app.namespace => app...
  }
}
```

## Argument Reference

The following arguments are supported:

- `cluster_id` - (Required) The ID of the cluster to query applications for.

## Attributes Reference

The following attributes are exported:

- `id` - The data source identifier.
- `cluster_id` - The cluster ID that was queried.
- `applications` - A list of application objects with the following attributes:
  - `id` - The unique identifier of the application.
  - `name` - The name of the application.
  - `namespace` - The Kubernetes namespace where the application is deployed.
  - `status` - The current status of the application (e.g., "Ready", "Installing", "Failed").
  - `created_at` - The timestamp when the application was deployed.
  - `source` - Information about the application source:
    - `chart_name` - The name of the Helm chart.
    - `chart_version` - The version of the Helm chart.
    - `repository_name` - The name of the Helm repository.
    - `repository_url` - The URL of the Helm repository.
  - `values` - A map of Helm values used for the application.
  - `labels` - A map of labels assigned to the application.