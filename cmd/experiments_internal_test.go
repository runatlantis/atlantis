// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"os"
	"strings"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestExecute_Experiments(t *testing.T) {
	cases := []struct {
		name      string
		yaml      string
		env       string
		args      []string
		want      []string
		apiSecret string
		settings  map[string]string
		drift     bool
		remediate bool
		err       string
	}{
		{name: "default"},
		{name: "api without permission", args: []string{"--api-secret=secret"}, err: "api-secret requires"},
		{name: "api enabled", args: []string{"--experiments=api", "--api-secret=secret"}, want: []string{"api"}, apiSecret: "secret"},
		{name: "drift enabled independently", args: []string{"--experiments=drift-detection", "--api-secret=secret", "--enable-drift-detection"}, want: []string{"drift-detection"}, apiSecret: "secret", drift: true},
		{name: "remediation enabled", args: []string{"--experiments=api,drift-detection", "--api-secret=secret", "--enable-drift-detection", "--enable-drift-remediation"}, want: []string{"api", "drift-detection"}, apiSecret: "secret", drift: true, remediate: true},
		{name: "drift without permission", args: []string{"--enable-drift-detection"}, err: `enable-drift-detection requires the "drift-detection" experiment`},
		{name: "remediation without permission", args: []string{"--enable-drift-remediation"}, err: `enable-drift-remediation requires the "drift-detection" experiment`},
		{name: "API permission does not permit drift", args: []string{"--experiments=api", "--api-secret=secret", "--enable-drift-detection"}, err: `enable-drift-detection requires the "drift-detection" experiment`},
		{name: "API yaml without permission", yaml: "api-secret: secret\n", err: "api-secret requires"},
		{name: "API environment without permission", settings: map[string]string{"ATLANTIS_API_SECRET": "secret"}, err: "api-secret requires"},
		{name: "valid API environment", env: "api", settings: map[string]string{"ATLANTIS_API_SECRET": "secret"}, want: []string{"api"}, apiSecret: "secret"},
		{name: "drift environment without permission", settings: map[string]string{"ATLANTIS_ENABLE_DRIFT_DETECTION": "true"}, err: `enable-drift-detection requires the "drift-detection" experiment`},
		{name: "remediation environment without permission", settings: map[string]string{"ATLANTIS_ENABLE_DRIFT_REMEDIATION": "true"}, err: `enable-drift-remediation requires the "drift-detection" experiment`},
		{name: "valid drift environment", env: "drift-detection", settings: map[string]string{"ATLANTIS_API_SECRET": "secret", "ATLANTIS_ENABLE_DRIFT_DETECTION": "true"}, want: []string{"drift-detection"}, apiSecret: "secret", drift: true},
		{name: "drift yaml without permission", yaml: "enable-drift-detection: true\n", err: `enable-drift-detection requires the "drift-detection" experiment`},
		{name: "valid drift yaml", yaml: "experiments: [drift-detection]\napi-secret: secret\nenable-drift-detection: true\n", want: []string{"drift-detection"}, apiSecret: "secret", drift: true},
		{name: "clearing permission rejects enabled API", yaml: "experiments: [api]\napi-secret: secret\n", args: []string{"--experiments="}, err: "api-secret requires"},
		{name: "api flag", args: []string{"--experiments=api"}, want: []string{"api"}},
		{name: "api environment", env: "api", want: []string{"api"}},
		{name: "api yaml", yaml: "experiments: [api]\n", want: []string{"api"}},
		{name: "independent experiments", args: []string{"--experiments=api,drift-detection"}, want: []string{"api", "drift-detection"}},
		{name: "different repeated flags", args: []string{"--experiments=api", "--experiments=drift-detection"}, want: []string{"api", "drift-detection"}},
		{name: "both environment", env: "api,drift-detection", want: []string{"api", "drift-detection"}},
		{name: "both yaml", yaml: "experiments: [api, drift-detection]\n", want: []string{"api", "drift-detection"}},
		{name: "flag", args: []string{"--experiments=drift-detection"}, want: []string{"drift-detection"}},
		{name: "repeated flags", args: []string{"--experiments=drift-detection", "--experiments=drift-detection"}, want: []string{"drift-detection", "drift-detection"}},
		{name: "comma separated flags", args: []string{"--experiments=drift-detection,drift-detection"}, want: []string{"drift-detection", "drift-detection"}},
		{name: "environment", env: "drift-detection", want: []string{"drift-detection"}},
		{name: "comma separated environment", env: "drift-detection,drift-detection", want: []string{"drift-detection", "drift-detection"}},
		{name: "yaml", yaml: "experiments:\n  - drift-detection\n", want: []string{"drift-detection"}},
		{name: "empty yaml list", yaml: "experiments: []\n"},
		{name: "environment overrides yaml", yaml: "experiments: [unknown]\n", env: "drift-detection", want: []string{"drift-detection"}},
		{name: "api overrides drift yaml", yaml: "experiments: [drift-detection]\n", env: "api", want: []string{"api"}},
		{name: "drift flag overrides api environment", env: "api", args: []string{"--experiments=drift-detection"}, want: []string{"drift-detection"}},
		{name: "flag overrides environment and yaml", yaml: "experiments: [unknown]\n", env: "unknown", args: []string{"--experiments=drift-detection"}, want: []string{"drift-detection"}},
		{name: "empty flag clears lower precedence", yaml: "experiments: [unknown]\n", env: "unknown", args: []string{"--experiments="}},
		{name: "unknown", args: []string{"--experiments=unknown"}, err: `unknown server experiment "unknown"`},
		{name: "unknown environment", env: "unknown", err: `unknown server experiment "unknown"`},
		{name: "unknown yaml", yaml: "experiments: [unknown]\n", err: `unknown server experiment "unknown"`},
		{name: "case sensitive", args: []string{"--experiments=Drift-Detection"}, err: `unknown server experiment "Drift-Detection"`},
		{name: "api case sensitive", args: []string{"--experiments=API"}, err: `unknown server experiment "API"`},
		{name: "no wildcard", args: []string{"--experiments=*"}, err: `unknown server experiment "*"`},
		{name: "no all", args: []string{"--experiments=all"}, err: `unknown server experiment "all"`},
		{name: "no whitespace normalization", env: " drift-detection", err: `unknown server experiment " drift-detection"`},
		{name: "no whitespace separated environment", env: "drift-detection drift-detection", err: "unknown server experiment"},
		{name: "empty entry", args: []string{"--experiments=drift-detection,"}, err: `unknown server experiment ""`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ATLANTIS_EXPERIMENTS", tc.env)
			for _, name := range []string{"ATLANTIS_API_SECRET", "ATLANTIS_ENABLE_DRIFT_DETECTION", "ATLANTIS_ENABLE_DRIFT_REMEDIATION"} {
				t.Setenv(name, "")
			}
			for name, value := range tc.settings {
				t.Setenv(name, value)
			}
			flags := map[string]any{}
			if tc.yaml != "" {
				file := tempFile(t, tc.yaml)
				t.Cleanup(func() { _ = os.Remove(file) })
				flags[ConfigFlag] = file
			}
			c := setupWithDefaults(flags, t)
			c.SetArgs(tc.args)
			err := c.Execute()
			if tc.err != "" {
				ErrContains(t, tc.err, err)
				if strings.Contains(tc.err, "unknown server experiment") {
					ErrContains(t, "supported names are api, drift-detection", err)
				}
				ErrContains(t, "https://www.runatlantis.io/versions/main/docs/experiments.html", err)
				Assert(t, !strings.Contains(err.Error(), "in --experiments"), "error assumes the setting came from a flag")
				return
			}
			Ok(t, err)
			Equals(t, tc.apiSecret, passedConfig.APISecret)
			Equals(t, len(tc.want), len(passedConfig.Experiments))
			for i, name := range tc.want {
				Equals(t, name, passedConfig.Experiments[i])
			}
			Equals(t, tc.drift, passedConfig.EnableDriftDetection)
			Equals(t, tc.remediate, passedConfig.EnableDriftRemediation)
		})
	}
}
