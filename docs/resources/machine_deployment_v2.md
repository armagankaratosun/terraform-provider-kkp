# kkp_machine_deployment_v2

Manages machine deployments (worker nodes) for KKP clusters. Machine deployments define groups of worker nodes with specific configurations and support autoscaling.

## Example Usage

### OpenStack Machine Deployment

```terraform
resource "kkp_machine_deployment_v2" "workers" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "worker-nodes"
  replicas   = 3
  
  cloud = "openstack"
  
  openstack {
    flavor             = "m1.medium"
    image              = "Ubuntu 22.04"
    use_floating_ip    = true
    disk_size          = 50
    availability_zone  = "nova"
  }
  
  spec {
    min_replicas = 1
    max_replicas = 10
    
    operating_system {
      ubuntu {
        dist_upgrade_on_boot = true
      }
    }
    
    taints = [
      {
        key    = "node-role"
        value  = "worker"
        effect = "NoSchedule"
      }
    ]
  }
}
```

### OpenStack Machine Deployment (using preset)

```terraform
resource "kkp_machine_deployment_v2" "preset_workers" {
  cluster_id = kkp_cluster_v2.openstack_preset.id
  name       = "preset-workers"
  replicas   = 5
  
  cloud  = "openstack"
  preset = "openstack-production-preset"
  
  # OpenStack block can override preset values
  openstack {
    flavor = "m1.large"  # Override preset flavor
  }
  
  spec {
    min_replicas = 2
    max_replicas = 20
    
    operating_system {
      ubuntu {
        dist_upgrade_on_boot = false
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

### General Arguments

- `cluster_id` - (Required) The ID of the cluster this machine deployment belongs to.
- `name` - (Required) The name of the machine deployment.
- `replicas` - (Required) The number of worker nodes to create.
- `cloud` - (Required) The cloud provider. Currently only `openstack` is supported.
- `preset` - (Optional) KKP preset name for predefined cloud configurations.

### OpenStack Configuration

The `openstack` block supports:

- `flavor` - (Required) OpenStack flavor name (e.g., "m1.small", "standard.medium").
- `image` - (Required) OpenStack image name or UUID.
- `use_floating_ip` - (Optional) Whether to assign floating IPs to nodes. Defaults to `false`.
- `disk_size` - (Optional) Boot disk size in GB. Defaults to `25`.
- `availability_zone` - (Optional) OpenStack availability zone.

### Note on Additional Cloud Providers

Currently, only OpenStack is supported. Other cloud providers are not implemented but may be added in future releases.

### Spec Configuration

The `spec` block supports:

- `min_replicas` - (Optional) Minimum number of replicas for autoscaling.
- `max_replicas` - (Optional) Maximum number of replicas for autoscaling.
- `max_surge` - (Optional) Maximum number of nodes that can be created above the desired number during updates.
- `max_unavailable` - (Optional) Maximum number of nodes that can be unavailable during updates.
- `node_deployment_strategy` - (Optional) Strategy for node updates. Values: `RollingUpdate`, `Recreate`.
- `paused` - (Optional) Whether the machine deployment is paused.
- `dynamic_config` - (Optional) Whether to enable dynamic kubelet configuration.

#### Operating System Configuration

The `operating_system` block supports:

**Ubuntu:**
```terraform
operating_system {
  ubuntu {
    dist_upgrade_on_boot = true
  }
}
```

**CentOS:**
```terraform
operating_system {
  centos {
    dist_upgrade_on_boot = false
  }
}
```

#### Taints

The `taints` block supports:

- `key` - (Required) The taint key.
- `value` - (Optional) The taint value.
- `effect` - (Required) The taint effect. Values: `NoSchedule`, `PreferNoSchedule`, `NoExecute`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the machine deployment.
- `creation_time` - The timestamp when the machine deployment was created.
- `ready_replicas` - The number of ready replicas.
- `available_replicas` - The number of available replicas.

## Import

Machine deployments can be imported using their cluster ID and machine deployment ID:

```bash
terraform import kkp_machine_deployment_v2.example cluster-id/machine-deployment-id
```
