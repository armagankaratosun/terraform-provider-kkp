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
  description = "Skip TLS verification"
  type        = bool
  default     = false
}

variable "cluster_id" {
  description = "Target cluster ID"
  type        = string
}

variable "addon_name" {
  description = "Addon name (e.g., logging, prometheus)"
  type        = string
}
variable "continuously_reconcile" {
  description = "Optional: keep addon reconciled"
  type        = bool
  default     = false
}
variable "variables_json" {
  description = "Optional: JSON string with values for addon templates"
  type        = string
  default     = null
}
variable "wait_for_ready" {
  description = "Optional: wait for addon to be ready (default true)"
  type        = bool
  default     = true
}
variable "timeout_minutes" {
  description = "Optional: timeout in minutes while waiting (default 20)"
  type        = number
  default     = 20
}
