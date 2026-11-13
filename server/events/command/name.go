// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Name is which command to run.
type Name int

const (
	// Apply is a command to run terraform apply.
	Apply Name = iota
	// Plan is a command to run terraform plan.
	Plan
	// Unlock is a command to discard previous plans as well as the atlantis locks.
	Unlock
	// PolicyCheck is a command to run conftest test.
	PolicyCheck
	// ApprovePolicies is a command to approve policies with owner check
	ApprovePolicies
	// Autoplan is a command to run terraform plan on PR open/update if autoplan is enabled
	Autoplan
	// Version is a command to run terraform version.
	Version
	// Import is a command to run terraform import
	Import
	// State is a command to run terraform state rm
	State
	// Cancel is a command to cancel running plan or apply operations
	Cancel
	// Output inspects root module outputs.
	Output
	// StateList lists managed resource addresses.
	StateList
	// StateShow inspects a managed resource.
	StateShow
	// Adding more? Don't forget to update String() below
)

type ArgCount struct {
	Min int
	Max int
}

// DefaultAPICommands preserves access to the existing command API endpoints.
var DefaultAPICommands = []Name{Plan, Apply}

// AllCommentCommands are list of commands that can be run from a comment.
var AllCommentCommands = []Name{
	Version,
	Plan,
	Apply,
	Cancel,
	Unlock,
	ApprovePolicies,
	Import,
	State,
	Output,
	StateList,
	StateShow,
}

// TitleString returns the string representation in title form.
// ie. policy_check becomes Policy Check
func (c Name) TitleString() string {
	return cases.Title(language.English).String(strings.ReplaceAll(strings.ToLower(c.String()), "_", " "))
}

// String returns the string representation of c.
func (c Name) String() string {
	switch c {
	case Apply:
		return "apply"
	case Plan, Autoplan:
		return "plan"
	case Unlock:
		return "unlock"
	case PolicyCheck:
		return "policy_check"
	case ApprovePolicies:
		return "approve_policies"
	case Version:
		return "version"
	case Import:
		return "import"
	case State:
		return "state"
	case Cancel:
		return "cancel"
	case Output:
		return "output"
	case StateList:
		return "state_list"
	case StateShow:
		return "state_show"
	}
	return ""
}

// DefaultUsage returns the command default usage
func (c Name) DefaultUsage() string {
	switch c {
	case Import:
		return "import ADDRESS ID"
	case State:
		return "state [rm ADDRESS... | list [ADDRESS...] | show ADDRESS]"
	case Output:
		return "output [NAME]"
	case StateList:
		return "state list [ADDRESS...]"
	case StateShow:
		return "state show ADDRESS"
	default:
		return c.String()
	}
}

// SubCommands returns the list of sub commands for the command
func (c Name) SubCommands() []string {
	switch c {
	case State:
		return []string{"rm", "list", "show"}
	default:
		return nil
	}
}

// CommandArgCount returns the number of required arguments for the command
func (c Name) CommandArgCount(subCommand string) (*ArgCount, error) {
	switch c {
	case Output:
		return &ArgCount{0, 1}, nil
	case StateList:
		return &ArgCount{0, -1}, nil
	case StateShow:
		return &ArgCount{1, 1}, nil
	case Import:
		return &ArgCount{2, 2}, nil // "atlantis import ADDRESS ID"
	case State:
		if subCommand == "list" {
			return &ArgCount{0, -1}, nil
		}
		if subCommand == "show" {
			return &ArgCount{1, 1}, nil
		}
		if subCommand == "rm" {
			return &ArgCount{1, -1}, nil // "atlantis state rm ADDRESS..."
		}
		return nil, fmt.Errorf("command arg count unknown sub command: %s", subCommand)
	default:
		return &ArgCount{0, 0}, nil // other command doesn't require any args
	}
}

// IsMatchCount returns true if the number of arguments matches the requirement
func (a ArgCount) IsMatchCount(count int) bool {
	if a.Min != -1 {
		if count < a.Min {
			return false
		}
	}
	if a.Max != -1 {
		if count > a.Max {
			return false
		}
	}
	return true
}

// ParseCommandName parses raw name into a command name.
func ParseCommandName(name string) (Name, error) {
	switch name {
	case "apply":
		return Apply, nil
	case "plan":
		return Plan, nil
	case "unlock":
		return Unlock, nil
	case "policy_check":
		return PolicyCheck, nil
	case "approve_policies":
		return ApprovePolicies, nil
	case "version":
		return Version, nil
	case "import":
		return Import, nil
	case "state":
		return State, nil
	case "cancel":
		return Cancel, nil
	case "output":
		return Output, nil
	case "state_list":
		return StateList, nil
	case "state_show":
		return StateShow, nil
	}
	return -1, fmt.Errorf("unknown command name: %s", name)
}

func (c Name) IsInspection() bool {
	return c == Output || c == StateList || c == StateShow
}

// ValidateInspectionArgs confines inspection to the selected project's backend.
func ValidateInspectionArgs(c Name, args []string) error {
	if !c.IsInspection() {
		return fmt.Errorf("%s is not an inspection command", c)
	}
	positional := 0
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") {
			positional++
			continue
		}
		if positional > 0 {
			return fmt.Errorf("terraform flags must precede addresses or output names")
		}
		flag, _, hasValue := strings.Cut(strings.TrimLeft(arg, "-"), "=")
		switch {
		case flag == "no-color" && !hasValue:
		case c == Output && (flag == "json" || flag == "raw") && (!hasValue || strings.HasSuffix(arg, "=true") || strings.HasSuffix(arg, "=false")):
		case c == StateList && flag == "id":
			if !hasValue {
				i++
				if i >= len(args) {
					return fmt.Errorf("flag -id requires a value")
				}
			}
		default:
			return fmt.Errorf("flag %q is not allowed for %s", arg, c)
		}
	}
	count, _ := c.CommandArgCount("")
	if !count.IsMatchCount(positional) {
		return fmt.Errorf("usage: %s", c.DefaultUsage())
	}
	return nil
}
