# Main output file for redis-service
# Outputs ONLY the XRD manifest
# Service config data (chart, plans, defaultHelmValues) will be inlined in Composition by platform layer

import service

# Output only the XRD
# Config data is NOT output here - it's consumed by platform layer
[service.xrd]