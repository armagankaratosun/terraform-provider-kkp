terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.4"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

# Optional: create a project-scoped SSH key
resource "kkp_ssh_key_v2" "example" {
  count      = var.enable_ssh_key ? 1 : 0
  name       = var.ssh_key_name
  public_key = var.ssh_public_key
}

resource "kkp_cluster_v2" "cluster" {
  name        = var.cluster_name
  k8s_version = var.cluster_k8s_version
  datacenter  = var.cluster_datacenter

  cloud = "openstack"

  openstack {
    application_credential_id     = var.os_app_cred_id
    application_credential_secret = var.os_app_cred_secret

    network          = var.os_network
    subnet_id        = var.os_subnet_id
    floating_ip_pool = var.os_floating_ip_pool
    security_groups  = var.os_security_groups
  }

}

resource "kkp_machine_deployment_v2" "workers" {
  cluster_id = kkp_cluster_v2.cluster.id
  depends_on = [kkp_cluster_v2.cluster]
  name       = var.md_name
  replicas   = var.md_replicas
  cloud      = "openstack"

  openstack {
    flavor          = var.md_os_flavor
    image           = var.md_os_image
    use_floating_ip = var.md_os_use_fip
    disk_size       = var.md_os_disk_size
  }
}

# Optional addon example (disabled by default)
resource "kkp_addon_v2" "logging" {
  count                  = var.enable_addon ? 1 : 0
  cluster_id             = kkp_cluster_v2.cluster.id
  name                   = var.addon_name
  continuously_reconcile = var.addon_continuously_reconcile
  variables              = var.addon_variables_json
  wait_for_ready         = var.addon_wait_for_ready
  timeout_minutes        = var.addon_timeout_minutes
}

# Optional application example (disabled by default)
resource "kkp_application_v2" "app" {
  count               = var.enable_application ? 1 : 0
  cluster_id          = kkp_cluster_v2.cluster.id
  name                = var.app_install_name
  namespace           = var.app_namespace
  application_name    = var.app_catalog_name
  application_version = var.app_version
  values              = var.app_values_yaml
  wait_for_ready      = var.app_wait_for_ready
  timeout_minutes     = var.app_timeout_minutes
}

output "cluster_id" {
  value = kkp_cluster_v2.cluster.id
}

output "machine_deployment_id" {
  value = kkp_machine_deployment_v2.workers.id
}
