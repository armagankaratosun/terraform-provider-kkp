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

variable "cluster_id" {
  description = "Target cluster ID"
  type        = string
}

variable "install_name" {
  description = "Application installation name"
  type        = string
}
variable "namespace" {
  description = "Optional: target namespace"
  type        = string
  default     = "kube-system"
}
variable "application_name" {
  description = "Catalog application name"
  type        = string
}
variable "application_version" {
  description = "Optional: catalog application version (semantic or channel)"
  type        = string
  default     = "latest"
}
variable "values_yaml" {
  description = "Optional: YAML values string for application"
  type        = string
  default     = null
}
variable "wait_for_ready" {
  description = "Optional: wait for application to be ready (default true)"
  type        = bool
  default     = true
}
variable "timeout_minutes" {
  description = "Optional: timeout in minutes while waiting (default 20)"
  type        = number
  default     = 20
}
