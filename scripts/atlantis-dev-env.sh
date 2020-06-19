#!/bin/bash
AWS_PROFILE=$1
TOKEN=$2
REPO_WHITELIST=$3
DEFAULT_WORKFLOW=$4
WEBHOOK=$5

function ngrok_start(){
    type ngrok >/dev/null 2>&1 || { echo >&2 "I require ngrok command but it's not installed.  Aborting."; exit 1; }
    ngrok http -log=stdout 4141 > /dev/null &
}

function get_ngrok_endpoint(){
    NGROK_ENPOINT=$(curl -s http://localhost:4040/api/tunnels|jq .tunnels[0].public_url)
    ENDPOINT=$(echo $NGROK_ENPOINT|sed 's/"//g')"/events"
}

function get_token_parameter(){
    AUTH=$(echo "Authorization: token "$TOKEN"")
}

function docker_build(){
    docker build -t atlantis-test .
}

function create_random_password(){
    SECRET=$(cat /dev/urandom | env LC_CTYPE=C tr -dc '0-9' | fold -w 10 | head -n 1)
}

function update_hook(){
    sleep 5
    get_ngrok_endpoint
    get_token_parameter
    create_random_password
    DATA=$(echo {\"active\": true, \"config\":{ \"secret\": \""$SECRET"\", \"url\": \""$ENDPOINT"\",\"content_type\": \"json\"}})
    URL=$WEBHOOK
    echo $AUTH $URL $DATA
    curl -s -L -H "$AUTH" "$URL" --request PATCH -d "$DATA"
}

function docker_run(){
    sleep 5
    docker run -v $HOME/.aws:/home/atlantis/.aws -e AWS_PROFILE=$AWS_PROFILE -p 4141:4141 atlantis-test:latest server  \
    --gh-user=letgo-atlantis \
    --gh-token=$TOKEN \
    --atlantis-url=$ENDPOINT \
    --gh-webhook-secret=$SECRET \
    --repo-whitelist=$REPO_WHITELIST \
    --repo-config=$DEFAULT_WORKFLOW
}

function tear_down(){
    killall ngrok
}

ngrok_start
update_hook
docker_build
docker_run
trap tear_down 1 2 3 9 15
