#!/usr/bin/env bash
# Copyright 2025 The Atlantis Authors
# SPDX-License-Identifier: Apache-2.0


set -euo pipefail

if [[ "${1:-}" == "--check" ]]; then
  echo "checking SPDX headers..."
  MODE="-check"
else
  echo "adding/updating SPDX headers..."
  MODE=""
fi

addlicense $MODE \
  -s=only \
  -c "The Atlantis Authors" \
  $(find . -name '*.go' | grep -v _mock)
