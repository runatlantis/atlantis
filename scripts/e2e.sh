#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

cd e2e/

# Download dependencies
echo "Running 'make deps'"
make deps

# Run tests
echo "Running 'make test'"
make test

# Build packages to make sure they can be compiled
echo "Running 'make build'"
make build

# Run e2e tests
echo "Running e2e test: 'make run'"
make run
