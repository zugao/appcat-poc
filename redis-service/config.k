# Redis Service Configuration
# Contains: chart info, default Helm values, mapping from XRD spec to Helm values, and connection secret template

# Import schema definitions from platform
import platform.defs.composition

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

    # Connection secret specification - defines the structure and content of connection secrets
    # Runtime will substitute variables: ${instanceName}, ${namespace}, ${password}
    connectionSecret = composition.ConnectionSecretSpec {
        # Tell Helm chart to use our secret instead of creating its own
        existingSecretPath = "auth.existingSecret"

        fields = [
            composition.SecretFieldTemplate {
                key = "password"
                value = "\${password}"
            }
            composition.SecretFieldTemplate {
                key = "redis-password"  # Required by Bitnami chart when using auth.existingSecret
                value = "\${password}"
            }
            composition.SecretFieldTemplate {
                key = "host"
                value = "\${instanceName}-master.\${namespace}.svc.cluster.local"
            }
            composition.SecretFieldTemplate {
                key = "port"
                value = "6379"
            }
            composition.SecretFieldTemplate {
                key = "url"
                value = "redis://default:\${password}@\${instanceName}-master.\${namespace}.svc.cluster.local:6379"
            }
        ]
    }
}
