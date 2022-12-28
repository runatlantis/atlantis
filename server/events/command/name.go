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
	// Adding more? Don't forget to update String() below
)

type ArgCount struct {
	Min int
	Max int
}

// AllCommentCommands are list of commands that can be run from a comment.
var AllCommentCommands = []Name{
	Version,
	Plan,
	Apply,
	Unlock,
	ApprovePolicies,
	Import,
	State,
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
	}
	return ""
}

// DefaultUsage returns the command default usage
func (c Name) DefaultUsage() string {
	switch c {
	case Import:
		return "import ADDRESS ID"
	case State:
		return "state [rm ADDRESS...]"
	default:
		return c.String()
	}
}

// SubCommands returns the list of sub commands for the command
func (c Name) SubCommands() []string {
	switch c {
	case State:
		return []string{"rm"}
	default:
		return nil
	}
}

// CommandArgCount returns the number of required arguments for the command
func (c Name) CommandArgCount(subCommand string) (*ArgCount, error) {
	switch c {
	case Import:
		return &ArgCount{2, 2}, nil // "atlantis import ADDRESS ID"
	case State:
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
	}
	return -1, fmt.Errorf("unknown command name: %s", name)
}
