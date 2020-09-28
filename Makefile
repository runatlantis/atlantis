BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
#WORKSPACE := $(shell pwd)
PKG := $(shell go list ./... | grep -v e2e | grep -v vendor | grep -v static | grep -v mocks | grep -v testing)
PKG_COMMAS := $(shell go list ./... | grep -v e2e | grep -v vendor | grep -v static | grep -v mocks | grep -v testing | tr '\n' ',')
IMAGE_NAME := runatlantis/atlantis

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

build-service: ## Build the main Go service
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -v -o atlantis .

go-generate: ## Run go generate in all packages
	go generate $(PKG)

regen-mocks: ## Delete all mocks and matchers and then run go generate to regen them.
	find . -type f | grep mocks/mock_ | grep -v vendor | xargs rm
	find . -type f | grep mocks/matchers | grep -v vendor | xargs rm
	@# not using $(PKG) here because that includes directories that have now
	@# been made empty, causing go generate to fail.
	go list ./... | grep -v e2e | grep -v vendor | grep -v static | xargs go generate

test: ## Run tests
	@go test -mod=vendor -short $(PKG)

test-all: ## Run tests including integration
	@go test -mod=vendor  $(PKG)

test-coverage:
	@mkdir -p .cover
	@go test -covermode atomic -coverprofile .cover/cover.out $(PKG)

test-coverage-html:
	@mkdir -p .cover
	@go test -covermode atomic -coverpkg $(PKG_COMMAS) -coverprofile .cover/cover.out $(PKG)
	go tool cover -html .cover/cover.out

dev-docker:
	GOOS=linux GOARCH=amd64 go build -mod=vendor -o atlantis .
	docker build -f dev-Dockerfile -t atlantis-dev .

dist: ## Package up everything in static/ using go-bindata-assetfs so it can be served by a single binary
	rm -f server/static/bindata_assetfs.go && go-bindata-assetfs -pkg static -prefix server server/static/... && mv bindata_assetfs.go server/static

release: ## Create packages for a release
	docker run -v $$(pwd):/go/src/github.com/runatlantis/atlantis circleci/golang:1.14 sh -c 'cd /go/src/github.com/runatlantis/atlantis && scripts/binary-release.sh'

fmt: ## Run goimports (which also formats)
	goimports -w $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")

lint: ## Run linter locally
	golangci-lint run

check-lint: ## Run linter in CI/CD. If running locally use 'lint'
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ./bin v1.23.8
	./bin/golangci-lint run -j 4

check-fmt: ## Fail if not formatted
	if [[ $$(goimports -l $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")) ]]; then exit 1; fi

end-to-end-deps: ## Install e2e dependencies
	./scripts/e2e-deps.sh

end-to-end-tests: ## Run e2e tests
	./scripts/e2e.sh

website-dev:
	yarn website:dev
