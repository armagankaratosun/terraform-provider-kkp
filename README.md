# Terraform Provider for Kubermatic Kubernetes Platform (KKP)

[![CI](https://github.com/armagankaratosun/terraform-provider-kkp/actions/workflows/ci.yml/badge.svg)](https://github.com/armagankaratosun/terraform-provider-kkp/actions/workflows/ci.yml)
[![Release](https://github.com/armagankaratosun/terraform-provider-kkp/actions/workflows/release.yml/badge.svg)](https://github.com/armagankaratosun/terraform-provider-kkp/actions/workflows/release.yml)
[![Latest Version](https://img.shields.io/github/v/release/armagankaratosun/terraform-provider-kkp?sort=semver)](https://github.com/armagankaratosun/terraform-provider-kkp/releases)
[![KKP Compatibility](https://img.shields.io/badge/KKP-2.28.2%20CE-blue)](https://github.com/kubermatic/kubermatic/releases/tag/v2.28.2)
[![Go Report Card](https://goreportcard.com/badge/github.com/armagankaratosun/terraform-provider-kkp)](https://goreportcard.com/report/github.com/armagankaratosun/terraform-provider-kkp)

The Terraform Provider for Kubermatic Kubernetes Platform (KKP) allows you to manage KKP resources using Infrastructure as Code. This provider enables you to create and manage Kubernetes clusters, machine deployments, SSH keys, addons, and applications on various cloud providers through KKP.

## ⚠️ Warning - Here Be Dragons!

This provider is currently in active development. APIs may change and some features may not be fully implemented. Use with caution in production environments.

## Compatibility

The current provider version is tested with Kubermatic Kubernetes Platform (KKP) v2.28.2 — Community Edition.
Other KKP versions may work but are not validated. See COMPATIBILITY.md for a simple mapping of provider ↔ KKP versions.

## Quickstart

1) Configure the provider in `main.tf`:

```hcl
terraform {
  required_providers {
    kkp = {
      source  = "armagankaratosun/kkp"
      version = "~> 0.1.8" # pin to a minor series
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

#### Addons
- Purpose: discover installable addons for a given cluster.
- How: query the data source, then pick a name from `available`.

```hcl
data "kkp_addons_v2" "this" {
  cluster_id = "<cluster-id>"
}

output "available_addons" {
  value = data.kkp_addons_v2.this.available
}
```

Next: choose an addon and set `addon_name` in the [addon example](examples/addons/README.md).

#### Applications
- Purpose: list applications currently installed in a cluster.
- How: query the data source; catalog names/versions come from your KKP setup.

```hcl
data "kkp_applications_v2" "this" {
  cluster_id = "<cluster-id>"
}

output "installed_applications" {
  value = data.kkp_applications_v2.this.applications
}
```

Next: pick a valid `application_name` and `application_version`, then apply the [application example](examples/applications/README.md).

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

This will install the provider to `~/.terraform.d/plugins/registry.terraform.io/armagankaratosun/kkp/`.

### Branching model

These are the branches used in this repository:

* `main` represents the current release
* `release/v*` (e.g. `release/v0.0.1`) represents the latest state of a particular release branch. These branches are created when needed from `main`

### Release

Releases are built automatically via GitHub Actions and [GoReleaser](https://goreleaser.com/) when a semver tag is pushed.

To cut a release:
1. Tag and push:
   ```bash
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
2. A GitHub Release will be created with binaries, checksums, and autogenerated release notes.

Note: Publishing to the Terraform/OpenTofu registries will follow Terraform Registry best practices. Provider documentation in `docs/` is autogenerated with tfplugindocs.

### Documentation

Provider and resource documentation under `docs/` is generated using HashiCorp's terraform-plugin-docs (tfplugindocs) and is required for publishing on the Terraform Registry.

- Generate docs locally: `make docs`
- Verify docs are up-to-date: `make docs-check`
- Do not edit files under `docs/` by hand; changes will be overwritten by generation.

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

See the docs for the full, up‑to‑date lists and details:

- Resources: [Supported Resources](docs/index.md#supported-resources) and per‑resource docs under [docs/resources/](docs/resources/)
- Data sources: [Data Sources](docs/index.md#data-sources) and per‑data‑source docs under [docs/data-sources/](docs/data-sources/)
- Cloud provider support: [Cloud Provider Support](docs/index.md#cloud-provider-support)

Note: Currently OpenStack is supported (Beta). Other providers may follow.

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on contributing to this provider.

## Support

- [GitHub Issues](https://github.com/armagankaratosun/terraform-provider-kkp/issues)
- [GitHub Discussions](https://github.com/armagankaratosun/terraform-provider-kkp/discussions)

## License

This provider is distributed under the terms of the [Apache License 2.0](LICENSE).
