# Addon Example

Installs a specific addon into an existing cluster. Keep it simple: set `cluster_id` and `addon_name` and apply. To discover available addon names for your cluster, you can use the `kkp_addons_v2` data source (see `examples/data-sources/`).

Steps:
1. Review `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Run `terraform init && terraform apply`.

Related:
- [Examples Index](../README.md)
- [OpenStack Cluster Examples](../cluster/openstack/README.md)
- [Application Install](../applications/README.md)
- [Data Sources](../data-sources/README.md)
