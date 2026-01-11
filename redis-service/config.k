# Redis Service Configuration
# Contains: chart info, default Helm values, mapping from XRD spec to Helm values, and connection secret template

# SecretFieldTemplate - Single secret field with templated value
# Supports variable substitution: ${instanceName}, ${namespace}, ${password}
schema SecretFieldTemplate:
    key: str        # Secret key name (e.g., "host", "port", "password")
    value: str      # Template with variables

# ConnectionSecretSpec - Connection secret configuration
schema ConnectionSecretSpec:
    fields: [SecretFieldTemplate]  # List of secret fields to create
    passwordPath?: str             # Optional: Helm value path where password is injected
    existingSecretPath?: str       # Optional: Helm value path where secret name is injected

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
    connectionSecret = ConnectionSecretSpec {
        passwordPath = ""                       # Don't inject password - use existingSecret instead
        existingSecretPath = "auth.existingSecret"  # Tell chart to use our secret
        fields = [
            SecretFieldTemplate {
                key = "password"
                value = "\${password}"
            }
            SecretFieldTemplate {
                key = "redis-password"  # Required by Bitnami chart when using existingSecret
                value = "\${password}"
            }
            SecretFieldTemplate {
                key = "host"
                value = "\${instanceName}-master.\${namespace}.svc.cluster.local"
            }
            SecretFieldTemplate {
                key = "port"
                value = "6379"
            }
            SecretFieldTemplate {
                key = "url"
                value = "redis://default:\${password}@\${instanceName}-master.\${namespace}.svc.cluster.local:6379"
            }
        ]
    }
}
