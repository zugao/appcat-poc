# Provider Configurations
# Defines Kubernetes and Helm providers with in-cluster configurations

# Kubernetes Provider
provider_kubernetes = {
    apiVersion = "pkg.crossplane.io/v1"
    kind = "Provider"
    metadata = {
        name = "provider-kubernetes"
    }
    spec = {
        package = "xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.15.0"
    }
}

# Helm Provider
provider_helm = {
    apiVersion = "pkg.crossplane.io/v1"
    kind = "Provider"
    metadata = {
        name = "provider-helm"
    }
    spec = {
        package = "xpkg.upbound.io/crossplane-contrib/provider-helm:v0.20.0"
    }
}

# ProviderConfig for Kubernetes (in-cluster)
providerconfig_kubernetes = {
    apiVersion = "kubernetes.crossplane.io/v1alpha1"
    kind = "ProviderConfig"
    metadata = {
        name = "default"
    }
    spec = {
        credentials = {
            source = "InjectedIdentity"
        }
    }
}

# ProviderConfig for Helm (in-cluster)
providerconfig_helm = {
    apiVersion = "helm.crossplane.io/v1beta1"
    kind = "ProviderConfig"
    metadata = {
        name = "default"
    }
    spec = {
        credentials = {
            source = "InjectedIdentity"
        }
    }
}