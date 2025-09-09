# OpenStack Examples

This directory groups all OpenStack-focused examples for the KKP provider. Jump straight into each tutorial:

- [Minimal](examples/cluster/openstack/minimal/README.md): Single cluster using application credentials.
- [Complete](examples/cluster/openstack/complete/README.md): Cluster + machine deployment, optional addon and application.
- [Preset](examples/cluster/openstack/preset/README.md): Cluster using a KKP preset (token-based auth from preset).
- [From Template](examples/cluster/openstack/template/README.md): Cluster created from a Cluster Template (by ID or unique name), supports replicas.
- [Multi-Cluster](examples/cluster/openstack/multi-cluster/README.md): Create multiple clusters in one apply using `for_each` and per-cluster app credentials.

General steps:
1. Review each example's `variables.tf` for required inputs.
2. Provide values via a `terraform.tfvars` file you create or `-var` flags.
3. Run `terraform init && terraform apply`.

See also:
- [Examples Index](examples/README.md)
- [Root README](../../README.md)
- [Addon Example](../addons/README.md)
- [Application Example](../applications/README.md)
- [Data Sources](../data-sources/README.md)
