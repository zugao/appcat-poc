# Crossplane Infrastructure
# Defines the Function package for the composition function

# Crossplane system namespace
namespace = {
    apiVersion = "v1"
    kind = "Namespace"
    metadata = {
        name = "crossplane-system"
    }
}

# Function package - references the composition function container image
function_package = {
    apiVersion = "pkg.crossplane.io/v1"
    kind = "Function"
    metadata = {
        name = "function-appcat-poc"
    }
    spec = {
        package = "ghcr.io/zugao/function-appcat-poc:v0.1.3"
        runtimeConfigRef = {
            name = "function-appcat-poc"
        }
    }
}

deploymentruntimeconfig_function= {
    apiVersion = "pkg.crossplane.io/v1beta1"
    kind = "DeploymentRuntimeConfig"
    metadata = {
        name = "function-appcat-poc"
    }
    spec = {
        serviceAccountTemplate = {
            metadata = {
                name = "function-appcat-poc"
            }
        }
        deploymentTemplate = {
            spec = {
                replicas = 1
                selector = {
                    matchLabels = {
                        "pkg.crossplane.io/function" = "function-appcat-poc"
                    }
                }
                template = {
                    metadata = {
                        labels = {
                            "pkg.crossplane.io/function" = "function-appcat-poc"
                        }
                    }
                    spec = {
                        serviceAccountName = "function-appcat-poc"
                        containers = [
                            {
                                name = "package-runtime"
                                args = [
                                    "--proxy"
                                    "host.docker.internal:9443"
                                ]
                                image = "ghcr.io/zugao/function-appcat-poc:v0.1.3"
                                ports = [
                                    {
                                        name = "grpc"
                                        containerPort = 9443
                                        protocol = "TCP"
                                    },
                                ]
                            }
                        ]
                    }
                }
            }
        }
    }
}

clusterrolebinding_function = {
    apiVersion = "rbac.authorization.k8s.io/v1"
    kind = "ClusterRoleBinding"
    metadata = {
        name = "function-appcat-poc-admin"
    }
    roleRef = {
        apiGroup = "rbac.authorization.k8s.io"
        kind = "ClusterRole"
        name = "cluster-admin"
    }
    subjects = [
        {
            kind = "ServiceAccount"
            name = "function-appcat-poc"
            namespace = "crossplane-system"
        }
    ]
}

serviceaccount_function = {
    apiVersion = "v1"
    kind = "ServiceAccount"
    metadata = {
        name = "function-appcat-poc"
        namespace = "crossplane-system"
    }
}
