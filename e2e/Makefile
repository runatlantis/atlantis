WORKSPACE := $(shell pwd)

.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

debug: ## Output internal make variables
	@echo WORKSPACE = $(WORKSPACE)

deps: ## Get go dependencies
	go get -v .

build: ## Build the main Go service
	go build -v -o atlantis-tests .

deps-test: ## Run tests for go dependencies
	go get -t

test: ## Run tests, coverage reports, and clean (coverage taints the compiled code)
	go test -v .

run: ## Run e2e tests
	./atlantis-tests
