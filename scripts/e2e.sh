#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# download all the tooling needed for e2e tests
${CIRCLE_WORKING_DIRECTORY}/scripts/e2e-deps.sh
cd "${CIRCLE_WORKING_DIRECTORY}/e2e"

# start atlantis server in the background and wait for it to start
./atlantis server --gh-user="$GITHUB_USERNAME" --gh-token="$GITHUB_PASSWORD" --data-dir="/tmp" --log-level="debug" --repo-whitelist="github.com/runatlantis/atlantis-tests" &> /tmp/atlantis-server.log &
sleep 2

# start ngrok in the background and wait for it to start
./ngrok http 4141 > /tmp/ngrok.log &
sleep 2

# find out what URL ngrok has given us
export ATLANTIS_URL=$(curl -s 'http://localhost:4040/api/tunnels' | jq -r '.tunnels[] | select(.proto=="http") | .public_url')

# Now we can start the e2e tests
echo "Running 'make deps'"
make deps

echo "Running 'make test'"
make test

echo "Running 'make build'"
make build

echo "Running e2e test: 'make run'"
make run
