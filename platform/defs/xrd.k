# XRD schema patterns for AppCat services

# size_spec_schema - Standard size specification for all services
size_spec_schema = {
    type = "object"
    properties = {
        cpu = {
            type = "string"
            description = "CPU request (e.g., '500m', '1000m')"
        }
        memory = {
            type = "string"
            description = "Memory request (e.g., '2Gi', '4Gi')"
        }
        disk = {
            type = "string"
            description = "Disk size (e.g., '16Gi', '32Gi')"
        }
    }
}

# replicas_schema - Standard replicas specification
replicas_schema = {
    type = "integer"
    description = "Number of replicas"
    minimum = 1
}

# write_connection_secret_ref_schema - Standard secret ref for all services
write_connection_secret_ref_schema = {
    type = "object"
    properties = {
        name = {
            type = "string"
            description = "Name of the connection secret"
        }
        namespace = {
            type = "string"
            description = "Namespace for the connection secret"
        }
    }
}
