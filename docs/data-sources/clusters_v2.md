# kkp_clusters_v2

Provides information about KKP clusters in a project.

## Example Usage

```terraform
# Get all clusters in the project
data "kkp_clusters_v2" "all" {}

# Output cluster information
output "cluster_names" {
  value = [for cluster in data.kkp_clusters_v2.all.clusters : cluster.name]
}

output "openstack_clusters" {
  value = [for cluster in data.kkp_clusters_v2.all.clusters : cluster if cluster.cloud == "openstack"]
}
```

## Attributes Reference

The following attributes are exported:

- `id` - The data source identifier.
- `clusters` - A list of cluster objects with the following attributes:
  - `id` - The unique identifier of the cluster.
  - `name` - The name of the cluster.
  - `cloud` - The cloud provider (e.g., "openstack", "aws").
  - `datacenter_name` - The datacenter name where the cluster is deployed.
  - `version` - The Kubernetes version.
  - `type` - The cluster type.
  - `creation_time` - The timestamp when the cluster was created.
  - `labels` - A map of labels assigned to the cluster.