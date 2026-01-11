.PHONY: build build-proxy clean

# Default proxy endpoint (can be overridden: make build-proxy PROXY_ENDPOINT=localhost:9443)
PROXY_ENDPOINT ?= host.docker.internal:9443

# Default: Build without proxy mode (production)
build:
	@echo "Compiling KCL platform configuration (production mode)..."
	mkdir -p rendered
	kcl run main.k -o rendered/platform.yaml

# Proxy/Debug mode: Enable proxy to forward requests to local endpoint
build-proxy:
	@echo "Compiling KCL platform configuration (proxy/debug mode)..."
	@echo "Proxy endpoint: $(PROXY_ENDPOINT)"
	mkdir -p rendered
	kcl run main.k -D proxy_mode=true -D proxy_endpoint=$(PROXY_ENDPOINT) -o rendered/platform.yaml

clean:
	rm -rf rendered/*.yaml
