# Redis Service Configuration
# Contains: chart info, default Helm values, and mapping from XRD spec to Helm values

# Service configuration (Helm chart info + defaults + mapping)
service_config = {
    # Bitnami Redis Helm chart information
    chart = {
        repository = "https://charts.bitnami.com/bitnami"
        name = "redis"
        defaultVersion = "18.0.0"
    }

    # Default Helm values (partial overrides, rest comes from chart defaults)
    defaultHelmValues = {
        architecture = "standalone"
        auth = {
            enabled = True
        }
        image = {
            # Use a tag that is known to exist in Docker Hub; the chart's default tag was pruned
            tag = "latest"
        }
    }

    # Mapping: XRD spec field paths to Helm value paths
    # This tells the function how to inject user runtime parameters into helm values
    mapping = {
        "spec.size.cpu" = "master.resources.requests.cpu"
        "spec.size.memory" = "master.resources.requests.memory"
        "spec.size.disk" = "master.persistence.size"
        "spec.replicas" = "master.count"
    }
}
