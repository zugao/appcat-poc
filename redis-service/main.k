# Main output file for redis-service
# Outputs the XRD and Composition manifests

import service
import composition

# Output XRD and Composition (config.k is embedded in Composition's inline input)
[service.xrd, composition.composition]
