#!/bin/sh
# Taken and modified from https://github.com/mlafeldt/chef-runner/blob/v0.7.0/script/coverage
# Generate test coverage statistics for Go packages.
#
# Works around the fact that `go test -coverprofile` currently does not work
# with multiple packages, see https://code.google.com/p/go/issues/detail?id=6909
#
# Usage: coverage.sh packages...
# Example: coverage.sh github.com/runatlantis/atlantis github.com/runatlantis/atlantis/testdrive
#

set -e

workdir=.cover
profile="$workdir/cover.out"
mode=count

generate_cover_data() {
    rm -rf "$workdir"
    mkdir "$workdir"

    pkgs=$@
    for pkg in $pkgs; do
        f="$workdir/$(echo $pkg | tr / -).cover"
        go test -covermode="$mode" -coverprofile="$f" "$pkg"
    done

    echo "mode: $mode" >"$profile"
    grep -h -v "^mode:" "$workdir"/*.cover >>"$profile"
}

generate_cover_data $@
