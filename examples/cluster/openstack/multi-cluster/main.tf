terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.10"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

locals {
  clusters = var.clusters
}

resource "kkp_cluster_v2" "cluster" {
  for_each    = local.clusters
  name        = each.value.name
  k8s_version = each.value.k8s_version
  datacenter  = each.value.datacenter

  cloud = "openstack"

  openstack {
    application_credential_id     = each.value.openstack.application_credential_id
    application_credential_secret = each.value.openstack.application_credential_secret
    domain                        = each.value.openstack.domain
    network                       = each.value.openstack.network
    subnet_id                     = each.value.openstack.subnet_id
    floating_ip_pool              = each.value.openstack.floating_ip_pool
    security_groups               = each.value.openstack.security_groups
  }
}

output "cluster_ids" {
  value = { for k, c in kkp_cluster_v2.cluster : k => c.id }
}
