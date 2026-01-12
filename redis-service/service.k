# Redis Service XRD Definition
# Defines the Composite Resource (XR) API for XVSHNRedis

import platform.defs.xrd as platform_xrd

# XRD for XVSHNRedis Composite Resource
xrd = {
    apiVersion = "apiextensions.crossplane.io/v2"
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
        scope = "Namespaced"
        defaultCompositionRef = {
            name = "xvshnredis.appcat.vshn.io"
        }
        versions = [
            {
                name = "v1alpha1"
                served = True
                referenceable = True
                schema = {
                    openAPIV3Schema = {
                        type = "object"
                        properties = {
                            spec = {
                                type = "object"
                                properties = {
                                    size = platform_xrd.size_spec_schema
                                    replicas = platform_xrd.replicas_schema
                                    writeConnectionSecretToRef = platform_xrd.write_connection_secret_ref_schema
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
