# Examples for terraform-provider-kkp

This directory contains runnable examples that showcase how to use the KKP Terraform provider. Each example defines inputs in `variables.tf` with clear descriptions. Provide values via a `terraform.tfvars` file you create, `-var` flags, or environment variables, then run:

```
terraform init
terraform plan
terraform apply
```

Notes:
- Provider type is `kkp` and resources are prefixed with `kkp_...`.
- OpenStack examples assume you will provide either KKP preset credentials or OpenStack application credentials. These examples use application credentials.
- Some resources (e.g., addons/applications) may require that corresponding catalogs are available in your KKP installation. They are disabled by default using `count` flags you can toggle.
- SSH key resource type in this codebase is `kkp_project_sshkey`.

Examples:
- [OpenStack Minimal](examples/cluster/openstack/minimal/README.md) – Minimal cluster creation on OpenStack.
- [OpenStack Complete](examples/cluster/openstack/complete/README.md) – Cluster + machine deployment (+ optional addon and application).
- [OpenStack Preset](examples/cluster/openstack/preset/README.md) – Cluster creation using a KKP preset for OpenStack credentials.
- [OpenStack Template](examples/cluster/openstack/template/README.md) – Cluster creation from a Cluster Template (supports replicas).
- [OpenStack Multi-Cluster](examples/cluster/openstack/multi-cluster/README.md) – Create multiple clusters via `for_each` with per-cluster app credentials.
- [Data Sources](examples/data-sources/README.md) – Read-only examples listing existing resources.

See also:
- [OpenStack Cluster Examples Index](examples/cluster/openstack/README.md)
- [Addon Install](examples/addons/README.md)
- [Application Install](examples/applications/README.md)
- [Provider Overview](../README.md)
