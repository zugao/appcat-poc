# AppCat PoC - Framework 2.0 Implementation

Proof of concept for the Application Catalog Framework 2.0 using Crossplane 2.0, KCL, and composition functions.

## Overview

This PoC demonstrates a generic, configuration-driven approach to managing application services using:

- **Crossplane 2.0** - Kubernetes-native control plane with composition functions
- **KCL (Kubernetes Configuration Language)** - Type-safe configuration for service definitions
- **Go Composition Function** - Generic runtime that merges service defaults with user parameters
- **Namespace-scoped composites** - Services deployed within user namespaces

### Key Features

- ✅ Generic composition function (no service-specific code)
- ✅ KCL-based service configuration with type safety
- ✅ Template-based connection secret generation
- ✅ Helm chart integration
- ✅ Local debugging support with proxy mode

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│ User creates XVSHNRedis instance                                 │
│   spec: { size: { cpu, memory, disk }, replicas }               │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ Composition Function (Go)                                        │
│ 1. Reads service config from Composition input (KCL-generated)  │
│ 2. Merges defaultHelmValues + user spec (via mapping)          │
│ 3. Generates: HelmRelease + Secret                             │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│ Provider-Helm deploys Redis chart with merged values            │
└─────────────────────────────────────────────────────────────────┘
```

## Prerequisites

- Docker or OrbStack
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) v0.20+
- [Helm](https://helm.sh/docs/intro/install/) v3.12+
- [kubectl](https://kubernetes.io/docs/tasks/tools/) v1.27+
- [KCL](https://kcl-lang.io/docs/user_docs/getting-started/install) v0.10+
- [Go](https://go.dev/doc/install) 1.21+ (for local debugging)

## Quick Start (Production Mode)

### 1. Setup Cluster and Crossplane

```bash
make setup
```

This creates a Kind cluster named `appcat-poc` and installs Crossplane 2.0.

### 2. Build and Load Runtime

```bash
make build-all
```

This:
- Compiles KCL service configurations
- Builds the Go composition function
- Loads the function image into Kind

### 3. Deploy Platform and Service

```bash
make deploy-all
```

Wait for providers to become healthy (~2-3 minutes):

```bash
kubectl get providers
```

### 4. Create a Redis Instance

```bash
make test
```

Or create manually:

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

### 5. Check Status

```bash
# Check composite status
kubectl get xvshnredis

# Check HelmRelease
kubectl get release -n default

# Check connection secret
kubectl get secret redis-credentials -n default -o yaml

# Check Redis pods
kubectl get pods -n default | grep redis
```

### 6. Connect to Redis

```bash
# Get password
PASSWORD=$(kubectl get secret redis-credentials -n default -o jsonpath='{.data.password}' | base64 -d)

# Port forward
kubectl port-forward -n default svc/my-redis-master 6379:6379

# Connect with redis-cli
redis-cli -h localhost -p 6379 -a "$PASSWORD"
```

## Debug/Proxy Mode (Local Development)

When developing the composition function, use proxy mode to test changes without rebuilding and reloading images.

### How Proxy Mode Works

```
┌──────────────────────┐
│ Function Pod in Kind │
│  (with --proxy flag) │
└──────────┬───────────┘
           │
           │ Forwards all requests to
           ▼
┌─────────────────────────┐
│ host.docker.internal:9443 │  ← Your local machine
└─────────────────────────┘
           │
           ▼
    ┌──────────────┐
    │ go run .     │  ← Runs latest code
    │ --insecure   │
    │ --addr :9443 │
    └──────────────┘
```

### Debug Workflow

#### 1. Build in Proxy Mode

```bash
make build-all-proxy
```

This renders platform manifests with proxy enabled (no image loading needed).

**Custom proxy endpoint:**
```bash
# Use a different endpoint (e.g., for remote debugging)
make build-all-proxy PROXY_ENDPOINT=192.168.1.100:9443
```

Default endpoint is `host.docker.internal:9443` (works for Docker Desktop and OrbStack on macOS).

#### 2. Deploy Platform

```bash
make deploy-platform
make deploy-service
```

#### 3. Start Local Function

```bash
make debug-start
```

This runs the composition function locally on port 9443. The function pod in the cluster will forward all requests to your local endpoint.

**Keep this terminal open** - you'll see live logs of function execution.

#### 4. Make Code Changes

Edit Go files in `appcat-runtime/`. Changes take effect immediately after saving:

1. Stop the local function (Ctrl+C)
2. Make your code changes
3. Restart with `make debug-start`

No need to rebuild images or reload into Kind!

#### 5. Test Your Changes

In another terminal:

```bash
# Create/update test instance
make test

# Watch function logs in the debug-start terminal
```

#### 6. Stop Debugging

```bash
# In another terminal
make debug-stop

# Or just Ctrl+C in the debug-start terminal
```

### Switch Back to Production Mode

To disable proxy mode and use the in-cluster image:

```bash
# Rebuild without proxy
make build-all

# Redeploy platform
make deploy-platform

# Restart function pod
kubectl delete pod -n crossplane-system -l pkg.crossplane.io/function=function-appcat-poc
```

## Project Structure

```
appcat-poc/
├── Makefile                       # Root build orchestration
├── README.md                      # This file
│
├── appcat-runtime/                # Composition function (Go)
│   ├── main.go                    # Entry point
│   ├── manager.go                 # Request handler
│   ├── merge.go                   # Config merging logic
│   ├── resources.go               # Resource generation
│   ├── builders.go                # Kubernetes resource builders
│   ├── Dockerfile                 # Function container image
│   └── Makefile                   # Build/push targets
│
├── platform/                      # Platform infrastructure (KCL)
│   ├── main.k                     # Output aggregation
│   ├── crossplane.k               # Function deployment
│   ├── providers.k                # Provider installations
│   ├── Makefile                   # KCL rendering
│   └── rendered/                  # Generated YAML manifests
│       └── platform.yaml
│
├── redis-service/                 # Redis service definition (KCL)
│   ├── main.k                     # Output aggregation
│   ├── config.k                   # Service configuration
│   ├── service.k                  # XRD definition
│   ├── composition.k              # Composition + inline input
│   ├── Makefile                   # KCL rendering
│   └── rendered/                  # Generated YAML manifests
│       └── redis-service.yaml
│
└── examples/
    └── redis-instance.yaml        # Sample instance
```

## Service Configuration (redis-service/config.k)

Services are defined in KCL with:

```kcl
service_config = {
    # Helm chart information
    chart = {
        repository = "https://charts.bitnami.com/bitnami"
        name = "redis"
        defaultVersion = "18.0.0"
    }

    # Default Helm values (baseline)
    defaultHelmValues = {
        architecture = "standalone"
        auth.enabled = True
        image.tag = "latest"
    }

    # Mapping: XRD spec → Helm values
    mapping = {
        "spec.size.cpu" = "master.resources.requests.cpu"
        "spec.size.memory" = "master.resources.requests.memory"
        "spec.size.disk" = "master.persistence.size"
        "spec.replicas" = "master.count"
    }

    # Connection secret template
    connectionSecret = {
        existingSecretPath = "auth.existingSecret"  # Tell Helm to use our secret
        fields = [
            { key = "password", value = "${password}" }
            { key = "host", value = "${instanceName}-master.${namespace}.svc.cluster.local" }
            { key = "port", value = "6379" }
            { key = "url", value = "redis://default:${password}@${instanceName}-master.${namespace}.svc.cluster.local:6379" }
        ]
    }
}
```

## Makefile Targets

### Production Workflow

| Target | Description |
|--------|-------------|
| `make setup` | Create Kind cluster + install Crossplane |
| `make build-all` | Build all layers + load runtime into Kind |
| `make deploy-all` | Deploy platform + service manifests |
| `make test` | Create test Redis instance |
| `make clean` | Clean all build artifacts |
| `make kind-delete` | Delete Kind cluster |

### Debug Workflow

| Target | Description |
|--------|-------------|
| `make build-all-proxy` | Build with proxy mode enabled |
| `make build-all-proxy PROXY_ENDPOINT=<host:port>` | Build with custom proxy endpoint |
| `make debug-start` | Start local function (blocking) |
| `make debug-stop` | Stop local debugging function |

## Troubleshooting

### Function pod in CrashLoopBackOff

Check logs:
```bash
kubectl logs -n crossplane-system -l pkg.crossplane.io/function=function-appcat-poc
```

Common causes:
- Proxy mode enabled but local function not running
- Image not loaded into Kind cluster

### "connectionSecret not found in merged config"

The function pod is using an old image. Rebuild and reload:

```bash
cd appcat-runtime && make kind-load
kubectl delete pod -n crossplane-system -l pkg.crossplane.io/function=function-appcat-poc
```

### Redis instance stuck in "Synced: False"

Check composition function logs:
```bash
kubectl logs -n crossplane-system -l pkg.crossplane.io/function=function-appcat-poc --tail=50
```

### Proxy mode not connecting

Ensure local function is running:
```bash
# Check if process is running
ps aux | grep "go run.*appcat-runtime"

# Check if port is listening
lsof -i :9443
```

If using a custom proxy endpoint, verify:
1. The endpoint is reachable from inside the Kind container
2. The local function is listening on the correct address/port
3. Firewall rules allow the connection

**Example for different environments:**
- **Docker Desktop/OrbStack (macOS):** `host.docker.internal:9443` (default)
- **Linux with host network:** `localhost:9443` or `127.0.0.1:9443`
- **Remote machine:** Use the machine's IP address (e.g., `192.168.1.100:9443`)

## Development Guide

### Adding a New Service (e.g., PostgreSQL)

1. Create `postgresql-service/` directory
2. Define service in `config.k`:
   - Chart info (Bitnami PostgreSQL)
   - Default Helm values
   - Parameter mappings
   - Connection secret template
3. Create `service.k` with XRD definition
4. Create `composition.k` referencing the function
5. Import in `platform/main.k`
6. Build and test

**No changes to Go code needed!** The runtime is fully generic.

### Modifying the Composition Function

When making changes to `appcat-runtime/`:

1. Use proxy mode: `make build-all-proxy`
2. Deploy: `make deploy-platform`
3. Start debugging: `make debug-start`
4. Make code changes
5. Restart debug session (Ctrl+C, then `make debug-start`)
6. Test changes

### Running Unit Tests

```bash
cd appcat-runtime
go test ./...
```

## Next Steps

- [ ] Add service plans/tiers (t-shirt sizing)
- [ ] Implement comprehensive status field aggregation
- [ ] Add backup/restore integration (K8up)
- [ ] Implement observability (metrics, logs)
- [ ] Add PostgreSQL service
- [ ] Add validation and constraints
- [ ] Implement testing strategy (unit, integration, e2e)

## References

- [Crossplane 2.0 Documentation](https://docs.crossplane.io/)
- [KCL Language Documentation](https://kcl-lang.io/)
- [Function SDK Go](https://github.com/crossplane/function-sdk-go)
- [Framework 2.0 ADRs](https://github.com/vshn/application-catalog-docs/tree/master/docs/modules/ROOT/pages/framework)
