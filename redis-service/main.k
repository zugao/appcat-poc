# Main output file for redis-service
# Outputs the XRD and Composition manifests
# Service config data (chart, plans, defaultHelmValues) is inlined in Composition

import service

# Output only the XRD
# Config data is NOT output here - it's consumed by platform layer
[service.xrd]