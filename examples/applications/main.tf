terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

resource "kkp_application_v2" "application" {
  cluster_id          = var.cluster_id
  name                = var.install_name
  namespace           = var.namespace
  application_name    = var.application_name
  application_version = var.application_version
  values              = var.values_yaml
  wait_for_ready      = var.wait_for_ready
  timeout_minutes     = var.timeout_minutes
}

output "application_installation_id" {
  value = kkp_application_v2.application.id
}
