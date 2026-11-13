// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"maps"
	"strings"
	"testing"

	version "github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/core/terraform"
	terraformmocks "github.com/runatlantis/atlantis/server/core/terraform/mocks"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestMultiEnvStepRunner_Run(t *testing.T) {
	cases := []struct {
		Command     string
		ProjectName string
		Env         map[string]string
		Output      []valid.PostProcessRunOutputOption
		ExpOut      string
		ExpErr      string
		ExpErrOmits string
		ExpEnv      map[string]string
	}{
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_ONE=value1'`,
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpOut:  "Dynamic environment variables added:\nTF_VAR_REPODEFINEDVARIABLE_ONE\n",
			ExpEnv: map[string]string{
				"TF_VAR_REPODEFINEDVARIABLE_ONE": "value1",
			},
		},
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_TWO=value=1='`,
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpOut:  "Dynamic environment variables added:\nTF_VAR_REPODEFINEDVARIABLE_TWO\n",
			ExpEnv: map[string]string{
				"TF_VAR_REPODEFINEDVARIABLE_TWO": "value=1=",
			},
		},
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_NO_VALUE'`,
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpErr:  "invalid environment variable definition: remaining unparsed data at offset 0",
			ExpEnv:  map[string]string{},
		},
		{
			Command: `echo 'TF_VAR1_MULTILINE="foo\\nbar",TF_VAR2_VALUEWITHCOMMA="one,two",TF_VAR3_CONTROL=true'`,
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpOut:  "Dynamic environment variables added:\nTF_VAR1_MULTILINE\nTF_VAR2_VALUEWITHCOMMA\nTF_VAR3_CONTROL\n",
			ExpEnv: map[string]string{
				"TF_VAR1_MULTILINE":      "foo\\nbar",
				"TF_VAR2_VALUEWITHCOMMA": "one,two",
				"TF_VAR3_CONTROL":        "true",
			},
		},
		{
			Command: `echo 'TF_VAR_REPODEFINEDVARIABLE_HIDE=value1'`,
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputHide},
			ExpOut:  "",
			ExpEnv: map[string]string{
				"TF_VAR_REPODEFINEDVARIABLE_HIDE": "value1",
			},
		},
		{
			// An unparseable line must never be echoed back: it is the
			// name=value output this step exists to carry.
			Command:     `echo 'A=1,B="x'`,
			Output:      []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputHide},
			ExpErr:      "invalid environment variable definition: misquoted at offset 8",
			ExpErrOmits: `A=1`,
		},
		{
			Command:     `echo 'A=1,B="x'`,
			Output:      []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpErr:      "invalid environment variable definition: misquoted at offset 8",
			ExpErrOmits: `A=1`,
		},
		{
			// A failing command with `output: hide` reports the failure
			// without the output.
			Command:     `echo "TOKEN=$SECRET"; exit 1`,
			Env:         map[string]string{"SECRET": "hunter2"},
			Output:      []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputHide},
			ExpErr:      "exit status 1: running",
			ExpErrOmits: "hunter2",
		},
		{
			Command: `echo "TOKEN=$SECRET"; exit 1`,
			Env:     map[string]string{"SECRET": "hunter2"},
			Output:  []valid.PostProcessRunOutputOption{valid.PostProcessRunOutputShow},
			ExpErr:  "TOKEN=hunter2",
		},
	}
	RegisterMockTestingT(t)
	tfClient := tfclientmocks.NewMockClient()
	mockDownloader := terraformmocks.NewMockDownloader()
	tfDistribution := terraform.NewDistributionTerraformWithDownloader(mockDownloader)
	tfVersion, err := version.NewVersion("0.12.0")
	Ok(t, err)
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	runStepRunner := runtime.RunStepRunner{
		TerraformExecutor:       tfClient,
		DefaultTFDistribution:   tfDistribution,
		DefaultTFVersion:        tfVersion,
		ProjectCmdOutputHandler: projectCmdOutputHandler,
	}
	multiEnvStepRunner := runtime.MultiEnvStepRunner{
		RunStepRunner: &runStepRunner,
	}
	for _, c := range cases {
		t.Run(c.Command, func(t *testing.T) {
			tmpDir := t.TempDir()
			ctx := command.ProjectContext{
				BaseRepo: models.Repo{
					Name:  "basename",
					Owner: "baseowner",
				},
				HeadRepo: models.Repo{
					Name:  "headname",
					Owner: "headowner",
				},
				Pull: models.PullRequest{
					Num:        2,
					HeadBranch: "add-feat",
					BaseBranch: "main",
					Author:     "acme",
				},
				User: models.User{
					Username: "acme-user",
				},
				Log:              logging.NewNoopLogger(t),
				Workspace:        "myworkspace",
				RepoRelDir:       "mydir",
				TerraformVersion: tfVersion,
				ProjectName:      c.ProjectName,
			}
			envMap := make(map[string]string)
			maps.Copy(envMap, c.Env)
			value, err := multiEnvStepRunner.Run(ctx, nil, c.Command, tmpDir, envMap, c.Output)
			if c.ExpErr != "" {
				ErrContains(t, c.ExpErr, err)
				if c.ExpErrOmits != "" {
					Assert(t, !strings.Contains(err.Error(), c.ExpErrOmits), "error %q should not contain %q", err.Error(), c.ExpErrOmits)
				}
				return
			}
			Ok(t, err)
			Equals(t, c.ExpOut, value)
			Equals(t, c.ExpEnv, envMap)
		})
	}
}
