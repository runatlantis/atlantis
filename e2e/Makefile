WORKSPACE := $(shell pwd)

.PHONY: test

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

debug: ## Output internal make variables
	@echo WORKSPACE = $(WORKSPACE)

build: ## Build the main Go service
	rm -f atlantis-tests
	go build -v -o atlantis-tests .

run: ## Run e2e tests
	./atlantis-tests
