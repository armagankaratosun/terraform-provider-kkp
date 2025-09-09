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

variable "clusters" {
  description = "Map of clusters to create"
  type = map(object({
    name        = string
    k8s_version = string
    datacenter  = string
    openstack = object({
      application_credential_id     = string
      application_credential_secret = string
      domain                        = string
      network                       = string
      subnet_id                     = string
      floating_ip_pool              = string
      security_groups               = string
    })
  }))
}
