// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAllowAPICommands(t *testing.T) {
	for _, tc := range []struct {
		input   string
		want    []command.Name
		invalid bool
	}{
		{"", []command.Name{command.Plan, command.Apply}, false},
		{"plan,apply", []command.Name{command.Plan, command.Apply}, false},
		{"plan,apply,", []command.Name{command.Plan, command.Apply}, false},
		{" plan, , apply ", []command.Name{command.Plan, command.Apply}, false},
		{" , ", []command.Name{}, false},
		{" none, , ", []command.Name{}, false},
		{" all, , ", []command.Name{command.Plan, command.Apply, command.Output, command.StateList, command.StateShow}, false},
		{"none,all", nil, true},
		{"output,state_list,state_show", []command.Name{command.Output, command.StateList, command.StateShow}, false},
		{"all", []command.Name{command.Plan, command.Apply, command.Output, command.StateList, command.StateShow}, false},
		{"none", []command.Name{}, false},
		{"state", nil, true},
		{"import", nil, true},
		{"all,invalid", nil, true},
		{"none,output", nil, true},
	} {
		t.Run(tc.input, func(t *testing.T) {
			got, err := (server.UserConfig{AllowAPICommands: tc.input}).ToAllowAPICommandNames()
			if tc.invalid {
				Assert(t, err != nil, "expected validation error")
				return
			}
			Ok(t, err)
			Equals(t, tc.want, got)
		})
	}
}
