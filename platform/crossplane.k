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
        package = "ghcr.io/vshn/function-appcat-poc:v0.1.0"
    }
}