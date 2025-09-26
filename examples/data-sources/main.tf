terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.7"
    }
  }
}

provider "kkp" {
  endpoint             = var.kkp_endpoint
  token                = var.kkp_token
  project_id           = var.kkp_project_id
  insecure_skip_verify = var.kkp_insecure_skip_verify
}

data "kkp_ssh_keys_v2" "all" {}
data "kkp_clusters_v2" "all" {}
data "kkp_machine_deployments_v2" "by_cluster" {
  # Provide a cluster ID to fetch deployments
  cluster_id = var.cluster_id_for_mds
}
data "kkp_addons_v2" "by_cluster" {
  cluster_id = var.cluster_id_for_mds
}
data "kkp_applications_v2" "by_cluster" {
  cluster_id = var.cluster_id_for_mds
}
data "kkp_cluster_templates_v2" "all" {}

# Fetch kubeconfig for an existing cluster (provide cluster_id_for_kubeconfig)
data "kkp_cluster_kubeconfig_v2" "kube" {
  cluster_id = var.cluster_id_for_kubeconfig
}

output "ssh_keys" {
  value = data.kkp_ssh_keys_v2.all.ssh_keys
}

output "cluster_count" {
  value = length(data.kkp_clusters_v2.all.clusters)
}

output "machine_deployments" {
  value = data.kkp_machine_deployments_v2.by_cluster.machine_deployments
}

# Preview first 120 characters of the kubeconfig (to avoid dumping secrets)
output "kubeconfig_preview" {
  value = var.cluster_id_for_kubeconfig == "" ? null : substr(nonsensitive(data.kkp_cluster_kubeconfig_v2.kube.content), 0, 120)
}
