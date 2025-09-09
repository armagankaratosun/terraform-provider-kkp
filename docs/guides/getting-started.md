# Getting Started with the KKP Provider

This guide will walk you through setting up and using the Terraform KKP provider to manage your Kubermatic Kubernetes Platform resources.

## Prerequisites

- Terraform >= 1.5.0
- Access to a KKP instance
- A KKP service account with appropriate permissions

## Step 1: Set up Authentication

### Create a Service Account

1. Log into your KKP dashboard
2. Navigate to your project settings
3. Go to "Service Accounts"
4. Create a new service account
5. Assign the "Editor" role for full resource management
6. Generate a token for the service account

### Configure Environment Variables

```bash
export KKP_ENDPOINT="https://your-kkp-instance.com"
export KKP_TOKEN="your-service-account-token"
export KKP_PROJECT_ID="your-project-id"
```

## Step 2: Configure the Provider

Create a `main.tf` file:

```terraform
terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.0.1"
    }
  }
}

provider "kkp" {
  endpoint   = var.kkp_endpoint
  token      = var.kkp_token
  project_id = var.kkp_project_id
}
```

Create a `variables.tf` file:

```terraform
variable "kkp_endpoint" {
  description = "KKP API endpoint"
  type        = string
}

variable "kkp_token" {
  description = "KKP service account token"
  type        = string
  sensitive   = true
}

variable "kkp_project_id" {
  description = "KKP project ID"
  type        = string
}
```

## Step 3: Create Your First Resources

### SSH Key

```terraform
resource "kkp_ssh_key_v2" "example" {
  name       = "my-terraform-key"
  public_key = file("~/.ssh/id_rsa.pub")
  
  labels = {
    managed-by = "terraform"
  }
}
```

### Cluster (OpenStack)

```terraform
resource "kkp_cluster_v2" "example" {
  name    = "my-first-cluster"
  version = "1.29.0"
  
  cloud = "openstack"
  
  openstack {
    application_credential_id     = var.openstack_app_cred_id
    application_credential_secret = var.openstack_app_cred_secret
    domain                       = "default"
    network                      = "public"
    subnet_id                    = var.openstack_subnet_id
    floating_ip_pool             = "public"
    security_groups              = "default"
  }
  
  labels = {
    environment = "development"
    managed-by  = "terraform"
  }
}
```

### Machine Deployment

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
  }
}
```

## Step 4: Apply Your Configuration

1. Initialize Terraform:
   ```bash
   terraform init
   ```

2. Plan your changes:
   ```bash
   terraform plan
   ```

3. Apply the configuration:
   ```bash
   terraform apply
   ```

## Step 5: Query Resources with Data Sources

```terraform
# Get information about all clusters
data "kkp_clusters_v2" "all" {}

# Get information about all SSH keys
data "kkp_ssh_keys_v2" "all" {}

# Output cluster information
output "clusters" {
  value = {
    for cluster in data.kkp_clusters_v2.all.clusters : cluster.name => {
      id      = cluster.id
      version = cluster.version
      cloud   = cluster.cloud
    }
  }
}
```

## Next Steps

- Explore the [resource documentation](../resources/) for detailed configuration options
- Check out the [data source documentation](../data-sources/) for querying existing resources
- Review best practices for managing KKP resources with Terraform

## Troubleshooting

### Common Issues

1. **Authentication errors**: Verify your service account has the correct permissions
2. **Network timeouts**: Check if your KKP endpoint is accessible
3. **Resource conflicts**: Ensure resource names are unique within the project

### Debug Mode

Enable debug logging:

```bash
export TF_LOG=DEBUG
terraform apply
```
