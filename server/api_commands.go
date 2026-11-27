// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"slices"
	"strings"

	"github.com/runatlantis/atlantis/server/events/command"
)

func (u UserConfig) ToAllowAPICommandNames() ([]command.Name, error) {
	allowed := []command.Name{}
	input := u.AllowAPICommands
	if input == "" {
		return slices.Clone(command.DefaultAPICommands), nil
	}
	hasAll := false
	hasNone := false
	for item := range strings.SplitSeq(input, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		switch item {
		case "all":
			hasAll = true
			continue
		case "none":
			hasNone = true
			continue
		}
		name, err := command.ParseCommandName(item)
		if err != nil {
			return nil, err
		}
		if name != command.Plan && name != command.Apply && !name.IsInspection() {
			return nil, fmt.Errorf("unsupported API command %q", item)
		}
		allowed = append(allowed, name)
	}
	if hasNone {
		if hasAll || len(allowed) > 0 {
			return nil, fmt.Errorf("none must be used alone")
		}
		return allowed, nil
	}
	if hasAll {
		return []command.Name{command.Plan, command.Apply, command.Output, command.StateList, command.StateShow}, nil
	}
	return allowed, nil
}
