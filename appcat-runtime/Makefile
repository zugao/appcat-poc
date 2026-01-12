.PHONY: build kind-load clean

IMAGE ?= ghcr.io/zugao/function-appcat-poc
TAG ?= v0.1.0
KIND_CLUSTER ?= appcat-poc

# Auto-detect platform
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),arm64)
	GOARCH ?= arm64
else ifeq ($(UNAME_M),aarch64)
	GOARCH ?= arm64
else
	GOARCH ?= amd64
endif

# Build Go binary
build:
	@echo "Building Go composition function for linux/$(GOARCH)..."
	@CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -o function-appcat-poc .
	@echo "Binary built: function-appcat-poc"

# Build and load into Kind cluster
kind-load: build
	@echo "Building Docker image for Kind cluster (linux/$(GOARCH))..."
	@docker build -t $(IMAGE):$(TAG) .
	@echo "Loading image into Kind cluster: $(KIND_CLUSTER)..."
	@kind load docker-image $(IMAGE):$(TAG) --name $(KIND_CLUSTER)
	@echo "Image loaded into Kind cluster"

# Clean build artifacts
clean:
	@rm -f function-appcat-poc
	@echo "Cleaned build artifacts"
