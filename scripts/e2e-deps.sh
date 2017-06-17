#!/usr/bin/env bash

echo "Preparing to run e2e tests"
mv atlantis ${WORKDIR}/e2e/

# cd into e2e folder
cd e2e/
# Encrypting secrets for atlantis runtime: https://github.com/circleci/encrypted-files
openssl aes-256-cbc -d -in secrets-envs -k $KEY >> ~/.circlerc
# Download terraform
curl -LOk https://releases.hashicorp.com/terraform/${TERRAFORM_VERSION}/terraform_${TERRAFORM_VERSION}_linux_amd64.zip
unzip terraform_${TERRAFORM_VERSION}_linux_amd64.zip -d /home/ubuntu/bin
# Download ngrok to create a tunnel to expose atlantis server
wget https://bin.equinox.io/c/4VmDzA7iaHb/ngrok-stable-linux-amd64.zip
unzip ngrok-stable-linux-amd64.zip 
chmod +x ngrok 
wget https://stedolan.github.io/jq/download/linux64/jq
chmod +x jq
# Copy github config file - replace with circleci user later
cp .gitconfig ~/.gitconfig