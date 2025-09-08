# kkp_ssh_keys_v2

Provides information about SSH keys in a KKP project.

## Example Usage

```terraform
# Get all SSH keys in the project
data "kkp_ssh_keys_v2" "all" {}

# Output SSH key information
output "ssh_key_names" {
  value = [for key in data.kkp_ssh_keys_v2.all.ssh_keys : key.name]
}

# Filter SSH keys by labels
locals {
  production_keys = [
    for key in data.kkp_ssh_keys_v2.all.ssh_keys : key
    if lookup(key.labels, "environment", "") == "production"
  ]
}
```

## Attributes Reference

The following attributes are exported:

- `id` - The data source identifier.
- `ssh_keys` - A list of SSH key objects with the following attributes:
  - `id` - The unique identifier of the SSH key.
  - `name` - The name of the SSH key.
  - `creation_time` - The timestamp when the SSH key was created.
  - `labels` - A map of labels assigned to the SSH key.

**Note:** The public key content is not returned by the API for security reasons.