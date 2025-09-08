# kkp_machine_deployments_v2

Provides information about machine deployments in a KKP cluster.

## Example Usage

```terraform
# Get all machine deployments for a cluster
data "kkp_machine_deployments_v2" "cluster_workers" {
  cluster_id = kkp_cluster_v2.example.id
}

# Output machine deployment information
output "machine_deployment_names" {
  value = [for md in data.kkp_machine_deployments_v2.cluster_workers.machine_deployments : md.name]
}

# Filter machine deployments by cloud provider
locals {
  openstack_deployments = [
    for md in data.kkp_machine_deployments_v2.cluster_workers.machine_deployments : md
    if md.cloud == "openstack"
  ]
}

# Calculate total worker nodes
locals {
  total_workers = sum([
    for md in data.kkp_machine_deployments_v2.cluster_workers.machine_deployments : md.replicas
  ])
}
```

## Argument Reference

The following arguments are supported:

- `cluster_id` - (Required) The ID of the cluster to query machine deployments for.

## Attributes Reference

The following attributes are exported:

- `id` - The data source identifier.
- `cluster_id` - The cluster ID that was queried.
- `machine_deployments` - A list of machine deployment objects with the following attributes:
  - `id` - The unique identifier of the machine deployment.
  - `name` - The name of the machine deployment.
  - `cloud` - The cloud provider (e.g., "openstack", "aws").
  - `replicas` - The desired number of replicas.
  - `ready_replicas` - The number of ready replicas.
  - `available_replicas` - The number of available replicas.
  - `creation_time` - The timestamp when the machine deployment was created.
  - `spec` - The machine deployment specification containing:
    - `min_replicas` - Minimum number of replicas for autoscaling.
    - `max_replicas` - Maximum number of replicas for autoscaling.
    - `paused` - Whether the machine deployment is paused.
  - `cloud_spec` - Cloud-specific configuration (structure varies by provider).