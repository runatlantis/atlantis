#!/usr/bin/env bash

# Exit immediately if a command returns a non-zero code
set -e

echo "Preparing to run e2e tests"
if [ ! -f atlantis ]; then
    echo "atlantis binary not found. exiting...."
    exit 1
fi
cp atlantis ${GITHUB_WORKSPACE}/e2e/
echo "${GITHUB_WORKSPACE}/e2e" >> $GITHUB_PATH

# cd into e2e folder
cd e2e/

# Download terraform
curl -sLOk https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip
chmod +x terraform

# Download ngrok to create a tunnel to expose atlantis server
wget https://bin.equinox.io/c/4VmDzA7iaHb/ngrok-stable-linux-amd64.zip
unzip ngrok-stable-linux-amd64.zip 
chmod +x ngrok

# Download jq
wget https://stedolan.github.io/jq/download/linux64/jq
chmod +x jq

# Copy github config file
cp .gitconfig ~/.gitconfig
