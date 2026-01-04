# Composition for VSHNRedis
# Imports service config from redis-service repo to avoid duplication

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
                # INLINE input - imports service config from redis-service repo
                input = {
                    apiVersion = "fn.appcat.vshn.io/v1alpha1"
                    kind = "RedisServiceConfig"
                    metadata = {
                        name = "redis-config"
                    }
                    # Service config imported from redis-service/config.k
                    data = {
                        chart = redis_config.service_config.chart
                        defaultHelmValues = redis_config.service_config.defaultHelmValues
                        mapping = redis_config.service_config.mapping
                    }
                }
            }
        ]
    }
}