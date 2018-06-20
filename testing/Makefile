TEST_IMAGE_NAME := runatlantis/testing-env

.DEFAULT_GOAL := help
help: ## List targets & descriptions
	@cat Makefile* | grep -E '^[a-zA-Z_-]+:.*?## .*$$' | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build-testing-image: ## Build and push the testing image
	docker build -t $(TEST_IMAGE_NAME):$$(git rev-parse HEAD) .
	docker tag $(TEST_IMAGE_NAME):$$(git rev-parse HEAD) $(TEST_IMAGE_NAME):latest
	docker push $(TEST_IMAGE_NAME):$$(git rev-parse HEAD)
	docker push $(TEST_IMAGE_NAME):latest
