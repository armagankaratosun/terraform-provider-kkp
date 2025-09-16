terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.3"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

resource "kkp_cluster_v2" "cluster" {
  name        = var.cluster_name
  k8s_version = var.cluster_k8s_version
  datacenter  = var.cluster_datacenter

  cloud = "openstack"

  # When not using presets, configure application credentials directly
  openstack {
    application_credential_id     = var.os_app_cred_id
    application_credential_secret = var.os_app_cred_secret
    network                       = var.os_network
    subnet_id                     = var.os_subnet_id
    floating_ip_pool              = var.os_floating_ip_pool
    security_groups               = var.os_security_groups
  }
}

output "cluster_id" {
  value = kkp_cluster_v2.cluster.id
}
