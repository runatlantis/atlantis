BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
WORKSPACE := $(shell pwd)
PKG := $(shell go list ./... | grep -v e2e | grep -v static | grep -v mocks | grep -v testing)
PKG_COMMAS := $(shell go list ./... | grep -v e2e | grep -v static | grep -v mocks | grep -v testing | tr '\n' ',')
IMAGE_NAME := runatlantis/atlantis

.DEFAULT_GOAL := help

.PHONY: help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z\/_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: id
id: ## Output BUILD_ID being used
	@echo $(BUILD_ID)

.PHONY: debug
debug: ## Output internal make variables
	@echo BUILD_ID = $(BUILD_ID)
	@echo IMAGE_NAME = $(IMAGE_NAME)
	@echo WORKSPACE = $(WORKSPACE)
	@echo PKG = $(PKG)

.PHONY: build-service
build-service: ## Build the main Go service
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o atlantis .

.PHONY: build
build: build-service ## Runs make build-service

.PHONY: all
all: build-service ## Runs make build-service

.PHONY: clean
clean: ## Cleans compiled binary
	@rm -f atlantis

.PHONY: go-generate
go-generate: ## Run go generate in all packages
	./scripts/go-generate.sh

.PHONY: regen-mocks
regen-mocks: ## Delete and regenerate all mocks
	find . -type f | grep mocks/mock_ | xargs rm
	find . -type f | grep mocks/matchers | xargs rm
	@# not using $(PKG) here because that includes directories that have now
	@# been made empty, causing go generate to fail.
	./scripts/go-generate.sh

.PHONY: test
test: ## Run tests
	@go test -short $(PKG)

.PHONY: docker/test
docker/test: ## Run tests in docker
	docker run -it -v $(PWD):/atlantis ghcr.io/runatlantis/testing-env:latest sh -c "cd /atlantis && make test"

.PHONY: test-all
test-all: ## Run tests including integration
	@go test  $(PKG)

.PHONY: docker/test-all
docker/test-all: ## Run all tests in docker
	docker run -it -v $(PWD):/atlantis ghcr.io/runatlantis/testing-env:latest sh -c "cd /atlantis && make test-all"

.PHONY: test-coverage
test-coverage: ## Show test coverage
	@mkdir -p .cover
	@go test -covermode atomic -coverprofile .cover/cover.out $(PKG)

.PHONY: test-coverage-html
test-coverage-html: ## Show test coverage and output html
	@mkdir -p .cover
	@go test -covermode atomic -coverpkg $(PKG_COMMAS) -coverprofile .cover/cover.out $(PKG)
	go tool cover -html .cover/cover.out

.PHONY: docker/dev
docker/dev: ## Build dev Dockerfile as atlantis-dev
	GOOS=linux GOARCH=amd64 go build -o atlantis .
	docker build -f Dockerfile.dev -t atlantis-dev .

.PHONY: release
release: ## Create packages for a release
	docker run -v $$(pwd):/go/src/github.com/runatlantis/atlantis cimg/go:1.20 sh -c 'cd /go/src/github.com/runatlantis/atlantis && scripts/binary-release.sh'

.PHONY: fmt
fmt: ## Run goimports (which also formats)
	goimports -w $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "**/mocks/*")

.PHONY: lint
lint: ## Run linter locally
	golangci-lint run

.PHONY: check-lint
check-lint: ## Run linter in CI/CD. If running locally use 'lint'
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.49.0
	./bin/golangci-lint run -j 4 --timeout 5m

.PHONY: check-fmt
check-fmt: ## Fail if not formatted
	if [[ $$(goimports -l $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "**/mocks/*")) ]]; then exit 1; fi

.PHONY: end-to-end-deps
end-to-end-deps: ## Install e2e dependencies
	./scripts/e2e-deps.sh

.PHONY: end-to-end-tests
end-to-end-tests: ## Run e2e tests
	./scripts/e2e.sh

.PHONY: website-dev
website-dev: ## Run runatlantic.io on localhost:8080
	pnpm website:dev
