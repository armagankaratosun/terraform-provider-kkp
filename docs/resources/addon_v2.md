# kkp_addon_v2

Manages cluster addons in KKP. Addons provide additional functionality to Kubernetes clusters such as ingress controllers, monitoring, logging, and networking components.

## Example Usage

### Basic Addon Installation

```terraform
resource "kkp_addon_v2" "nginx_ingress" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "nginx-ingress-controller"
  
  spec {
    variables = {
      "controller.service.type" = "LoadBalancer"
      "controller.replicaCount" = "2"
    }
  }
}
```

### Addon with Wait for Ready

```terraform
resource "kkp_addon_v2" "monitoring" {
  cluster_id      = kkp_cluster_v2.example.id
  name           = "prometheus"
  wait_for_ready = true
  
  spec {
    variables = {
      "prometheus.retention"     = "30d"
      "prometheus.storageClass" = "fast-ssd"
      "grafana.enabled"         = "true"
    }
  }
}
```

### Addon with Custom Configuration

```terraform
resource "kkp_addon_v2" "cert_manager" {
  cluster_id = kkp_cluster_v2.example.id
  name       = "cert-manager"
  
  spec {
    variables = {
      "installCRDs"                = "true"
      "webhook.timeoutSeconds"     = "30"
      "cainjector.replicaCount"    = "2"
    }
  }
  
  labels = {
    environment = "production"
    managed-by  = "terraform"
  }
}
```

## Argument Reference

The following arguments are supported:

### General Arguments

- `cluster_id` - (Required) The ID of the cluster to install the addon on.
- `name` - (Required) The name of the addon to install. Must match available addon names in KKP.
- `wait_for_ready` - (Optional) Whether to wait for the addon to be ready before completing. Defaults to `false`.
- `labels` - (Optional) A map of labels to assign to the addon.

### Spec Configuration

The `spec` block supports:

- `variables` - (Optional) A map of configuration variables for the addon. Available variables depend on the specific addon.

## Common Addon Names

The following addon names are commonly available in KKP:

### Networking
- `cilium` - Cilium CNI plugin
- `canal` - Canal CNI plugin
- `calico` - Calico CNI plugin

### Ingress
- `nginx-ingress-controller` - NGINX Ingress Controller
- `ingress-nginx` - Ingress NGINX Controller

### Monitoring & Logging
- `prometheus` - Prometheus monitoring
- `grafana` - Grafana dashboards
- `loki` - Loki log aggregation
- `jaeger` - Jaeger distributed tracing

### Security
- `cert-manager` - Certificate management
- `oauth` - OAuth2 authentication
- `rbac` - Role-based access control

### Storage
- `openebs` - OpenEBS storage
- `longhorn` - Longhorn distributed storage

### Other
- `kubernetes-dashboard` - Kubernetes web dashboard
- `helm` - Helm package manager
- `multus` - Multus CNI for multiple networks

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `id` - The unique identifier of the addon installation.
- `created_at` - The timestamp when the addon was installed.
- `status` - The current status of the addon.

## Addon Variables

Addon variables are specific to each addon type. Common patterns include:

### NGINX Ingress Controller
```terraform
spec {
  variables = {
    "controller.service.type"                = "LoadBalancer"
    "controller.replicaCount"               = "2"
    "controller.resources.requests.cpu"     = "100m"
    "controller.resources.requests.memory"  = "90Mi"
    "controller.nodeSelector.node-role"     = "ingress"
  }
}
```

### Prometheus
```terraform
spec {
  variables = {
    "prometheus.retention"                   = "15d"
    "prometheus.storageClass"               = "fast"
    "prometheus.resources.requests.memory"  = "2Gi"
    "grafana.enabled"                       = "true"
    "alertmanager.enabled"                  = "true"
  }
}
```

### Cert-Manager
```terraform
spec {
  variables = {
    "installCRDs"                = "true"
    "webhook.timeoutSeconds"     = "30"
    "cainjector.replicaCount"    = "1"
    "startupapicheck.enabled"    = "true"
  }
}
```

## Import

Addons can be imported using their cluster ID and addon name:

```bash
terraform import kkp_addon_v2.example cluster-id/addon-name
```

## Notes

- Some addons may conflict with each other (e.g., multiple CNI plugins)
- Addon availability depends on your KKP installation and configuration
- Some addons require specific cluster configurations or node requirements
- Wait for addon readiness when dependencies exist between addons