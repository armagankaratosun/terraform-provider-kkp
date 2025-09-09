# OpenStack Complete Example

Creates a KKP cluster on OpenStack, a machine deployment, and an SSH key. Optional addon and application resources are included but disabled by default via `enable_*` flags.

Steps:
1. Review `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Toggle `enable_addon` / `enable_application` if your KKP installation provides the required catalogs.
4. Run `terraform init && terraform apply`.

Related:
- [Cluster Examples Index](../README.md)
- [Minimal Cluster](../minimal/README.md)
- [Addon Install](../../addons/README.md)
- [Application Install](../../applications/README.md)
