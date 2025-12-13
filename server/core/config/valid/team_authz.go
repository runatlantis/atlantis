// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid

type TeamAuthz struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args" json:"args"`
}
