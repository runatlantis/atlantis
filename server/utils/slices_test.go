// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package utils_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/utils"
	. "github.com/runatlantis/atlantis/testing"
)

func TestSlicesContains_String(t *testing.T) {
	cases := []struct {
		Name string
		In   []string
		Find string
		Want bool
	}{
		{
			Name: "present",
			In:   []string{"atlantis", "terraform", "opentofu"},
			Find: "terraform",
			Want: true,
		},
		{
			Name: "absent",
			In:   []string{"atlantis", "terraform", "opentofu"},
			Find: "helm",
			Want: false,
		},
		{
			Name: "nil slice",
			In:   nil,
			Find: "terraform",
			Want: false,
		},
	}

	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			Equals(t, c.Want, utils.SlicesContains(c.In, c.Find))
		})
	}
}

func TestSlicesContains_Int(t *testing.T) {
	Equals(t, true, utils.SlicesContains([]int{1, 2, 3}, 2))
	Equals(t, false, utils.SlicesContains([]int{1, 2, 3}, 4))
}
