# Main output file for platform layer
# Outputs all platform infrastructure manifests

import crossplane
import providers
import composition

# Output all platform manifests
[
    crossplane.namespace,
    crossplane.function_package,
    providers.provider_kubernetes,
    providers.provider_helm,
    providers.providerconfig_kubernetes,
    providers.providerconfig_helm,
    composition.composition
]