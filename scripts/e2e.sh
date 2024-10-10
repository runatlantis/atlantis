#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
ATLANTIS_PID=""
NGROK_PID=""

function cleanup() {
    cleanupPid "$ATLANTIS_PID"
    cleanupPid "$NGROK_PID"
}

function cleanupPid() {
    local pid="$1"
    # Never set, no need to clean up
    if [[ "$pid" == "" ]]
    then
        return
    fi
    # Somehow pid was not number, just being careful
    if ! [[ "$pid" =~ ^[0-9]+$ ]]
    then
        return
    fi
    # Not currently running, no need to kill
    if ! ps -p "$pid" &>/dev/null
    then
        return
    fi
    kill $pid
}


# start atlantis server in the background and wait for it to start
# It's the responsibility of the caller of this script to set the github, gitlab, etc.
# permissions via environment variable
./atlantis server \
  --data-dir="/tmp" \
  --log-level="debug" \
  --repo-allowlist="github.com/runatlantis/atlantis-tests,gitlab.com/run-atlantis/atlantis-tests" \
  --repo-config-json='{"repos":[{"id":"/.*/", "allowed_overrides":["apply_requirements","workflow"], "allow_custom_workflows":true}]}' \
  &> /tmp/atlantis-server.log &
ATLANTIS_PID=$!
sleep 2
if ! ps -p "$ATLANTIS_PID" &>/dev/null
then
    echo "Atlantis failed to start"
    cat /tmp/atlantis-server.log
    exit 1
fi
echo "Atlantis is running..."

# start ngrok in the background and wait for it to start
./ngrok config add-authtoken $NGROK_AUTH_TOKEN > /dev/null 2>&1
./ngrok http 4141 > /tmp/ngrok.log 2>&1 &
NGROK_PID=$!
sleep 2
if ! ps -p "$NGROK_PID" &>/dev/null
then
    cleanup
    echo "Ngrok failed to start"
    cat /tmp/ngrok.log
    exit 1
fi
echo "Ngrok is running..."

# find out what URL ngrok has given us
export ATLANTIS_URL=$(curl -s 'http://localhost:4040/api/tunnels' | jq -r '.tunnels[] | select(.proto=="https") | .public_url')

# Now we can start the e2e tests
cd "${GITHUB_WORKSPACE:-$(git rev-parse --show-toplevel)}/e2e"
echo "Running 'make build'"
make build

echo "Running e2e test: 'make run'"
set +e
estatus=0
make run
if [[ $? -eq 0 ]]
then
  echo "e2e tests passed"
else
  echo "e2e tests failed"
  echo "atlantis logs:"
  cat /tmp/atlantis-server.log
  estatus=1
fi
cleanup
exit $estatus
