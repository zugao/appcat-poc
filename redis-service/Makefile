.PHONY: build push clean

# Package configuration
XPKG_REGISTRY ?= ghcr.io/zugao/redis-service
XPKG_VERSION := v0.1.0

# Build Crossplane package (xpkg)
build:
	@echo "Building redis-service package..."
	@mkdir -p rendered package
	@echo "Rendering KCL configuration..."
	@kcl run main.k -D output_type=yaml -o rendered/redis-service.yaml
	@echo "Building xpkg..."
	@cp crossplane.yaml package/
	@cp rendered/redis-service.yaml package/
	@crossplane xpkg build --package-root=package --package-file=redis-service-$(XPKG_VERSION).xpkg
	@echo "Package built: redis-service-$(XPKG_VERSION).xpkg"

# Push package to registry
push: build
	@echo "Pushing package to $(XPKG_REGISTRY):$(XPKG_VERSION)"
	@crossplane xpkg push $(XPKG_REGISTRY):$(XPKG_VERSION) -f redis-service-$(XPKG_VERSION).xpkg
	@echo "Package pushed to $(XPKG_REGISTRY):$(XPKG_VERSION)"

# Clean build artifacts
clean:
	@rm -rf rendered package *.xpkg
	@echo "Cleaned build artifacts"
