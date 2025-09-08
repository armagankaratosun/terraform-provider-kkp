# kkp_cluster_v2

Manages Kubernetes clusters in KKP. Supports multiple cloud providers with different configuration options.

## Example Usage

### OpenStack Cluster

```terraform
resource "kkp_cluster_v2" "openstack" {
  name    = "openstack-cluster"
  version = "1.29.0"
  
  cloud = "openstack"
  
  openstack {
    application_credential_id     = var.openstack_app_cred_id
    application_credential_secret = var.openstack_app_cred_secret
    domain                       = "default"
    network                      = "public"
    subnet_id                    = "subnet-12345"
    floating_ip_pool             = "public"
    security_groups              = "default"
  }
  
  labels = {
    environment = "development"
    team        = "platform"
  }
}
```

### Create from Cluster Template (V2)

```terraform
resource "kkp_cluster_v2" "from_template" {
  # Expected name produced by the template
  name = "demo-cluster-from-template"

  # Enable template instantiation path
  use_template      = true
  template_id       = var.cluster_template_id
  template_replicas = 1
}
```

### Create from Cluster Template by Name

```terraform
resource "kkp_cluster_v2" "from_template_name" {
  # Expected name produced by the template
  name = "demo-cluster-from-template"

  use_template      = true
  template_name     = "my-shared-template"
  template_replicas = 1
}
```

### OpenStack Cluster (using preset)

```terraform
resource "kkp_cluster_v2" "openstack_preset" {
  name       = "openstack-preset-cluster"
  version    = "1.29.0"
  datacenter = "openstack-eu-west"
  
  cloud  = "openstack"
  preset = "openstack-production-preset"
  
  # OpenStack block can be omitted when using presets
  # or used to override specific preset values
  openstack {
    network = "custom-network"  # Override preset network
  }
}
```

## Argument Reference

The following arguments are supported:

### General Arguments

- `name` - (Required) The name of the cluster.
- `version` - (Required) The Kubernetes version to use (e.g., "1.29.0").
- `cloud` - (Required) The cloud provider. Currently only `openstack` is supported.
- `preset` - (Optional) KKP preset name for predefined cloud configurations.
- `labels` - (Optional) A map of labels to assign to the cluster.

Template-based creation (optional):
- `use_template` - (Optional) When true with `template_id` set, creates the cluster by instantiating a Cluster Template (V2).
- `template_id` - (Optional) Cluster Template ID to instantiate.
- `template_name` - (Optional) Cluster Template name to instantiate. If multiple templates share the same name, apply will fail and you should use `template_id`.
- `template_replicas` - (Optional) Number of template instances to create. Defaults to `1`.

### OpenStack Configuration

The `openstack` block supports:

- `application_credential_id` - (Optional) OpenStack application credential ID.
- `application_credential_secret` - (Optional) OpenStack application credential secret.
- `domain` - (Optional) OpenStack domain name (e.g., "default").
- `network` - (Optional) OpenStack network name.
- `subnet_id` - (Optional) OpenStack subnet ID.
- `floating_ip_pool` - (Optional) External network name for floating IPs.
- `security_groups` - (Optional) Security group names (comma-separated).
- `use_token` - (Optional) Use token-based authentication. Defaults to `true`.

### Note on Additional Cloud Providers

Currently, only OpenStack is supported. Other cloud providers (AWS, Azure, vSphere, GCP) are not implemented but may be added in future releases based on community demand.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the cluster.
- `creation_time` - The timestamp when the cluster was created.

## Import

Clusters can be imported using their ID:

```bash
terraform import kkp_cluster_v2.example cluster-id
```
