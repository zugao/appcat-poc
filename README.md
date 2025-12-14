# AppCat PoC - Runtime Config Merging

Proof-of-concept for mapping-based configuration merging between service defaults and user runtime parameters.

## Architecture

**3-folder structure** (simulates separate repositories):
- `redis-service/` - Service config (KCL): chart, default Helm values, XRD spec → Helm value mappings
- `platform/` - Platform config (KCL): imports service config, defines Composition
- `appcat-runtime/` - Generic function (Go): merges service defaults with user runtime params

## How It Works

1. **Service** defines mapping: `spec.size.cpu` → `master.resources.requests.cpu`
2. **Platform** imports service config into Composition input (KCL import)
3. **Function** at runtime:
   - Starts with service's `defaultHelmValues`
   - Uses mapping to inject user XRD spec values into Helm values
   - Generates: Namespace, Secret, HelmRelease

## Quick Start

```bash
# 1. Build all layers
make build-all

# 2. Setup Kind cluster with Crossplane
make setup

# 3. Deploy to cluster
make deploy-all

# 4. Create test instance
make test
```

## Individual Targets

```bash
make help                # Show all available targets
make kind-create         # Create Kind cluster
make crossplane-install  # Install Crossplane via Helm
make deploy-platform     # Deploy Composition and Providers
make deploy-service      # Deploy XRD
make kind-delete         # Teardown cluster
```

## Example

**User creates:**
```yaml
apiVersion: appcat.vshn.io/v1alpha1
kind: XVSHNRedis
metadata:
  name: my-redis
spec:
  size:
    cpu: "1000m"
    memory: "4Gi"
  replicas: 3
```

**Function merges** service defaults with user spec → final Helm values