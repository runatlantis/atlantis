#!/bin/bash -eu
# Copyright 2025 The Atlantis Authors
# SPDX-License-Identifier: Apache-2.0

# Register go-118-fuzz-build so compile_native_go_fuzzer can inject its harness.
printf "package main\nimport _ \"github.com/AdamKorcz/go-118-fuzz-build/testing\"\n" > /tmp/register.go
cp /tmp/register.go "$SRC/atlantis/register.go"
cd "$SRC/atlantis" && go get github.com/AdamKorcz/go-118-fuzz-build && go mod tidy

compile_native_go_fuzzer github.com/runatlantis/atlantis/server/core/config/cfgfuzz FuzzParseRepoCfgData FuzzParseRepoCfgData
