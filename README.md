# Terraform Provider for Kubermatic Kubernetes Platform (KKP)

[![Go Report Card](https://goreportcard.com/badge/github.com/armagankaratosun/terraform-provider-kkp)](https://goreportcard.com/report/github.com/armagankaratosun/terraform-provider-kkp)

The Terraform Provider for Kubermatic Kubernetes Platform (KKP) allows you to manage KKP resources using Infrastructure as Code. This provider enables you to create and manage Kubernetes clusters, machine deployments, SSH keys, addons, and applications on various cloud providers through KKP.

## ⚠️ Warning - Here Be Dragons!

This provider is currently in active development. APIs may change and some features may not be fully implemented. Use with caution in production environments.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.5.0
- [Go](https://golang.org/doc/install) >= 1.24.1

## Building The Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using the Make command:

```shell
make build
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```shell
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

### Provider Configuration

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
  endpoint           = "https://your-kkp.example.com"
  token             = "your-api-token"
  project_id        = "your-project-id"
  insecure_skip_verify = false
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `KKP_ENDPOINT` | KKP API endpoint URL |
| `KKP_TOKEN` | KKP API token |
| `KKP_PROJECT_ID` | Default project ID |
| `KKP_INSECURE_SKIP_VERIFY` | Skip TLS verification (default: false) |

## Developing the Provider

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

## Testing the Provider

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

## Installing the Provider Locally

To install the provider locally for development:

```shell
$ make install
```

This will install the provider to `~/.terraform.d/plugins/registry.opentofu.org/armagankaratosun/kkp/`.

## Branching model

These are the branches used in this repository:

* `main` represents the current release
* `release/v*` (e.g. `release/v0.1`) represents the latest state of a particular release branch. These branches are created when needed from `main`

## Release

This provider is automatically released using GitHub Actions and [GoReleaser](https://goreleaser.com/) when a tag is pushed to the repository.

To release a new version:
1. Update the CHANGELOG.md file with the new version number and changes.
2. Create a Git tag for the release:
   ```
   git tag vX.Y.Z
   git push origin vX.Y.Z
   ```
3. The release will be automatically built and published to the Terraform Registry.

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

## Example Usage

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
  endpoint   = "https://kkp.example.com"
  token      = var.kkp_token
  project_id = var.project_id
}

resource "kkp_cluster_v2" "example" {
  name       = "example-cluster"
  version    = "1.29.0"
  datacenter = "openstack-eu-west"
  
  cloud = "openstack"
  
  openstack {
    application_credential_id     = var.openstack_app_cred_id
    application_credential_secret = var.openstack_app_cred_secret
    domain                       = "default"
    network                      = "public"
    subnet_id                    = var.openstack_subnet_id
    floating_ip_pool             = "public"
    security_groups              = "default"
  }
  
  labels = {
    environment = "test"
  }
}

resource "kkp_machine_deployment_v2" "workers" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "worker-nodes"
  replicas   = 3
  
  cloud = "openstack"
  
  openstack {
    flavor             = "m1.medium"
    image              = "Ubuntu 22.04"
    use_floating_ip    = true
    disk_size          = 50
    availability_zone  = "nova"
  }
}
```

## Make Targets

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

## Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on contributing to this provider.

## Support

- [GitHub Issues](https://github.com/armagankaratosun/terraform-provider-kkp/issues)
- [GitHub Discussions](https://github.com/armagankaratosun/terraform-provider-kkp/discussions)

## License

This provider is distributed under the terms of the [Apache License 2.0](LICENSE).
