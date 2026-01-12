# Provider Configurations
# Defines Helm provider with in-cluster configuration

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

# ClusterProviderConfig for Helm - namespace-scoped managed resource API
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

# ClusterRoleBinding for provider-helm
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