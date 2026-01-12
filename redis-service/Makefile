.PHONY: build clean build-xpkg clean-xpkg

# Package configuration
XPKG_NAME := redis-service
XPKG_VERSION := v0.1.0
XPKG_OUTPUT := $(XPKG_NAME)-$(XPKG_VERSION).xpkg

build:
	@echo "Compiling KCL service configuration..."
	mkdir -p rendered
	kcl run main.k -D output_type=yaml -o rendered/redis-service.yaml

# Build Crossplane package (xpkg)
build-xpkg: build
	@echo "Building Crossplane package: $(XPKG_OUTPUT)"
	@mkdir -p package
	@cp crossplane.yaml package/
	@cp rendered/redis-service.yaml package/
	@crossplane xpkg build --package-root=package --package-file=$(XPKG_OUTPUT)
	@echo "✅ Package built: $(XPKG_OUTPUT)"

# Push package to registry (requires XPKG_REGISTRY variable)
push-xpkg: build-xpkg
	@if [ -z "$(XPKG_REGISTRY)" ]; then \
		echo "❌ Error: XPKG_REGISTRY variable not set"; \
		echo "Usage: make push-xpkg XPKG_REGISTRY=xpkg.upbound.io/myorg/redis-service"; \
		exit 1; \
	fi
	@echo "Pushing package to $(XPKG_REGISTRY):$(XPKG_VERSION)"
	@crossplane xpkg push $(XPKG_REGISTRY):$(XPKG_VERSION) -f $(XPKG_OUTPUT)
	@echo "✅ Package pushed to $(XPKG_REGISTRY):$(XPKG_VERSION)"

clean:
	rm -rf rendered/*.yaml

clean-xpkg:
	rm -rf package *.xpkg