BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
WORKSPACE := $(shell pwd)
PKG := $(shell go list ./... | grep -v e2e | grep -v vendor | grep -v static)

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
	@echo PKG = $(PKG)

deps: ## Download dependencies
	go get -u github.com/kardianos/govendor

build-service: ## Build the main Go service
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o atlantis .

test: ## Run tests, coverage reports, and clean (coverage taints the compiled code)
	go test $(PKG)

test-coverage:
	go test $(PKG) -coverprofile=c.out
	go tool cover -html=c.out -o coverage.html

dist: ## Package up everything in static/ using go-bindata-assetfs so it can be served by a single binary
	go-bindata-assetfs -pkg server static/... && mv bindata_assetfs.go server

release: ## Create packages for a release
	gox -os="darwin linux" -arch="amd64"

vendor-status:
	@govendor status

fmt: ## Run goimports (which also formats)
	goimports -w $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/bindata_assetfs.go")
