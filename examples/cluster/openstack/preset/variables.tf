variable "kkp_endpoint" {
  description = "KKP API endpoint"
  type        = string
}
variable "kkp_token" {
  description = "KKP bearer token"
  type        = string
  sensitive   = true
}
variable "kkp_project_id" {
  description = "KKP project ID"
  type        = string
}
variable "kkp_insecure_skip_verify" {
  description = "Optional: skip TLS verification"
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
  description = "KKP datacenter (e.g. openstack-eu-west)"
  type        = string
}

variable "kkp_preset" {
  description = "KKP preset name with OpenStack credentials"
  type        = string
}

variable "os_domain" {
  description = "Optional: OpenStack domain (if required by your environment)"
  type        = string
}
variable "os_network" {
  description = "Optional: Neutron network name or ID"
  type        = string
}
variable "os_subnet_id" {
  description = "Optional: IPv4 subnet ID"
  type        = string
}
variable "os_floating_ip_pool" {
  description = "Optional: external network / Floating IP pool"
  type        = string
}
variable "os_security_groups" {
  description = "Optional: security group"
  type        = string
}
