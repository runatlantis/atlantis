#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# Download dependencies
echo "Running 'make deps'"
make deps

# Test dependencies
echo "Running 'make deps-test'"
make deps-test

# Run tests
echo "Running 'make test'"
make test

# Build packages to make sure they can be compiled
echo "Running 'make build'"
make build

cat ~/.circlerc
echo "ATLANTIS_URL="${ATLANTIS_URL}

# Run e2e tests
echo "Running e2e test: 'make run'"
make run
