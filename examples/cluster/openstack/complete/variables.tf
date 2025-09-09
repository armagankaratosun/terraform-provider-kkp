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

variable "enable_ssh_key" {
  description = "Optional: create SSH key resource"
  type        = bool
  default     = false
}
variable "ssh_key_name" {
  description = "Optional: SSH key name (required if enable_ssh_key = true)"
  type        = string
  default     = "example-key"
}
variable "ssh_public_key" {
  description = "Optional: OpenSSH public key (required if enable_ssh_key = true)"
  type        = string
  default     = ""
  sensitive   = true
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

variable "os_app_cred_id" {
  description = "OpenStack application credential ID"
  type        = string
  sensitive   = true
}
variable "os_app_cred_secret" {
  description = "OpenStack application credential secret"
  type        = string
  sensitive   = true
}
variable "os_network" {
  description = "OpenStack network name or ID"
  type        = string
}
variable "os_subnet_id" {
  description = "OpenStack subnet ID"
  type        = string
}
variable "os_floating_ip_pool" {
  description = "OpenStack external network / FIP pool"
  type        = string
}
variable "os_security_groups" {
  description = "OpenStack security group"
  type        = string
}

variable "md_name" {
  description = "Optional: machine deployment name"
  type        = string
  default     = "workers"
}
variable "md_replicas" {
  description = "Optional: worker replicas"
  type        = number
  default     = 3
}
variable "md_os_flavor" {
  description = "OpenStack flavor"
  type        = string
}
variable "md_os_image" {
  description = "OpenStack image"
  type        = string
}
variable "md_os_use_fip" {
  description = "Optional: assign floating IPs to workers"
  type        = bool
  default     = true
}
variable "md_os_disk_size" {
  description = "Optional: root disk size (GB)"
  type        = number
  default     = 50
}

variable "enable_addon" {
  description = "Optional: create addon example"
  type        = bool
  default     = false
}
variable "addon_name" {
  description = "Optional: addon name (e.g., logging)"
  type        = string
  default     = "logging"
}
variable "addon_continuously_reconcile" {
  description = "Optional: keep addon reconciled"
  type        = bool
  default     = false
}
variable "addon_variables_json" {
  description = "Optional: addon variables JSON string"
  type        = string
  default     = null
}
variable "addon_wait_for_ready" {
  description = "Optional: wait for addon ready"
  type        = bool
  default     = true
}
variable "addon_timeout_minutes" {
  description = "Optional: addon wait timeout (minutes)"
  type        = number
  default     = 20
}

variable "enable_application" {
  description = "Optional: create application example"
  type        = bool
  default     = false
}
variable "app_install_name" {
  description = "Optional: application installation name"
  type        = string
  default     = "example-app"
}
variable "app_namespace" {
  description = "Optional: application target namespace"
  type        = string
  default     = "kube-system"
}
variable "app_catalog_name" {
  description = "Optional: catalog application name"
  type        = string
  default     = "metrics-server"
}
variable "app_version" {
  description = "Optional: catalog version (semantic or channel)"
  type        = string
  default     = "latest"
}
variable "app_values_yaml" {
  description = "Optional: YAML values for application"
  type        = string
  default     = null
}
variable "app_wait_for_ready" {
  description = "Optional: wait for application ready"
  type        = bool
  default     = true
}
variable "app_timeout_minutes" {
  description = "Optional: application wait timeout (minutes)"
  type        = number
  default     = 20
}
