variable "kkp_endpoint" {
  description = "KKP API endpoint (with or without /api)"
  type        = string
}

variable "kkp_token" {
  description = "KKP bearer token for API"
  type        = string
  sensitive   = true
}

variable "kkp_project_id" {
  description = "KKP project ID"
  type        = string
}

variable "kkp_insecure_skip_verify" {
  description = "Optional: skip TLS verification (dev only)"
  type        = bool
  default     = false
}

variable "cluster_name" {
  description = "Cluster name"
  type        = string
}

variable "cluster_k8s_version" {
  description = "Kubernetes version (e.g. 1.29.0)"
  type        = string
}

variable "cluster_datacenter" {
  description = "KKP datacenter name (e.g. openstack-eu-west)"
  type        = string
}

variable "os_app_cred_id" {
  description = "OpenStack application credential ID"
  type        = string
  sensitive   = true
}

variable "os_app_cred_secret" {
  description = "OpenStack application credential Secret"
  type        = string
  sensitive   = true
}


variable "os_network" {
  description = "Neutron network name or ID"
  type        = string
}

variable "os_subnet_id" {
  description = "IPv4 subnet ID"
  type        = string
}

variable "os_floating_ip_pool" {
  description = "External network / Floating IP pool (e.g. public)"
  type        = string
}

variable "os_security_groups" {
  description = "Security group name or ID"
  type        = string
}
