# Application Example

Installs a specific application from the KKP catalog into an existing cluster. Keep it simple: set `cluster_id`, `application_name`, and `application_version`. To inspect what is already installed, use the `kkp_applications_v2` data source (see `examples/data-sources/`).

Steps:
1. Review `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Run `terraform init && terraform apply`.

Related:
- [Examples Index](../README.md)
- [OpenStack Cluster Examples](../cluster/openstack/README.md)
- [Addon Install](../addons/README.md)
- [Data Sources](../data-sources/README.md)
