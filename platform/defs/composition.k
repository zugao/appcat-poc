# Composition-related type schemas for AppCat Framework 2.0
# Service Maintainers import these schemas to define service configurations

# SecretFieldTemplate - Single secret field with templated value
# Supports variable substitution: ${instanceName}, ${namespace}, ${password}
schema SecretFieldTemplate:
    key: str                      # Secret key name (e.g., "host", "port", "password")
    value: str                    # Template with variables (e.g., "${instanceName}-master.${namespace}.svc.cluster.local")

# ConnectionSecretSpec - Connection secret configuration
# Defines the structure and content of connection secrets created by the runtime
schema ConnectionSecretSpec:
    fields: [SecretFieldTemplate] # List of secret fields to create
    passwordPath?: str            # Optional: Helm value path where password is injected (e.g., "auth.password")
    existingSecretPath?: str      # Optional: Helm value path where secret name is injected (e.g., "auth.existingSecret")
