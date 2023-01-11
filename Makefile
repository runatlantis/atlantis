BUILD_ID := $(shell git rev-parse --short HEAD 2>/dev/null || echo no-commit-id)
WORKSPACE := $(shell pwd)
PKG := $(shell go list ./... | grep -v e2e | grep -v static | grep -v mocks | grep -v testing)
PKG_COMMAS := $(shell go list ./... | grep -v e2e | grep -v static | grep -v mocks | grep -v testing | tr '\n' ',')
IMAGE_NAME := runatlantis/atlantis

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
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o atlantis .

go-generate: ## Run go generate in all packages
	./scripts/go-generate.sh

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

test-all: ## Run tests including integration
	@go test  $(PKG)

.PHONY: docker/test-all
docker/test-all: ## Run all tests in docker
	docker run -it -v $(pwd):/atlantis ghcr.io/runatlantis/testing-env:latest sh -c "cd /atlantis && make test-all"

test-coverage:
	@mkdir -p .cover
	@go test -covermode atomic -coverprofile .cover/cover.out $(PKG)

test-coverage-html:
	@mkdir -p .cover
	@go test -covermode atomic -coverpkg $(PKG_COMMAS) -coverprofile .cover/cover.out $(PKG)
	go tool cover -html .cover/cover.out

dev-docker:
	GOOS=linux GOARCH=amd64 go build -o atlantis .
	docker build -f Dockerfile.dev -t atlantis-dev .

dist: ## Package static/ using go-bindata-assetfs into a single binary
	rm -f server/static/bindata_assetfs.go && go-bindata-assetfs -o bindata_assetfs.go -pkg static -prefix server server/static/... && mv bindata_assetfs.go server/static

release: ## Create packages for a release
	docker run -v $$(pwd):/go/src/github.com/runatlantis/atlantis cimg/go:1.19 sh -c 'cd /go/src/github.com/runatlantis/atlantis && scripts/binary-release.sh'

fmt: ## Run goimports (which also formats)
	goimports -w $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")

lint: ## Run linter locally
	golangci-lint run

check-lint: ## Run linter in CI/CD. If running locally use 'lint'
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b ./bin v1.49.0
	./bin/golangci-lint run -j 4 --timeout 5m

check-fmt: ## Fail if not formatted
	if [[ $$(goimports -l $$(find . -type f -name '*.go' ! -path "./vendor/*" ! -path "./server/static/bindata_assetfs.go" ! -path "**/mocks/*")) ]]; then exit 1; fi

end-to-end-deps: ## Install e2e dependencies
	./scripts/e2e-deps.sh

end-to-end-tests: ## Run e2e tests
	./scripts/e2e.sh

website-dev: ## Run runatlantic.io on localhost:8080
	yarn website:dev
