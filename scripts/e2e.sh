#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# start atlantis server in the background and wait for it to start
./atlantis server \
  --data-dir="/tmp" \
  --log-level="debug" \
  --repo-allowlist="github.com/runatlantis/atlantis-tests" \
  --repo-config-json='{"repos":[{"id":"/.*/", "allowed_overrides":["apply_requirements","workflow"], "allow_custom_workflows":true}]}' \
  &> /tmp/atlantis-server.log &
sleep 2

# start ngrok in the background and wait for it to start
./ngrok config add-authtoken $NGROK_AUTH_TOKEN > /dev/null 2>&1
./ngrok http 4141 > /tmp/ngrok.log 2>&1 &
sleep 2

# find out what URL ngrok has given us
export ATLANTIS_URL=$(curl -s 'http://localhost:4040/api/tunnels' | jq -r '.tunnels[] | select(.proto=="https") | .public_url')

# Now we can start the e2e tests
cd "${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}/e2e"
echo "Running 'make build'"
make build

echo "Running e2e test: 'make run'"
set +e
make run
if [[ $? -eq 0 ]]
then
  echo "e2e tests passed"
else
  echo "e2e tests failed"
  echo "atlantis logs:"
  cat /tmp/atlantis-server.log
  exit 1
fi
