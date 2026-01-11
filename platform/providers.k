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
        package = "xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v1.2.0"
    }
}

# Helm Provider with namespace-scoped Release support
provider_helm = {
    apiVersion = "pkg.crossplane.io/v1"
    kind = "Provider"
    metadata = {
        name = "provider-helm"
    }
    spec = {
        package = "xpkg.upbound.io/crossplane-contrib/provider-helm:v1.0.6"
        runtimeConfigRef = {
            name = "provider-helm-runtime"
        }
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

# ProviderConfig for Helm - legacy cluster-scoped API (helm.crossplane.io)
# Note: Both old and new APIs configured for compatibility
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

# ClusterProviderConfig for Helm - namespace-scoped API (helm.m.crossplane.io)
# This is used by our namespace-scoped HelmReleases in Crossplane 2.0
clusterproviderconfig_helm = {
    apiVersion = "helm.m.crossplane.io/v1beta1"
    kind = "ClusterProviderConfig"
    metadata = {
        name = "default"
    }
    spec = {
        credentials = {
            source = "InjectedIdentity"
        }
    }
}

# ServiceAccount for provider-helm
serviceaccount_helm = {
    apiVersion = "v1"
    kind = "ServiceAccount"
    metadata = {
        name = "provider-helm-sa"
        namespace = "crossplane-system"
    }
}

# ClusterRoleBinding for provider-helm (cluster-admin for PoC)
clusterrolebinding_helm = {
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind = "ClusterRoleBinding"
    metadata = {
        name = "provider-helm-admin"
    }
    roleRef = {
        apiGroup = "rbac.authorization.k8s.io"
        kind = "ClusterRole"
        name = "cluster-admin"
    }
    subjects = [
        {
            kind = "ServiceAccount"
            name = "provider-helm-sa"
            namespace = "crossplane-system"
        }
    ]
}

# DeploymentRuntimeConfig for provider-helm
# Configures the provider to use the custom service account
deploymentruntimeconfig_helm = {
    apiVersion = "pkg.crossplane.io/v1beta1"
    kind = "DeploymentRuntimeConfig"
    metadata = {
        name = "provider-helm-runtime"
    }
    spec = {
        serviceAccountTemplate = {
            metadata = {
                name = "provider-helm-sa"
            }
        }
    }
}