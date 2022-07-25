package server_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

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

func TestUserConfig_ToLyftMode(t *testing.T) {
	cases := []struct {
		userMode string
		expMode  server.Mode
	}{
		{
			"default",
			server.Default,
		},
		{
			"gateway",
			server.Gateway,
		},
		{
			"worker",
			server.Worker,
		},
		{
			"unknown",
			server.Default,
		},
		{
			"",
			server.Default,
		},
	}

	for _, c := range cases {
		t.Run(c.userMode, func(t *testing.T) {
			u := server.UserConfig{
				LyftMode: c.userMode,
			}
			Equals(t, c.expMode, u.ToLyftMode())
		})
	}
}
