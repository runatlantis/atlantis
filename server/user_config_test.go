// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package server_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			allowCommands: "apply,plan,cancel,unlock,policy_check,approve_policies,version,import,state",
			want: []command.Name{
				command.Apply, command.Plan, command.Cancel, command.Unlock, command.PolicyCheck, command.ApprovePolicies, command.Version, command.Import, command.State,
			},
		},
		{
			name:          "all",
			allowCommands: "all",
			want: []command.Name{
				command.Version, command.Plan, command.Apply, command.Cancel, command.Unlock, command.ApprovePolicies, command.Import, command.State,
			},
		},
		{
			name:          "all with others returns same with all result",
			allowCommands: "all,plan",
			want: []command.Name{
				command.Version, command.Plan, command.Apply, command.Cancel, command.Unlock, command.ApprovePolicies, command.Import, command.State,
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
				require.ErrorContains(t, err, tt.wantErr, "ToAllowCommandNames()")
			}
			assert.Equalf(t, tt.want, got, "ToAllowCommandNames()")
		})
	}
}

func TestUserConfig_ToBlockedExtraArgs(t *testing.T) {
	tests := []struct {
		name             string
		blockedExtraArgs string
		want             []string
	}{
		{
			name:             "empty returns defaults",
			blockedExtraArgs: "",
			want:             events.DefaultBlockedExtraArgs,
		},
		{
			name:             "single flag",
			blockedExtraArgs: "-chdir",
			want:             []string{"-chdir"},
		},
		{
			name:             "multiple flags comma-separated",
			blockedExtraArgs: "-chdir,--chdir,-plugin-dir,--plugin-dir",
			want:             []string{"-chdir", "--chdir", "-plugin-dir", "--plugin-dir"},
		},
		{
			name:             "custom flag list overrides defaults",
			blockedExtraArgs: "-no-color,--no-color",
			want:             []string{"-no-color", "--no-color"},
		},
		{
			name:             "whitespace around flags is trimmed",
			blockedExtraArgs: " -chdir , --chdir ",
			want:             []string{"-chdir", "--chdir"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := server.UserConfig{
				BlockedExtraArgs: tt.blockedExtraArgs,
			}
			got := u.ToBlockedExtraArgs()
			assert.Equalf(t, tt.want, got, "ToBlockedExtraArgs()")
		})
	}
}

func TestUserConfig_ToWebhookHttpHeaders(t *testing.T) {
	tcs := []struct {
		name  string
		given string
		want  map[string][]string
		err   error
	}{
		{
			name:  "empty",
			given: "",
			want:  nil,
		},
		{
			name:  "happy path",
			given: `{"Authorization":"Bearer some-token","X-Custom-Header":["value1","value2"]}`,
			want: map[string][]string{
				"Authorization":   {"Bearer some-token"},
				"X-Custom-Header": {"value1", "value2"},
			},
		},
		{
			name:  "invalid json",
			given: `{"X-Custom-Header":true}`,
			err:   errors.New("expected string or array, got bool"),
		},
		{
			name:  "invalid json array element",
			given: `{"X-Custom-Header":[1, 2]}`,
			err:   errors.New("expected string array element, got float64"),
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			u := server.UserConfig{
				WebhookHttpHeaders: tc.given,
			}
			got, err := u.ToWebhookHttpHeaders()
			Equals(t, tc.want, got)
			Equals(t, tc.err, err)
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

func TestUserConfig_ToWebhookSecret(t *testing.T) {
	secretValue := "fakesecret"
	secretValueWithNewline := "fakesecret\n"

	tempDir := t.TempDir()
	secretFilePath := filepath.Join(tempDir, "webhook-secret")
	secretWithNewlineFilePath := filepath.Join(tempDir, "webhook-secret-newline")

	os.WriteFile(secretFilePath, []byte(secretValue), 0600)                       // nolint: errcheck
	os.WriteFile(secretWithNewlineFilePath, []byte(secretValueWithNewline), 0600) // nolint: errcheck

	cases := []struct {
		name   string
		config server.UserConfig
	}{
		{
			name: "secret from file",
			config: server.UserConfig{
				BitbucketWebhookSecretFile: secretFilePath,
				GithubWebhookSecretFile:    secretFilePath,
				GiteaWebhookSecretFile:     secretFilePath,
				GitlabWebhookSecretFile:    secretFilePath,
			},
		},
		{
			name: "secret from file with trailing newline",
			config: server.UserConfig{
				BitbucketWebhookSecretFile: secretWithNewlineFilePath,
				GithubWebhookSecretFile:    secretWithNewlineFilePath,
				GiteaWebhookSecretFile:     secretWithNewlineFilePath,
				GitlabWebhookSecretFile:    secretWithNewlineFilePath,
			},
		},
		{
			name: "secret from config flag",
			config: server.UserConfig{
				BitbucketWebhookSecret: secretValue,
				GithubWebhookSecret:    secretValue,
				GiteaWebhookSecret:     secretValue,
				GitlabWebhookSecret:    secretValue,
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			toWebhookSecretInner(t, c.config, []byte(secretValue))
		})
	}
}

type getSecret func() ([]byte, error)

func toWebhookSecretInner(t *testing.T, config server.UserConfig, expected []byte) {
	cases := []struct {
		name string
		f    getSecret
	}{
		{
			"bitbucket",
			config.ToBitbucketWebhookSecret,
		},
		{
			"github",
			config.ToGithubWebhookSecret,
		},
		{
			"gitea",
			config.ToGiteaWebhookSecret,
		},
		{
			"gitlab",
			config.ToGitlabWebhookSecret,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			secret, err := c.f()
			if err != nil {
				t.Errorf("failed to read file: %v", err)
			}

			Equals(t, expected, secret)
		})
	}
}
