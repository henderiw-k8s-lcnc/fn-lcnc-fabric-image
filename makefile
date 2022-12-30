VERSION ?= latest
REGISTRY ?= yndd
IMG ?= $(REGISTRY)/fn-lcnc-fabric-image:${VERSION}

# Private Github REPOs
GITHUB_TOKEN ?= $(shell cat ~/.config/github.token)

.PHONY: all
all: test

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

test: fmt vet ## Run tests.

docker-build: test ## Build docker images.
	docker build --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} -f Dockerfile -t ${IMG} .

docker-push: ## Build docker images.
	docker push ${IMG}