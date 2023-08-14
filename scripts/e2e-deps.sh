#!/usr/bin/env bash

# Exit immediately if a command returns a non-zero code
set -e

echo "Preparing to run e2e tests"
if [ ! -f atlantis ]; then
    echo "atlantis binary not found. exiting...."
    exit 1
fi
cp atlantis ${CIRCLE_WORKING_DIRECTORY}/e2e/

# cd into e2e folder
cd e2e/
# Download terraform
curl -LOk https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip
chmod +x terraform
cp terraform /home/circleci/go/bin
# Download ngrok to create a tunnel to expose atlantis server
wget https://bin.equinox.io/c/bNyj1mQVY4c/ngrok-v2-stable-linux-amd64.zip -O ngrok-stable-linux-amd64.zip
unzip ngrok-stable-linux-amd64.zip
chmod +x ngrok
# Download jq
wget -O jq https://github.com/jqlang/jq/releases/download/jq-1.6/jq-linux64
chmod +x jq
# Copy github config file - replace with circleci user later
cp .gitconfig ~/.gitconfig
