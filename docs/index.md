# KKP Provider

The Kubermatic Kubernetes Platform (KKP) provider is used to interact with KKP resources. The provider needs to be configured with the proper credentials before it can be used.

## Example Usage

```terraform
terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.0.1"
    }
  }
}

# Configure the KKP Provider
provider "kkp" {
  endpoint   = "https://kkp.example.com"
  token      = var.kkp_token
  project_id = var.project_id
}

# Create an SSH key
resource "kkp_ssh_key_v2" "example" {
  name       = "example-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# Create a cluster
resource "kkp_cluster_v2" "example" {
  name    = "example-cluster"
  version = "1.29.0"
  
  cloud = "openstack"
  
  openstack {
    application_credential_id     = var.openstack_app_cred_id
    application_credential_secret = var.openstack_app_cred_secret
    domain                       = "default"
    network                      = "public"
    subnet_id                    = "subnet-123"
    floating_ip_pool             = "public"
    security_groups              = "default"
  }
}
```

## Argument Reference

The following arguments are supported in the `provider` block:

- `endpoint` - (Required) The KKP API endpoint URL. Can also be set via the `KKP_ENDPOINT` environment variable.

- `token` - (Required) Bearer token for KKP API authentication. Should be a service account token with appropriate permissions. Can also be set via the `KKP_TOKEN` environment variable.

- `project_id` - (Required) The default project ID for resources. Can also be set via the `KKP_PROJECT_ID` environment variable.

- `insecure_skip_verify` - (Optional) Skip TLS certificate verification. Defaults to `false`. Can also be set via the `KKP_INSECURE_SKIP_VERIFY` environment variable.

## Environment Variables

The provider can be configured using environment variables:

```bash
export KKP_ENDPOINT="https://kkp.example.com"
export KKP_TOKEN="your-service-account-token"
export KKP_PROJECT_ID="your-project-id"
export KKP_INSECURE_SKIP_VERIFY="false"
```

## Authentication

The KKP provider uses bearer token authentication. You'll need to:

1. Create a service account in your KKP project
2. Assign appropriate roles (typically "Editor" for resource management)
3. Generate a token for the service account
4. Use the token in the provider configuration

## Cloud Provider Support

The provider currently supports OpenStack as the primary cloud provider:

### OpenStack (Beta)
- **Application Credentials**: Direct configuration with application credential ID/secret
- **KKP Presets**: Use pre-configured credentials stored in KKP
- **Network Configuration**: Full control over networks, subnets, security groups
- **Machine Configuration**: Flavor and image selection, availability zones
- **Floating IPs**: Configurable floating IP assignment

**Note**: Other cloud providers (AWS, Azure, vSphere, GCP) are not currently supported but may be added based on community demand and KKP API capabilities.

## Supported Resources

| Resource | Status | Description |
|----------|--------|-------------|
| `kkp_ssh_key_v2` | ✅ Stable | SSH key management (cloud-agnostic) |
| `kkp_cluster_v2` | 🚧 Beta | Cluster lifecycle with health checking (OpenStack only) |
| `kkp_machine_deployment_v2` | 🚧 Beta | Worker node management with autoscaling (OpenStack only) |
| `kkp_addon_v2` | ✅ Stable | Cluster addon installation and management |
| `kkp_application_v2` | ✅ Stable | Application deployment to clusters |

Planned: `kkp_role_v2`, `kkp_role_binding_v2`, `kkp_namespace_v2`, `kkp_constraint_v2`, `kkp_cluster_template`, `kkp_backup_config`

## Data Sources

| Data Source | Status | Description |
|-------------|--------|-------------|
| `kkp_ssh_keys_v2` | ✅ Stable | Query existing SSH keys |
| `kkp_clusters_v2` | ✅ Stable | Query existing clusters |
| `kkp_machine_deployments_v2` | 🚧 Beta | Query machine deployments (cluster-dependent) |
| `kkp_addons_v2` | ✅ Stable | Query cluster addons |
| `kkp_applications_v2` | ✅ Stable | Query cluster applications |
| `kkp_cluster_templates_v2` | ✅ Stable | Query cluster templates |

Planned: `kkp_roles_v2`, `kkp_role_bindings_v2`, `kkp_namespaces_v2`, `kkp_constraints`, `kkp_projects`, `kkp_external_clusters`
