#!/usr/bin/env bash

set -euo pipefail

go install golang.org/x/tools/cmd/goimports@latest

gobin="$(go env GOPATH)/bin"
declare -r gobin

declare -a files
readarray -d '' files < <(find . -type f -name '*.go' ! -name 'mock_*' ! -path './vendor/*' ! -path '**/mocks/*' -print0)
declare -r files

output="$("${gobin}"/goimports -l "${files[@]}")"
declare -r output

if [[ -n "$output" ]]; then
    echo "These files had their 'import' changed - please fix them locally and push a fix"

    echo "$output"

    exit 1
fi
