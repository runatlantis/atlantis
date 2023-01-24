package server_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

func TestUserConfig_ToAllowCommandNames(t *testing.T) {
	tests := []struct {
		name          string
		allowCommands string
		want          []command.Name
		wantErr       string
	}{
		{
			name:          "full commands can be parsed by comma",
			allowCommands: "apply,plan,unlock,policy_check,approve_policies,version,import,state",
			want: []command.Name{
				command.Apply, command.Plan, command.Unlock, command.PolicyCheck, command.ApprovePolicies, command.Version, command.Import, command.State,
			},
		},
		{
			name:          "all",
			allowCommands: "all",
			want: []command.Name{
				command.Version, command.Plan, command.Apply, command.Unlock, command.ApprovePolicies, command.Import, command.State,
			},
		},
		{
			name:          "all with others returns same with all result",
			allowCommands: "all,plan",
			want: []command.Name{
				command.Version, command.Plan, command.Apply, command.Unlock, command.ApprovePolicies, command.Import, command.State,
			},
		},
		{
			name:          "empty",
			allowCommands: "",
			want:          nil,
		},
		{
			name:          "invalid command",
			allowCommands: "plan,all,invalid",
			wantErr:       "unknown command name: invalid",
		},
		{
			name:          "invalid command",
			allowCommands: "invalid,plan,all",
			wantErr:       "unknown command name: invalid",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := server.UserConfig{
				AllowCommands: tt.allowCommands,
			}
			got, err := u.ToAllowCommandNames()
			if err != nil {
				assert.ErrorContains(t, err, tt.wantErr, "ToAllowCommandNames()")
			}
			assert.Equalf(t, tt.want, got, "ToAllowCommandNames()")
		})
	}
}

func TestUserConfig_ToLogLevel(t *testing.T) {
	cases := []struct {
		userLvl string
		expLvl  logging.LogLevel
	}{
		{
			"debug",
			logging.Debug,
		},
		{
			"info",
			logging.Info,
		},
		{
			"warn",
			logging.Warn,
		},
		{
			"error",
			logging.Error,
		},
		{
			"unknown",
			logging.Info,
		},
	}

	for _, c := range cases {
		t.Run(c.userLvl, func(t *testing.T) {
			u := server.UserConfig{
				LogLevel: c.userLvl,
			}
			Equals(t, c.expLvl, u.ToLogLevel())
		})
	}
}
