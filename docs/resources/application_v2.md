# kkp_application_v2

Manages applications deployed to KKP clusters. Applications are Helm charts or other packaged software that can be deployed and managed through the KKP platform.

## Example Usage

### Basic Application Deployment

```terraform
resource "kkp_application_v2" "redis" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "redis"
  namespace  = "database"
  
  spec {
    source {
      helm {
        chart_name    = "redis"
        chart_version = "17.3.7"
        repository {
          name = "bitnami"
          url  = "https://charts.bitnami.com/bitnami"
        }
      }
    }
    
    values = {
      "auth.enabled"        = "true"
      "auth.password"       = var.redis_password
      "master.persistence.enabled" = "true"
      "master.persistence.size"    = "8Gi"
    }
  }
}
```

### Application with Wait for Ready

```terraform
resource "kkp_application_v2" "postgresql" {
  cluster_id      = kkp_cluster_v2.example.id
  name           = "postgresql"
  namespace      = "database"
  wait_for_ready = true
  
  spec {
    source {
      helm {
        chart_name    = "postgresql"
        chart_version = "12.1.2"
        repository {
          name = "bitnami"
          url  = "https://charts.bitnami.com/bitnami"
        }
      }
    }
    
    values = {
      "auth.postgresPassword"     = var.postgres_password
      "auth.database"            = "myapp"
      "primary.persistence.enabled" = "true"
      "primary.persistence.size"    = "20Gi"
      "primary.persistence.storageClass" = "fast-ssd"
    }
  }
}
```

### Application with Custom Namespace Creation

```terraform
resource "kkp_application_v2" "monitoring_stack" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "kube-prometheus-stack"
  namespace  = "monitoring"
  
  spec {
    namespace_create = true
    
    source {
      helm {
        chart_name    = "kube-prometheus-stack"
        chart_version = "45.7.1"
        repository {
          name = "prometheus-community"
          url  = "https://prometheus-community.github.io/helm-charts"
        }
      }
    }
    
    values = {
      "prometheus.prometheusSpec.retention"    = "30d"
      "prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage" = "50Gi"
      "grafana.adminPassword"                 = var.grafana_password
      "grafana.persistence.enabled"          = "true"
      "grafana.persistence.size"             = "10Gi"
    }
  }
  
  labels = {
    component   = "monitoring"
    environment = "production"
  }
}
```

## Argument Reference

The following arguments are supported:

### General Arguments

- `cluster_id` - (Required) The ID of the cluster to deploy the application to.
- `name` - (Required) The name of the application instance.
- `namespace` - (Required) The Kubernetes namespace to deploy the application to.
- `wait_for_ready` - (Optional) Whether to wait for the application to be ready before completing. Defaults to `false`.
- `labels` - (Optional) A map of labels to assign to the application.

### Spec Configuration

The `spec` block supports:

- `namespace_create` - (Optional) Whether to create the namespace if it doesn't exist. Defaults to `false`.
- `values` - (Optional) A map of Helm values to override defaults.

#### Source Configuration

The `source` block is required and supports:

**Helm Source:**
```terraform
source {
  helm {
    chart_name    = "chart-name"
    chart_version = "1.0.0"
    repository {
      name = "repo-name"
      url  = "https://charts.example.com"
    }
  }
}
```

- `chart_name` - (Required) The name of the Helm chart.
- `chart_version` - (Required) The version of the Helm chart to deploy.
- `repository` - (Required) The Helm repository configuration.
  - `name` - (Required) The name of the repository.
  - `url` - (Required) The URL of the repository.

## Common Application Examples

### Database Applications

**Redis:**
```terraform
values = {
  "auth.enabled"                = "true"
  "auth.password"              = var.redis_password
  "master.persistence.enabled" = "true"
  "master.persistence.size"    = "8Gi"
  "replica.replicaCount"       = "2"
}
```

**PostgreSQL:**
```terraform
values = {
  "auth.postgresPassword"       = var.postgres_password
  "auth.database"              = "myapp"
  "primary.persistence.enabled"   = "true"
  "primary.persistence.size"      = "20Gi"
  "primary.resources.requests.memory" = "256Mi"
}
```

**MySQL:**
```terraform
values = {
  "auth.rootPassword"          = var.mysql_root_password
  "auth.database"             = "myapp"
  "primary.persistence.enabled"  = "true"
  "primary.persistence.size"     = "20Gi"
}
```

### Monitoring Applications

**Prometheus Stack:**
```terraform
values = {
  "prometheus.prometheusSpec.retention" = "15d"
  "prometheus.prometheusSpec.storageSpec.volumeClaimTemplate.spec.resources.requests.storage" = "50Gi"
  "grafana.adminPassword"               = var.grafana_password
  "grafana.persistence.enabled"        = "true"
  "alertmanager.enabled"               = "true"
}
```

### Web Applications

**WordPress:**
```terraform
values = {
  "wordpressUsername"          = "admin"
  "wordpressPassword"          = var.wordpress_password
  "persistence.enabled"        = "true"
  "persistence.size"          = "10Gi"
  "mariadb.auth.rootPassword" = var.mariadb_password
  "mariadb.auth.database"     = "wordpress"
}
```

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the application.
- `created_at` - The timestamp when the application was deployed.
- `status` - The current status of the application deployment.

## Import

Applications can be imported using their cluster ID, namespace, and application name:

```bash
terraform import kkp_application_v2.example cluster-id/namespace/application-name
```

## Notes

- Applications are deployed using Helm charts
- The target namespace must exist unless `namespace_create` is set to `true`
- Some applications may require specific cluster configurations or resources
- Use `wait_for_ready` when you need to ensure the application is running before proceeding
- Helm values depend on the specific chart being deployed - refer to the chart's documentation