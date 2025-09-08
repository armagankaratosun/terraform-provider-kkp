---
page_title: "kkp_cluster_templates_v2 (Data Source)"
subcategory: "KKP V2"
description: |-
  Retrieve Cluster Templates for a project via the KKP V2 API.
---

# kkp_cluster_templates_v2

Retrieves cluster templates for the configured project using the KKP V2 API.

## Example Usage

```hcl
data "kkp_cluster_templates_v2" "all" {}

output "template_ids" {
  value = [for t in data.kkp_cluster_templates_v2.all.templates : t.id]
}
```

## Attributes Reference

- `id`: Data source identifier.
- `templates`: List of templates with the following attributes:
  - `id`: Cluster template ID.
  - `name`: Cluster template name.
  - `scope`: Template scope (e.g., `project`, `global`).

## Notes

- The data source uses the provider-level `project_id` and authenticates against the V2 endpoints.

