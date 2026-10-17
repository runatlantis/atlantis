// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// Package valid contains definitions of valid yaml configuration after its
// been parsed and validated.
package valid

const DefaultAutoPlanEnabled = true

// DefaultGroup is the group a project belongs to when the repo config doesn't
// set one. Projects can be planned/applied a group at a time with the
// -g/--group flag.
const DefaultGroup = "default"
