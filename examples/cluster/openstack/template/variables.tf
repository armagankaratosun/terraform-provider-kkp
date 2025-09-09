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
  description = "Template cluster name (must match)"
  type        = string
}

variable "template_id" {
  description = "Optional: cluster template ID to instantiate (preferred)"
  type        = string
  default     = ""
}
variable "template_name" {
  description = "Optional: cluster template name (must be unique if used)"
  type        = string
  default     = ""
}
variable "template_replicas" {
  description = "Optional: number of template instances"
  type        = number
  default     = 1
}

# Placeholders for schema; not used in template flow
variable "placeholder_k8s_version" {
  description = "Optional: placeholder Kubernetes version (schema requirement only)"
  type        = string
  default     = "1.29.0"
}
variable "placeholder_datacenter" {
  description = "Optional: placeholder KKP datacenter (schema requirement only)"
  type        = string
  default     = "openstack-eu-west"
}
