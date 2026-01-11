# Composition for VSHNRedis
# Imports service config from config.k and embeds it in the pipeline input

import config as redis_config

composition = {
    apiVersion = "apiextensions.crossplane.io/v1"
    kind = "Composition"
    metadata = {
        name = "xvshnredis.appcat.vshn.io"
        labels = {
            service = "redis"
        }
    }
    spec = {
        compositeTypeRef = {
            apiVersion = "appcat.vshn.io/v1alpha1"
            kind = "XVSHNRedis"
        }
        mode = "Pipeline"
        pipeline = [
            {
                step = "render-redis"
                functionRef = {
                    name = "function-appcat-poc"
                }
                # Inline input embedding service config from config.k
                input = {
                    apiVersion = "fn.appcat.vshn.io/v1alpha1"
                    kind = "AppCatServiceConfig"
                    metadata = {
                        name = "redis-config"
                        labels = {
                            service = "redis"
                        }
                    }
                    data = {
                        chart = redis_config.service_config.chart
                        defaultHelmValues = redis_config.service_config.defaultHelmValues
                        mapping = redis_config.service_config.mapping
                        connectionSecret = redis_config.service_config.connectionSecret
                    }
                }
            }
        ]
    }
}
