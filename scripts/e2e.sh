#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

# download all the tooling needed for e2e tests
CIRCLE_WORKING_DIRECTORY="${CIRCLE_WORKING_DIRECTORY/#\~/$HOME}" # https://discuss.circleci.com/t/circle-working-directory-doesnt-expand/17007/5
${CIRCLE_WORKING_DIRECTORY}/scripts/e2e-deps.sh
cd "${CIRCLE_WORKING_DIRECTORY}/e2e"

# start atlantis server in the background and wait for it to start
./atlantis server \
  --gh-user="$GITHUB_USERNAME" \
  --gh-token="$GITHUB_PASSWORD" \
  --data-dir="/tmp" \
  --log-level="debug" \
  --repo-allowlist="github.com/runatlantis/atlantis-tests" \
  --repo-config-json='{"repos":[{"id":"/.*/", "allowed_overrides":["apply_requirements","workflow"], "allow_custom_workflows":true}]}' \
  &> /tmp/atlantis-server.log &
sleep 2

# start ngrok in the background and wait for it to start
./ngrok http 4141 > /tmp/ngrok.log &
sleep 2

# find out what URL ngrok has given us
export ATLANTIS_URL=$(curl -s 'http://localhost:4040/api/tunnels' | jq -r '.tunnels[] | select(.proto=="http") | .public_url')

# Now we can start the e2e tests
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
