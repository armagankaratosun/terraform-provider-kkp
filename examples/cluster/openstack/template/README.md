# OpenStack Template Example

Creates clusters from a KKP Cluster Template (V2). You can reference the template by `template_id` or by `template_name` (if unique). `template_replicas` controls how many instances are created.

Notes:
- Schema requires `cloud`, `k8s_version`, and `datacenter`, but these are not used for template creation; they are placeholders.
- `name` must match the template's cluster name to resolve the resulting cluster ID after creation.

Steps:
1. Review `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file (you create) or `-var` flags.
3. Run `terraform init && terraform apply`.

Related:
- [Cluster Examples Index](../README.md)
- [Minimal Cluster](../minimal/README.md)
- [Multi-Cluster Example](../multi-cluster/README.md)
