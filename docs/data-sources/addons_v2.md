# kkp_addons_v2

Provides information about addons installed on a KKP cluster.

## Example Usage

```terraform
# Get all addons for a cluster
data "kkp_addons_v2" "cluster_addons" {
  cluster_id = kkp_cluster_v2.example.id
}

# Output addon information
output "installed_addons" {
  value = [for addon in data.kkp_addons_v2.cluster_addons.addons : addon.name]
}

# Filter addons by status
locals {
  ready_addons = [
    for addon in data.kkp_addons_v2.cluster_addons.addons : addon
    if addon.status == "Ready"
  ]
  
  failed_addons = [
    for addon in data.kkp_addons_v2.cluster_addons.addons : addon
    if addon.status == "Failed"
  ]
}

# Check if specific addons are installed
locals {
  has_ingress_controller = contains([
    for addon in data.kkp_addons_v2.cluster_addons.addons : addon.name
  ], "nginx-ingress-controller")
  
  has_monitoring = contains([
    for addon in data.kkp_addons_v2.cluster_addons.addons : addon.name
  ], "prometheus")
}
```

## Argument Reference

The following arguments are supported:

- `cluster_id` - (Required) The ID of the cluster to query addons for.

## Attributes Reference

The following attributes are exported:

- `id` - The data source identifier.
- `cluster_id` - The cluster ID that was queried.
- `addons` - A list of addon objects with the following attributes:
  - `id` - The unique identifier of the addon.
  - `name` - The name of the addon.
  - `status` - The current status of the addon (e.g., "Ready", "Installing", "Failed").
  - `created_at` - The timestamp when the addon was installed.
  - `variables` - A map of configuration variables used for the addon.
  - `labels` - A map of labels assigned to the addon.