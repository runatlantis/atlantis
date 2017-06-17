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

    # Set ATLANTIS_URL environment variable to be used by atlantis e2e test to create the webhook
echo 'export ATLANTIS_URL=$(curl -s 'http://localhost:4040/api/tunnels' | jq -r '.tunnels[1].public_url')' >> ~/.circlerc
cat ~/.circlerc
echo "ATLANTIS_URL="${ATLANTIS_URL}

# Run e2e tests
echo "Running e2e test: 'make run'"
make run
