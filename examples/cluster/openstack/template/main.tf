terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.8"
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
  # Required by schema but not used for template creation
  cloud       = "openstack"
  k8s_version = var.placeholder_k8s_version
  datacenter  = var.placeholder_datacenter

  # Required name: must match the template's cluster name to resolve ID
  name = var.cluster_name

  # Template-based creation
  use_template      = true
  template_id       = var.template_id
  template_name     = var.template_name
  template_replicas = var.template_replicas
}

output "cluster_id" {
  value = kkp_cluster_v2.cluster.id
}
