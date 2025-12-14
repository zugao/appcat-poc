.PHONY: build docker-build docker-push clean

IMAGE ?= ghcr.io/zugao/function-appcat-poc
TAG ?= v0.1.0

build:
	@echo "Building Go composition function..."
	CGO_ENABLED=0 go build -o function-appcat-poc .

docker-build: build
	@echo "Building Docker image..."
	docker build -t $(IMAGE):$(TAG) .

docker-push: docker-build
	@echo "Pushing Docker image..."
	docker push $(IMAGE):$(TAG)

clean:
	rm -f function