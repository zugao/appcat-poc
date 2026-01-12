# AppCat PoC - Framework 2.0 Implementation

Proof of concept for Application Catalog Framework 2.0 using Crossplane 2.0, KCL, and composition functions.

## Overview

Generic, configuration-driven approach for managing application services:

- **Crossplane 2.0** - Composition functions pipeline
- **KCL** - Type-safe service configuration
- **Go Function** - Generic runtime
- **Crossplane Packages** - Service distribution via xpkg

### Architecture

```
https://miro.com/app/board/uXjVJiczJH0=/
```

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/)
- [Kind](https://kind.sigs.k8s.io/) v0.20+
- [Helm](https://helm.sh/docs/intro/install/) v3.12+
- [kubectl](https://kubernetes.io/docs/tasks/tools/) v1.27+
- [KCL](https://kcl-lang.io/docs/user_docs/getting-started/install) v0.10+
- [Crossplane CLI](https://docs.crossplane.io/latest/cli/) v2.1+
- [Go](https://go.dev/doc/install) 1.21+ (for local debugging)

## Quick Start

### 1. Setup

```bash
make setup    # Create Kind cluster + install Crossplane
```

### 2. Build

```bash
make build    # Build platform, runtime, and redis-service package
```

This:
- Renders platform manifests (Function, Providers, ProviderConfigs)
- Builds and loads the composition function into Kind
- Builds the redis-service xpkg package

### 3. Deploy

```bash
make deploy   # Deploy platform + redis-service configuration
```

This:
- Deploys platform infrastructure
- Deploys `redis-service/configuration/config.yaml` which pulls the xpkg from GHCR

### 4. Create Redis Instance

```bash
make test     # Create test instance
```

Or manually:

```bash
kubectl apply -f - <<EOF
apiVersion: appcat.vshn.io/v1alpha1
kind: XVSHNRedis
metadata:
  name: my-redis
  namespace: default
spec:
  size:
    cpu: "1000m"
    memory: "4Gi"
    disk: "16Gi"
  replicas: 1
  writeConnectionSecretToRef:
    name: redis-credentials
    namespace: default
EOF
```

### 5. Verify

```bash
# Check instance
kubectl get xvshnredis

# Check HelmRelease
kubectl get release -n default

# Get connection details
kubectl get secret redis-credentials -n default -o yaml
```

### 6. Connect

```bash
PASSWORD=$(kubectl get secret redis-credentials -n default -o jsonpath='{.data.password}' | base64 -d)
kubectl port-forward -n default svc/my-redis-master 6379:6379
redis-cli -h localhost -p 6379 -a "$PASSWORD"
```

## Debug Mode

Develop the composition function locally without rebuilding images.

### Workflow

```bash
# 1. Build with proxy mode
make build-proxy

# 2. Deploy
make deploy

# 3. Start local function (in separate terminal)
make debug-start

# 4. Stop debugging
make debug-stop
```

The function pod forwards requests to `host.docker.internal:9443`. Custom endpoint: `make build-proxy PROXY_ENDPOINT=127.18.0.1:9443`

## Makefile Targets

| Target | Description |
|--------|-------------|
| `make setup` | Create Kind cluster + install Crossplane |
| `make build` | Build platform, runtime, and service package |
| `make build-proxy` | Build with proxy mode for debugging |
| `make deploy` | Deploy platform + service configuration |
| `make test` | Create test Redis instance |
| `make debug-start` | Start local function (blocking) |
| `make debug-stop` | Stop local function |
| `make clean` | Clean build artifacts |
| `make kind-delete` | Delete Kind cluster |

## References

- [Crossplane Documentation](https://docs.crossplane.io/)
- [KCL Language](https://kcl-lang.io/)
- [Function SDK Go](https://github.com/crossplane/function-sdk-go)
- [Framework 2.0 ADRs](https://github.com/vshn/application-catalog-docs/tree/master/docs/modules/ROOT/pages/framework)
