#!/bin/bash -eu
# Copyright 2025 The Atlantis Authors
# SPDX-License-Identifier: Apache-2.0

compile_native_go_fuzzer github.com/runatlantis/atlantis/server/core/config/cfgfuzz FuzzParseRepoCfgData FuzzParseRepoCfgData
