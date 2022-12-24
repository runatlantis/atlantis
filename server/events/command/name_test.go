package command_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/stretchr/testify/assert"
)

func TestName_TitleString(t *testing.T) {
	tests := []struct {
		c    command.Name
		want string
	}{
		{command.Apply, "Apply"},
		{command.PolicyCheck, "Policy Check"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.c.TitleString(); got != tt.want {
				t.Errorf("TitleString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_String(t *testing.T) {
	tests := []struct {
		c    command.Name
		want string
	}{
		{command.Apply, "apply"},
		{command.Plan, "plan"},
		{command.Unlock, "unlock"},
		{command.PolicyCheck, "policy_check"},
		{command.ApprovePolicies, "approve_policies"},
		{command.Version, "version"},
		{command.Import, "import"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_DefaultUsage(t *testing.T) {
	tests := []struct {
		c    command.Name
		want string
	}{
		{command.Apply, "apply"},
		{command.Plan, "plan"},
		{command.Unlock, "unlock"},
		{command.PolicyCheck, "policy_check"},
		{command.ApprovePolicies, "approve_policies"},
		{command.Version, "version"},
		{command.Import, "import ADDRESS ID"},
	}
	for _, tt := range tests {
		t.Run(tt.c.String(), func(t *testing.T) {
			if got := tt.c.DefaultUsage(); got != tt.want {
				t.Errorf("DefaultUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCommandName(t *testing.T) {
	tests := []struct {
		exp  command.Name
		name string
	}{
		{command.Apply, "apply"},
		{command.Plan, "plan"},
		{command.Unlock, "unlock"},
		{command.PolicyCheck, "policy_check"},
		{command.ApprovePolicies, "approve_policies"},
		{command.Version, "version"},
		{command.Import, "import"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := command.ParseCommandName(tt.name)
			assert.NoError(t, err)
			assert.Equal(t, tt.exp, got)
		})
	}

	t.Run("unknown command", func(t *testing.T) {
		_, err := command.ParseCommandName("unknown")
		assert.ErrorContains(t, err, "unknown command name: unknown")
	})
}
