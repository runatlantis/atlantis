#!/usr/bin/env bash
# Verify that the Go toolchain used to build the container images matches the
# version required by go.mod.
#
# The `go` directive in go.mod is maintained by Renovate's gomod manager, while
# the golang base images are pinned separately. When the two drift, the image
# build fails with "go.mod requires go >= X (running go Y; GOTOOLCHAIN=local)",
# and only when the Docker layer cache misses, which can be many commits after
# the drift was introduced.
set -euo pipefail

fail=0

required="$(awk '/^go [0-9]/ {print $2; exit}' go.mod)"
if [[ -z "${required}" ]]; then
  echo "could not read the go directive from go.mod" >&2
  exit 1
fi

check() {
  local file="$1" found="$2"
  if [[ -z "${found}" ]]; then
    echo "FAIL ${file}: could not find a golang version to check" >&2
    fail=1
  elif [[ "${found}" != "${required}" ]]; then
    echo "FAIL ${file}: pins go ${found}, but go.mod requires ${required}" >&2
    fail=1
  else
    echo "ok   ${file}: go ${found}"
  fi
}

# Dockerfile: ARG GOLANG_TAG=<version>-alpine<x.y>@sha256:...
check Dockerfile \
  "$(sed -n 's/^ARG GOLANG_TAG=\([0-9][0-9.]*\)-alpine.*/\1/p' Dockerfile | head -1)"

# testing/Dockerfile: FROM golang:<version>@sha256:...
check testing/Dockerfile \
  "$(sed -n 's/^FROM golang:\([0-9][0-9.]*\)[@ ].*/\1/p' testing/Dockerfile | head -1)"

# The nested modules must not require a newer toolchain than the images provide.
for mod in e2e/go.mod .github/scripts/update_top_issues_ranking/go.mod; do
  [[ -f "${mod}" ]] || continue
  nested="$(awk '/^go [0-9]/ {print $2; exit}' "${mod}")"
  if [[ "${nested}" != "${required}" ]]; then
    echo "FAIL ${mod}: requires go ${nested}, but go.mod requires ${required}" >&2
    fail=1
  else
    echo "ok   ${mod}: go ${nested}"
  fi
done

if [[ "${fail}" -ne 0 ]]; then
  echo >&2
  echo "Go toolchain pins are out of sync. Update the golang image pins to ${required}." >&2
  exit 1
fi
