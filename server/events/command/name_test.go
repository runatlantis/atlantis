package command_test

import (
	"fmt"
	"math"
	"reflect"
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
		{command.State, "state"},
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
		{command.State, "state [rm ADDRESS...]"},
	}
	for _, tt := range tests {
		t.Run(tt.c.String(), func(t *testing.T) {
			if got := tt.c.DefaultUsage(); got != tt.want {
				t.Errorf("DefaultUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_SubCommands(t *testing.T) {
	tests := []struct {
		c    command.Name
		want []string
	}{
		{c: command.Apply},
		{c: command.Plan},
		{c: command.Unlock},
		{c: command.PolicyCheck},
		{c: command.ApprovePolicies},
		{c: command.Version},
		{c: command.Import},
		{c: command.State, want: []string{"rm"}},
	}
	for _, tt := range tests {
		t.Run(tt.c.String(), func(t *testing.T) {
			if got := tt.c.SubCommands(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SubCommands() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_CommandArgCount(t *testing.T) {
	tests := []struct {
		c          command.Name
		subCommand string
		want       *command.ArgCount
		wantErr    bool
	}{
		{c: command.Apply, want: &command.ArgCount{}},
		{c: command.Plan, want: &command.ArgCount{}},
		{c: command.Unlock, want: &command.ArgCount{}},
		{c: command.PolicyCheck, want: &command.ArgCount{}},
		{c: command.ApprovePolicies, want: &command.ArgCount{}},
		{c: command.Version, want: &command.ArgCount{}},
		{c: command.Import, want: &command.ArgCount{Min: 2, Max: 2}},
		{c: command.State, subCommand: "rm", want: &command.ArgCount{Min: 1, Max: -1}},
		{c: command.State, subCommand: "unknown", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s %s", tt.c, tt.subCommand), func(t *testing.T) {
			got, err := tt.c.CommandArgCount(tt.subCommand)
			if (err != nil) != tt.wantErr {
				t.Errorf("CommandArgCount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CommandArgCount() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestArgCount_IsMatchCount(t *testing.T) {
	type fields struct {
		Min int
		Max int
	}
	tests := []struct {
		name   string
		fields fields
		count  int
		want   bool
	}{
		{name: "[0,0] success", fields: fields{Min: 0, Max: 0}, count: 0, want: true},
		{name: "[0,0] failure", fields: fields{Min: 0, Max: 0}, count: 1, want: false},
		{name: "[1,1] success", fields: fields{Min: 1, Max: 1}, count: 1, want: true},
		{name: "[1,1] failure1", fields: fields{Min: 1, Max: 1}, count: 0, want: false},
		{name: "[1,1] failure2", fields: fields{Min: 1, Max: 1}, count: 2, want: false},
		{name: "[-inf,1] success1", fields: fields{Min: -1, Max: 1}, count: 0, want: true},
		{name: "[-inf,1] success2", fields: fields{Min: -1, Max: 1}, count: 1, want: true},
		{name: "[-inf,1] failure", fields: fields{Min: -1, Max: 1}, count: 2, want: false},
		{name: "[1,inf] success1", fields: fields{Min: 1, Max: -1}, count: 1, want: true},
		{name: "[1,inf] success2", fields: fields{Min: 1, Max: -1}, count: math.MaxInt, want: true},
		{name: "[1,inf] failure", fields: fields{Min: 1, Max: -1}, count: 0, want: false},
		{name: "[-inf,inf] success", fields: fields{Min: -1, Max: -1}, count: 0, want: true},
		{name: "[-inf,inf] success", fields: fields{Min: -1, Max: -1}, count: math.MaxInt, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := command.ArgCount{
				Min: tt.fields.Min,
				Max: tt.fields.Max,
			}
			if got := a.IsMatchCount(tt.count); got != tt.want {
				t.Errorf("IsMatchCount() = %v, want %v", got, tt.want)
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
		{command.State, "state"},
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
