# OpenStack Multi-Cluster Example

Creates multiple KKP clusters on OpenStack in a single run using `for_each`. Each cluster can provide its own OpenStack application credentials and networking.

Steps:
1. Review `variables.tf` for the `clusters` object schema.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Run `terraform init && terraform apply`.

Related:
- [Cluster Examples Index](../README.md)
- [Minimal Cluster](../minimal/README.md)
- [Template-based Creation](../template/README.md)
