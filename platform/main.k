# Main output file for platform layer
# Outputs all platform infrastructure manifests

import crossplane
import providers

# Output all platform manifests
[
    crossplane.namespace,
    crossplane.function_package,
    crossplane.deploymentruntimeconfig_function,
    crossplane.serviceaccount_function,
    crossplane.clusterrolebinding_function,
    providers.serviceaccount_helm,
    providers.clusterrolebinding_helm,
    providers.deploymentruntimeconfig_helm,
    providers.provider_helm,
    providers.clusterproviderconfig_helm,
]