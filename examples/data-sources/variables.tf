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

variable "cluster_id_for_mds" {
  description = "Optional: cluster ID to fetch machine deployments/addons/applications (required for those queries)"
  type        = string
  default     = ""
}
