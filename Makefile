.PHONY: help build-all build-all-push clean kind-create kind-delete crossplane-install setup deploy-platform deploy-service deploy-all test

# Default target
help:
	@echo "AppCat PoC - Available targets:"
	@echo ""
	@echo "  Setup:"
	@echo "    kind-create          - Create Kind cluster"
	@echo "    crossplane-install   - Install Crossplane 2.0 via Helm"
	@echo "    setup                - Create cluster + install Crossplane"
	@echo ""
	@echo "  Build:"
	@echo "    build-all            - Build all layers and load runtime into Kind"
	@echo "    build-all-push       - Build all layers and push runtime to registry"
	@echo "    clean                - Clean all build artifacts"
	@echo ""
	@echo "  Deploy:"
	@echo "    deploy-platform      - Deploy platform manifests (Composition, Providers)"
	@echo "    deploy-service       - Deploy service manifests (XRD)"
	@echo "    deploy-all           - Deploy platform + service"
	@echo ""
	@echo "  Test:"
	@echo "    test                 - Create test Redis instance"
	@echo ""
	@echo "  Teardown:"
	@echo "    kind-delete          - Delete Kind cluster"

# Build targets
build-all:
	@echo "Building all layers..."
	@echo "Building redis-service..."
	cd redis-service && $(MAKE) build
	@echo "Building platform..."
	cd platform && $(MAKE) build
	@echo "Building appcat-runtime and loading into Kind..."
	cd appcat-runtime && $(MAKE) kind-load
	@echo "✅ Build complete!"

build-all-push:
	@echo "Building all layers and pushing to registry..."
	@echo "Building redis-service..."
	cd redis-service && $(MAKE) build
	@echo "Building platform..."
	cd platform && $(MAKE) build
	@echo "Building and pushing appcat-runtime..."
	cd appcat-runtime && $(MAKE) xpkg-push
	@echo "✅ Build and push complete!"

clean:
	@echo "Cleaning build artifacts..."
	cd redis-service && $(MAKE) clean
	cd platform && $(MAKE) clean
	cd appcat-runtime && $(MAKE) clean
	@echo "Clean complete!"

# Kind cluster targets
kind-create:
	@echo "Creating Kind cluster 'appcat-poc'..."
	kind create cluster --name appcat-poc
	@echo "Kind cluster created!"

kind-delete:
	@echo "Deleting Kind cluster 'appcat-poc'..."
	kind delete cluster --name appcat-poc
	@echo "Kind cluster deleted!"

# Crossplane installation
crossplane-install:
	@echo "Installing Crossplane 2.0..."
	kind get kubeconfig --name appcat-poc > ~/.kube/config
	helm repo add crossplane-stable https://charts.crossplane.io/stable
	helm repo update
	helm install crossplane crossplane-stable/crossplane \
		--namespace crossplane-system \
		--create-namespace \
		--wait
	@echo "Crossplane installed!"

# Combined setup
setup: kind-create crossplane-install
	@echo "Cluster setup complete!"

# Deploy targets
deploy-platform:
	@echo "Deploying platform manifests..."
	kubectl apply -f platform/rendered/
	@echo "Waiting for providers to be healthy..."
	@echo "Note: This may take a few minutes..."
	kubectl wait --for=condition=Healthy provider/provider-kubernetes --timeout=5m || true
	kubectl wait --for=condition=Healthy provider/provider-helm --timeout=5m || true
	@echo "Platform deployed!"

deploy-service:
	@echo "Deploying service manifests..."
	kubectl apply -f redis-service/rendered/
	@echo "Service deployed!"

deploy-all: deploy-platform deploy-service
	@echo "All manifests deployed!"

# Test target
test:
	@echo "Creating test Redis instance..."
	@cat <<EOF | kubectl apply -f -
	apiVersion: appcat.vshn.io/v1alpha1
	kind: XVSHNRedis
	metadata:
	  name: test-redis
	spec:
	  size:
	    cpu: "500m"
	    memory: "2Gi"
	  replicas: 1
	EOF
	@echo "Test instance created!"
	@echo "Check status with: kubectl get xvshnredis"