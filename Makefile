BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
WORKSPACE := $(shell pwd)

.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

id: ## Output BUILD_ID being used
	@echo $(BUILD_ID)

debug: ## Output internal make variables
	@echo BUILD_ID = $(BUILD_ID)
	@echo IMAGE_NAME = $(IMAGE_NAME)
	@echo WORKSPACE = $(WORKSPACE)

deps:
	go get .

build-service: ## Build the main Go service
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o atlantis .

deps-test:
	go get -t

test: ## Run tests, coverage reports, and clean (coverage taints the compiled code)
	go test -v ./...

test-coverage:
	go test -v ./... -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

dist: ## Package up everything in static/ using go-bindata-assetfs so it can be served by a single binary
	go-bindata-assetfs static/...

