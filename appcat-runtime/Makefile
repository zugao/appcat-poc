.PHONY: build docker-build docker-push clean

IMAGE ?= ghcr.io/zugao/function-appcat-poc
TAG ?= v0.1.0
PLATFORM ?= linux/amd64

build:
	@echo "Building Go composition function..."
	CGO_ENABLED=0 go build -o function-appcat-poc .

docker-build: build
	@echo "Building Docker image for local testing ($(PLATFORM))..."
	docker buildx build --platform $(PLATFORM) --load -t $(IMAGE):$(TAG) .

docker-push: docker-build
	@echo "Pushing multi-arch Docker image..."
	docker buildx build --platform $(PLATFORM) --push -t $(IMAGE):$(TAG) .

clean:
	rm -f function-appcat-poc
