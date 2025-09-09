# OpenStack Minimal Example

Creates a minimal KKP cluster on OpenStack using application credentials.

Steps:
1. Review `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Run `terraform init && terraform apply`.

Related:
- [Cluster Examples Index](../README.md)
- [Complete Setup (with workers)](../complete/README.md)
- [Preset-based Cluster](../preset/README.md)
- [From Cluster Template](../template/README.md)
- [Multiple Clusters with for_each](../multi-cluster/README.md)
