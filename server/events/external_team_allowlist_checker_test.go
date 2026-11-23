// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/runtime"
	runtime_mocks "github.com/runatlantis/atlantis/server/core/runtime/mocks"
	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

var extTeamAllowlistChecker events.ExternalTeamAllowlistChecker
var extTeamAllowlistCheckerRunner *runtime_mocks.MockExternalTeamAllowlistRunner

func externalTeamAllowlistCheckerSetup(t *testing.T) {
	RegisterMockTestingT(t)
	extTeamAllowlistCheckerRunner = runtime_mocks.NewMockExternalTeamAllowlistRunner()

	extTeamAllowlistChecker = events.ExternalTeamAllowlistChecker{
		ExternalTeamAllowlistRunner: extTeamAllowlistCheckerRunner,
	}
}

func TestIsCommandAllowedForTeam(t *testing.T) {
	ctx := models.TeamAllowlistCheckerContext{
		Log: logging.NewNoopLogger(t),
	}

	t.Run("allowed", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]())).ThenReturn("pass\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForTeam(ctx, "foo", "plan")
		Equals(t, true, res)
	})

	t.Run("denied", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]())).ThenReturn("nothing found\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForTeam(ctx, "foo", "plan")
		Equals(t, false, res)
	})
}

func TestIsCommandAllowedForAnyTeam(t *testing.T) {
	ctx := models.TeamAllowlistCheckerContext{
		Log: logging.NewNoopLogger(t),
	}

	t.Run("allowed", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]())).ThenReturn("pass\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, []string{"foo"}, "plan")
		Equals(t, true, res)
	})

	t.Run("denied", func(t *testing.T) {
		externalTeamAllowlistCheckerSetup(t)

		When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]())).ThenReturn("nothing found\n", nil)

		res := extTeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, []string{"foo"}, "plan")
		Equals(t, false, res)
	})
}

func TestExternalTeamAllowlistChecker_PassesVCSValuesAsArguments(t *testing.T) {
	externalTeamAllowlistCheckerSetup(t)
	extTeamAllowlistChecker.Command = "/scripts/check.sh"
	extTeamAllowlistChecker.ExtraArgs = []string{"--mode", "strict"}
	ctx := models.TeamAllowlistCheckerContext{
		Log:      logging.NewNoopLogger(t),
		BaseRepo: models.Repo{FullName: "x;echo pass #/repo", Owner: "x;echo pass #"},
	}
	When(extTeamAllowlistCheckerRunner.Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]())).ThenReturn("deny\n", nil)

	res := extTeamAllowlistChecker.IsCommandAllowedForAnyTeam(ctx, []string{"t1", "t2"}, "apply")
	Equals(t, false, res)

	_, shell, args := extTeamAllowlistCheckerRunner.VerifyWasCalledOnce().Run(Any[models.TeamAllowlistCheckerContext](), Any[string](), Any[[]string]()).GetCapturedArguments()
	Equals(t, "sh", shell)
	Equals(t, []string{
		"-c", `/scripts/check.sh --mode strict "$@"`, "atlantis-team-authz",
		"apply", "x;echo pass #/repo", "x;echo pass #/t1", "x;echo pass #/t2",
	}, args)
}

// writeTeamAuthzScript writes a script that records its arguments, one per
// line, and prints "pass" only when its first argument is "--ok".
func writeTeamAuthzScript(t *testing.T) (script string, argsFile string) {
	t.Helper()
	dir := t.TempDir()
	argsFile = filepath.Join(dir, "args")
	script = filepath.Join(dir, "check.sh")
	body := "#!/bin/sh\nprintf '%s\\n' \"$@\" > '" + argsFile + "'\n" +
		"if [ \"$1\" = \"--ok\" ]; then echo pass; else echo deny; fi\n"
	Ok(t, os.WriteFile(script, []byte(body), 0o700)) // nolint: gosec
	return script, argsFile
}

func TestExternalTeamAllowlistChecker_RepoNameIsNotShellSource(t *testing.T) {
	cases := map[string]string{
		"bitbucket server project name": "x;touch MARKER;echo pass #/repo",
		"azure devops repository name":  "org/project/x&&touch MARKER&&echo pass",
		"command substitution":          "org/$(touch MARKER)`touch MARKER`",
	}
	for name, fullName := range cases {
		t.Run(name, func(t *testing.T) {
			script, argsFile := writeTeamAuthzScript(t)
			workDir := t.TempDir()
			t.Chdir(workDir)
			checker := events.ExternalTeamAllowlistChecker{
				Command:                     script,
				ExternalTeamAllowlistRunner: runtime.DefaultExternalTeamAllowlistRunner{},
			}
			ctx := models.TeamAllowlistCheckerContext{
				Log:      logging.NewNoopLogger(t),
				BaseRepo: models.Repo{FullName: fullName, Owner: "org"},
			}

			Equals(t, false, checker.IsCommandAllowedForTeam(ctx, "team", "apply"))
			Equals(t, false, checker.IsCommandAllowedForAnyTeam(ctx, []string{"team"}, "apply"))

			_, err := os.Stat(filepath.Join(workDir, "MARKER"))
			Assert(t, os.IsNotExist(err), "injected command ran: %v", err)
			got, err := os.ReadFile(argsFile)
			Ok(t, err)
			Equals(t, "apply\n"+fullName+"\norg/team\n", string(got))
		})
	}
}

func TestExternalTeamAllowlistChecker_CommandWithEmbeddedArgs(t *testing.T) {
	script, argsFile := writeTeamAuthzScript(t)
	ctx := models.TeamAllowlistCheckerContext{
		Log:      logging.NewNoopLogger(t),
		BaseRepo: models.Repo{FullName: "org/repo", Owner: "org"},
	}

	t.Run("args inside command", func(t *testing.T) {
		checker := events.ExternalTeamAllowlistChecker{
			Command:                     script + " --ok",
			ExternalTeamAllowlistRunner: runtime.DefaultExternalTeamAllowlistRunner{},
		}
		Equals(t, true, checker.IsCommandAllowedForAnyTeam(ctx, []string{"a", "b"}, "plan"))
		got, err := os.ReadFile(argsFile)
		Ok(t, err)
		Equals(t, "--ok\nplan\norg/repo\norg/a\norg/b\n", string(got))
	})

	t.Run("args from config", func(t *testing.T) {
		checker := events.ExternalTeamAllowlistChecker{
			Command:                     script,
			ExtraArgs:                   []string{"--ok"},
			ExternalTeamAllowlistRunner: runtime.DefaultExternalTeamAllowlistRunner{},
		}
		Equals(t, true, checker.IsCommandAllowedForTeam(ctx, "a", "plan"))
		got, err := os.ReadFile(argsFile)
		Ok(t, err)
		Equals(t, "--ok\nplan\norg/repo\norg/a\n", string(got))
	})
}
