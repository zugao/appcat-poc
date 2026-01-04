# input.k - AppCatServiceConfig Schema Definition
# Defines the structure for Composition inline input
# This is NOT outputted as a resource - it's for documentation and reuse

import config

# AppCatServiceConfig inline input structure
# Used in Composition.spec.pipeline[].input
# This defines the generic input structure that works for all services (Redis, PostgreSQL, MySQL, etc.)
appCatServiceConfig = {
    apiVersion = "fn.appcat.vshn.io/v1alpha1"
    kind = "AppCatServiceConfig"
    metadata = {
        name = "redis-config"
        labels = {
            service = "redis"
        }
    }
    data = {
        # Helm chart information
        chart = config.service_config.chart

        # Service-level default Helm values
        defaultHelmValues = config.service_config.defaultHelmValues

        # Mapping: XRD spec paths -> Helm value paths
        # This tells the runtime function how to inject user parameters into Helm values
        mapping = config.service_config.mapping
    }
}