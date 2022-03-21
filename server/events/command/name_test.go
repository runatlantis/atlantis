package command_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	. "github.com/runatlantis/atlantis/testing"
)

func TestApplyCommand_String(t *testing.T) {
	uc := command.Apply

	Equals(t, "apply", uc.String())
}

func TestPlanCommand_String(t *testing.T) {
	uc := command.Plan

	Equals(t, "plan", uc.String())
}

func TestPolicyCheckCommand_String(t *testing.T) {
	uc := command.PolicyCheck

	Equals(t, "policy_check", uc.String())
}

func TestUnlockCommand_String(t *testing.T) {
	uc := command.Unlock

	Equals(t, "unlock", uc.String())
}
