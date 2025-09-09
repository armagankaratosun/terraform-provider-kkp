# OpenStack Preset Example

Creates a KKP cluster on OpenStack using a KKP preset (credential stored in KKP). This flow typically uses token-based auth from the preset.

Steps:
1. Review `variables.tf` for required inputs.
2. Ensure the preset exists in KKP and has OpenStack credentials.
3. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
4. Run `terraform init && terraform apply`.

Related:
- [Cluster Examples Index](../README.md)
- [Minimal Cluster](../minimal/README.md)
- [Complete Setup (with workers)](../complete/README.md)
