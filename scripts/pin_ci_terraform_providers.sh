#!/bin/bash

# Script to pin terraform providers in e2e tests

RANDOM_PROVIDER_VERSION="3.6.1"
NULL_PROVIDER_VERSION="3.2.3"

TEST_REPOS_DIR="server/controllers/events/testdata/test-repos"

for file in $(find $TEST_REPOS_DIR -name '*.tf')
do
    basename=$(basename $file)
    if [[ "$basename" == "versions.tf" ]]
    then
        continue
    fi
    if [[ "$basename" != "main.tf" ]]
    then
        echo "Found unexpected file: $file"
        exit 1
    fi
    has_null_provider=false
    has_random_provider=false

    version_file="$(dirname $file)/versions.tf"
    for resource in $(cat $file | grep '^resource' | awk '{print $2}' | tr -d '"')
    do
        if [[ "$resource" == "null_resource" ]]
        then
            has_null_provider=true
        elif [[ "$resource" == "random_id" ]]
        then
            has_random_provider=true
        else
            echo "Unknown resource $resource in $file"
            exit 1
        fi
    done
    if ! $has_null_provider && ! $has_random_provider
    then
        echo "No providers needed for $file"
        continue
    fi
    echo "Adding $version_file for $file"
    rm -f $version_file
    if $has_null_provider
    then
        echo 'provider "null" {}' >> $version_file
    fi
    if $has_random_provider
    then
        echo 'provider "random" {}' >> $version_file
    fi
    echo "terraform {" >> $version_file
    echo "  required_providers {" >> $version_file

    if $has_random_provider
    then
        echo "    random = {" >> $version_file
        echo '      source  = "hashicorp/random"' >> $version_file
        echo "      version = \"= $RANDOM_PROVIDER_VERSION\"" >> $version_file
        echo "    }" >> $version_file
    fi
    if $has_null_provider
    then
        echo "    null = {" >> $version_file
        echo '      source  = "hashicorp/null"' >> $version_file
        echo "      version = \"= $NULL_PROVIDER_VERSION\"" >> $version_file
        echo "    }" >> $version_file
    fi
    echo "  }" >> $version_file
    echo "}" >> $version_file

done
