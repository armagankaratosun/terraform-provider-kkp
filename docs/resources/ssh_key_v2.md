# kkp_ssh_key_v2

Manages SSH keys in a KKP project. SSH keys can be used to access cluster nodes and are required for some cluster operations.

## Example Usage

```terraform
# Basic SSH key
resource "kkp_ssh_key_v2" "example" {
  name       = "my-ssh-key"
  public_key = file("~/.ssh/id_rsa.pub")
}

# SSH key with labels
resource "kkp_ssh_key_v2" "labeled" {
  name       = "production-key"
  public_key = file("~/.ssh/prod_rsa.pub")
  
  labels = {
    environment = "production"
    team        = "platform"
  }
}
```

## Argument Reference

The following arguments are supported:

- `name` - (Required) The name of the SSH key. Must be unique within the project.
- `public_key` - (Required) The SSH public key content. Typically read from a file using `file()`.
- `labels` - (Optional) A map of labels to assign to the SSH key.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the SSH key.
- `creation_time` - The timestamp when the SSH key was created.

## Import

SSH keys can be imported using their ID:

```bash
terraform import kkp_ssh_key_v2.example ssh-key-id
```