# Terraform Provider for Kubermatic Kubernetes Platform (KKP)

[![Go Report Card](https://goreportcard.com/badge/github.com/armagankaratosun/terraform-provider-kkp)](https://goreportcard.com/report/github.com/armagankaratosun/terraform-provider-kkp)

The Terraform Provider for Kubermatic Kubernetes Platform (KKP) allows you to manage KKP resources using Infrastructure as Code. This provider enables you to create and manage Kubernetes clusters, machine deployments, SSH keys, addons, and applications on various cloud providers through KKP.

## ⚠️ Warning - Here Be Dragons!

This provider is currently in active development. APIs may change and some features may not be fully implemented. Use with caution in production environments.

## Quickstart

1) Configure the provider in `main.tf`:

```hcl
terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1"
    }
  }
}

provider "kkp" {
  # Replace with your KKP API endpoint and credentials
  endpoint             = "https://kkp.example.com"
  token                = "<your-service-account-token>"
  project_id           = "<your-project-id>"
  insecure_skip_verify = false # set true only for dev/self-signed
}
```

2) Minimal OpenStack cluster + workers (adjust inline values as needed):

```hcl
resource "kkp_cluster_v2" "cluster" {
  name        = "quickstart-cluster"     # change if you like
  k8s_version = "1.29.0"                  # use a supported version
  datacenter  = "openstack-eu-west"       # your KKP datacenter name
  cloud       = "openstack"

  openstack {
    application_credential_id     = "<app-credential-id>"     # required when not using presets
    application_credential_secret = "<app-credential-secret>" # required when not using presets
    network                       = "private"                 # neutron network name or ID
    subnet_id                     = "<subnet-id>"             # IPv4 subnet ID
    floating_ip_pool              = "public"                  # external network / FIP pool
    security_groups               = "default"                 # security group name or ID
  }
}

resource "kkp_machine_deployment_v2" "workers" {
  cluster_id = kkp_cluster_v2.cluster.id
  name       = "workers"
  replicas   = 3
  cloud      = "openstack"

  openstack {
    flavor          = "m1.medium"      # pick a valid flavor
    image           = "Ubuntu 22.04"   # pick a valid image name/UUID
    use_floating_ip = true
    disk_size       = 50               # GB
  }
}
```

3) Initialize and apply:

```bash
terraform init
terraform apply
```

For more detailed examples, see the [Examples](#examples) section below.

## Examples

See runnable configurations under [Examples](examples/README.md) for various use cases.

- [OpenStack Cluster Examples](examples/cluster/openstack/README.md)
- [Addon Install](examples/addons/README.md) – Install an addon into an existing cluster.
- [Application Install](examples/applications/README.md) – Install a catalog application into an existing cluster.
- [Data Sources](examples/data-sources/README.md) – Read-only listing of keys, clusters, MDs, addons, applications, templates.


### Discovering Options
- Addons: Use `data "kkp_addons_v2"` with a `cluster_id` to list `available` addon names for that cluster (see `examples/data-sources/`). Then set `addon_name` in `examples/addons/` to install one.
- Applications: Use `data "kkp_applications_v2"` with a `cluster_id` to see what’s installed (see `examples/data-sources/`). Application catalog entries (name/version) come from your KKP setup; pick a valid `application_name`/`application_version` and apply via `examples/applications/`.

## Development

### Overview
Requirements for developing and testing this provider:

- Terraform >= 1.5.0 (to run examples and local testing)
- Go >= 1.24.1 (to build, lint, and run tests)

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `./bin` directory.

### Development Dependencies

Install development dependencies (golangci-lint, etc.):

```shell
$ make dev-deps
```

### Code Formatting and Linting

Format code and run linting:

```shell
$ make fmt
$ make lint
```

### Pre-commit Checks

Run all pre-commit checks (format, lint, test):

```shell
$ make pre-commit
```

### Testing

In order to test the provider, you can simply run `make test`.

```shell
$ make test
```

For test coverage:

```shell
$ make test-coverage
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```shell
$ make testacc
```

### Installing Locally

To install the provider locally for development:

```shell
$ make install
```

This will install the provider to `~/.terraform.d/plugins/registry.opentofu.org/armagankaratosun/kkp/`.

### Branching model

These are the branches used in this repository:

* `main` represents the current release
* `release/v*` (e.g. `release/v0.1`) represents the latest state of a particular release branch. These branches are created when needed from `main`

### Release

This provider is automatically released using GitHub Actions and [GoReleaser](https://goreleaser.com/) when a tag is pushed to the repository.

To release a new version:
1. Update the CHANGELOG.md file with the new version number and changes.
2. Create a Git tag for the release:
   ```
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
3. The release will be automatically built and published to the Terraform Registry.

### Make Targets

| Target | Description |
|--------|-------------|
| `help` | Display available make targets |
| `build` | Build the Terraform provider binary |
| `test` | Run all tests |
| `test-coverage` | Run tests with coverage report |
| `lint` | Run linting with golangci-lint |
| `lint-fix` | Run linting with auto-fix where possible |
| `clean` | Clean build artifacts |
| `install` | Install the provider binary locally |
| `dev-deps` | Install development dependencies |
| `fmt` | Format Go code |
| `tidy` | Tidy up go.mod |
| `check` | Run both linting and tests |
| `pre-commit` | Run pre-commit checks (format, tidy, lint, test) |
| `dev` | Full development build (clean, format, tidy, build) |
| `release` | Release build (clean, test, lint, build) |

## Supported Resources & Cloud Providers

For detailed configuration and usage examples, see the [documentation](docs/).

### Resources

| Resource | Status | Description |
|----------|--------|-------------|
| `kkp_ssh_key_v2` | ✅ **Stable** | SSH key management (cloud-agnostic) |
| `kkp_cluster_v2` | 🚧 **Beta** | Cluster lifecycle with health checking (OpenStack only) |
| `kkp_machine_deployment_v2` | 🚧 **Beta** | Worker node management with autoscaling (OpenStack only) |
| `kkp_addon_v2` | ✅ **Stable** | Cluster addon installation and management |
| `kkp_application_v2` | ✅ **Stable** | Application deployment to clusters |

**Planned Resources:**
- `kkp_role_v2`, `kkp_role_binding_v2`, `kkp_namespace_v2`
- `kkp_constraint_v2`, `kkp_cluster_template`, `kkp_backup_config`

### Data Sources

| Data Source | Status | Description |
|-------------|--------|-------------|
| `kkp_ssh_keys_v2` | ✅ **Stable** | Query existing SSH keys |
| `kkp_clusters_v2` | ✅ **Stable** | Query existing clusters |
| `kkp_machine_deployments_v2` | 🚧 **Beta** | Query machine deployments (cluster-dependent) |
| `kkp_addons_v2` | ✅ **Stable** | Query cluster addons |
| `kkp_applications_v2` | ✅ **Stable** | Query cluster applications |
| `kkp_cluster_templates_v2` | ✅ **Stable** | Query cluster templates |

**Planned Data Sources:**
- `kkp_roles_v2`, `kkp_role_bindings_v2`, `kkp_namespaces_v2`
- `kkp_constraints`, `kkp_projects`, `kkp_external_clusters`

### Cloud Provider Support

Currently **OpenStack only**:
- ✅ Application credential authentication (direct configuration)
- ✅ KKP preset support (for pre-configured credentials)  
- ✅ Network and subnet configuration
- ✅ Security groups and floating IP pools
- ✅ Flavor and image selection for machine deployments
- ✅ Availability zone support

**Status:** 🚧 **Beta** - Fully functional for OpenStack, other cloud providers not yet implemented.

**Future:** Additional cloud providers (AWS, Azure, vSphere, GCP) may be added based on community demand.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on contributing to this provider.

## Support

- [GitHub Issues](https://github.com/armagankaratosun/terraform-provider-kkp/issues)
- [GitHub Discussions](https://github.com/armagankaratosun/terraform-provider-kkp/discussions)

## License

This provider is distributed under the terms of the [Apache License 2.0](LICENSE).
