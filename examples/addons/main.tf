terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.11"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

resource "kkp_addon_v2" "addon" {
  cluster_id             = var.cluster_id
  name                   = var.addon_name
  continuously_reconcile = var.continuously_reconcile
  variables              = var.variables_json
  wait_for_ready         = var.wait_for_ready
  timeout_minutes        = var.timeout_minutes
}

output "addon_id" {
  value = kkp_addon_v2.addon.id
}
