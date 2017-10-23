BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
WORKSPACE := $(shell pwd)
PKG := $(shell go list ./... | grep -v e2e | grep -v vendor | grep -v static)
IMAGE_NAME := hootsuite/atlantis

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
	go get -u github.com/golang/dep/cmd/dep
	dep ensure

build-service: ## Build the main Go service
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o atlantis .

go-generate: ## Run go generate in all packages
	go generate $(PKG)

test: ## Run tests
	go test $(PKG)

test-coverage:
	./scripts/coverage.sh $(PKG)

test-coverage-html:
	./scripts/coverage.sh $(PKG)
	go tool cover -html .cover/cover.out

dist: ## Package up everything in static/ using go-bindata-assetfs so it can be served by a single binary
	rm -f server/static/bindata_assetfs.go && go-bindata-assetfs -pkg static -prefix server server/static/... && mv bindata_assetfs.go server/static

release: ## Create packages for a release
	./scripts/binary-release.sh

fmt: ## Run goimports (which also formats)
	goimports -w $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")

check-fmt: ## Fail if not formatted
	go get golang.org/x/tools/cmd/goimports
	goimports -d $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")

check-govet: ## Fail if go vet finds errors
	go vet $$(go list ./... | grep -v vendor | grep -v static)

megacheck: ## Fail if megacheck finds errors
	go get honnef.co/go/tools/cmd/megacheck
	megacheck $$(go list ./... | grep -v vendor | grep -v static)

end-to-end-deps: ## Install e2e dependencies
	./scripts/e2e-deps.sh

end-to-end-tests: ## Run e2e tests
	./scripts/e2e.sh

generate-website-html: ## Generate HTML for website
	cd website/src && hugo -d ../html

upload-website-html: ## Upload generated website to s3
	aws s3 rm s3://atlantis.run/ --recursive
	aws s3 sync website/html/ s3://atlantis.run/
	rm -rf website/html/
