# Redis Service XRD Definition
# Defines the Composite Resource (XR) API for VSHNRedis

# XRD for VSHNRedis Composite Resource
xrd = {
    apiVersion = "apiextensions.crossplane.io/v1"
    kind = "CompositeResourceDefinition"
    metadata = {
        name = "xvshnredis.appcat.vshn.io"
    }
    spec = {
        group = "appcat.vshn.io"
        names = {
            kind = "XVSHNRedis"
            plural = "xvshnredis"
        }
        # Crossplane 2.0: Composites
        versions = [
            {
                name = "v1alpha1"
                served = True
                referenceable = True
                schema = {
                    openAPIV3Schema = {
                        type = "object"
                        properties = {
                            # COMPOSITE SPEC - direct user parameters
                            spec = {
                                type = "object"
                                properties = {
                                    # Resource size specification
                                    size = {
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
                                    # Replica count
                                    replicas = {
                                        type = "integer"
                                        description = "Number of replicas"
                                        minimum = 1
                                    }
                                }
                            }
                            status = {
                                type = "object"
                                properties = {}
                            }
                        }
                    }
                }
            }
        ]
    }
}