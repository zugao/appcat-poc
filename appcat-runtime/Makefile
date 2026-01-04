.PHONY: build docker-build docker-build-kind kind-load xpkg-build xpkg-push clean

IMAGE ?= ghcr.io/zugao/function-appcat-poc
TAG ?= v0.1.3
KIND_CLUSTER ?= appcat-poc
XPKG_OUTPUT ?= function-appcat-poc.xpkg

# Auto-detect platform (use native architecture for Kind clusters)
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),arm64)
	PLATFORM ?= linux/arm64
	GOARCH ?= arm64
else ifeq ($(UNAME_M),aarch64)
	PLATFORM ?= linux/arm64
	GOARCH ?= arm64
else
	PLATFORM ?= linux/amd64
	GOARCH ?= amd64
endif

build:
	@echo "Building Go composition function for $(PLATFORM)..."
	CGO_ENABLED=0 GOOS=linux GOARCH=$(GOARCH) go build -o function-appcat-poc .

docker-build: build
	@echo "Building Docker image for runtime ($(PLATFORM))..."
	docker buildx build --platform $(PLATFORM) --load -t $(IMAGE):$(TAG) .

docker-build-kind: build
	@echo "Building Docker image for Kind cluster ($(PLATFORM))..."
	docker build -t $(IMAGE):$(TAG) .

kind-load: docker-build-kind
	@echo "Loading image into Kind cluster: $(KIND_CLUSTER)..."
	kind load docker-image $(IMAGE):$(TAG) --name $(KIND_CLUSTER)
	@echo "✅ Image loaded into Kind cluster"

xpkg-build: docker-build-kind
	@echo "Building Crossplane function package (.xpkg)..."
	crossplane xpkg build \
		--package-root=. \
		--package-file=$(XPKG_OUTPUT) \
		--embed-runtime-image=$(IMAGE):$(TAG)
	@echo "✅ Function package built: $(XPKG_OUTPUT)"

xpkg-build-kind: docker-build-kind xpkg-build kind-load
	@echo "✅ Function package built and loaded into Kind cluster"

xpkg-push: xpkg-build
	@echo "Pushing Crossplane function package to registry..."
	crossplane xpkg push \
		--package-files=$(XPKG_OUTPUT) \
		$(IMAGE):$(TAG)
	@echo "✅ Function package pushed: $(IMAGE):$(TAG)"

clean:
	rm -f function-appcat-poc $(XPKG_OUTPUT)
