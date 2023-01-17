package command

import (
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
	// TODO: remove ApprovePolicies with policy v2
	ApprovePolicies
	// Autoplan is a command to run terrafor plan on PR open/update if autoplan is enabled
	Autoplan
	// Version is a command to run terraform version.
	Version
	// Adding more? Don't forget to update String() below
)

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
	}
	return ""
}
