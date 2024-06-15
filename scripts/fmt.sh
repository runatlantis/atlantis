#!/usr/bin/env bash

set -euo pipefail

declare -a files
readarray -d '' files < <(find . -type f -name '*.go' ! -name 'mock_*' ! -path './vendor/*' ! -path '**/mocks/*' -print0)
declare -r files

# Run go fmt on the files
gofmt -s -w "${files[@]}"

# Check for import changes using grep
changed_files=()
for file in "${files[@]}"; do
    if grep -q "^import (" "$file"; then
        changed_files+=("$file")
    fi
done

if [[ ${#changed_files[@]} -gt 0 ]]; then
    echo "These files had their 'import' changed - please fix them locally and push a fix"

    for file in "${changed_files[@]}"; do
        echo "$file"
    done

    exit 1
fi
