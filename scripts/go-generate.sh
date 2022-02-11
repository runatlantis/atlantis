#!/bin/bash

set -eou pipefail

pkgs=$(go list ./... | grep -v mocks | grep -v matchers | grep -v e2e | grep -v static)
for pkg in $pkgs; do
	go generate -v "$pkg"
done
