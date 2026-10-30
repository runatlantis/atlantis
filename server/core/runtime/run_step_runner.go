// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	TerraformExecutor     TerraformExec
	DefaultTFDistribution terraform.Distribution
	DefaultTFVersion      *version.Version
	// TerraformBinDir is the directory where Atlantis downloads Terraform binaries.
	TerraformBinDir         string
	ProjectCmdOutputHandler jobs.ProjectCommandOutputHandler
}

func (r *RunStepRunner) Run(
	ctx command.ProjectContext,
	shell *valid.CommandShell,
	command string,
	path string,
	envs map[string]string,
	streamOutput bool,
	postProcessOutput []valid.PostProcessRunOutputOption,
	postProcessFilterRegexes []*regexp.Regexp,
) (string, error) {
	tfDistribution := r.DefaultTFDistribution
	tfVersion := r.DefaultTFVersion
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution, "", terraform.APIAuth{})
	}
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	err := r.TerraformExecutor.EnsureVersion(ctx.Log, tfDistribution, tfVersion)
	if err != nil {
		err = fmt.Errorf("%s: Downloading terraform Version %s", err, tfVersion.String())
		ctx.Log.Debug("error: %s", err)
		return "", err
	}

	planFile := GetPlanFilePath(ctx, path)
	if err := EnsurePlanFileDir(ctx, path); err != nil {
		return "", err
	}

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"ATLANTIS_TERRAFORM_DISTRIBUTION": tfDistribution.BinName(),
		"ATLANTIS_TERRAFORM_VERSION":      tfVersion.String(),
		"BASE_BRANCH_NAME":                ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":                  ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":                 ctx.BaseRepo.Owner,
		"COMMENT_ARGS":                    strings.Join(ctx.EscapedCommentArgs, ","),
		"DIR":                             path,
		"HEAD_BRANCH_NAME":                ctx.Pull.HeadBranch,
		"HEAD_COMMIT":                     ctx.Pull.HeadCommit,
		"HEAD_REPO_NAME":                  ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":                 ctx.HeadRepo.Owner,
		"PATH":                            fmt.Sprintf("%s:%s", os.Getenv("PATH"), r.TerraformBinDir),
		"PLANFILE":                        planFile,
		"SHOWFILE":                        filepath.Join(path, ctx.GetShowResultFileName()),
		"POLICYCHECKFILE":                 filepath.Join(path, ctx.GetPolicyCheckResultFileName()),
		"PROJECT_NAME":                    ctx.ProjectName,
		"PULL_AUTHOR":                     ctx.Pull.Author,
		"PULL_NUM":                        fmt.Sprintf("%d", ctx.Pull.Num),
		"PULL_URL":                        ctx.Pull.URL,
		"REPO_REL_DIR":                    ctx.RepoRelDir,
		"USER_NAME":                       ctx.User.Username,
		"WORKSPACE":                       ctx.Workspace,
	}
	// Add PR metadata environment variables for plan and apply steps
	if ctx.CommandName.String() == "plan" || ctx.CommandName.String() == "apply" {
		customEnvVars["ATLANTIS_PR_APPROVED"] = strconv.FormatBool(ctx.PullReqStatus.ApprovalStatus.IsApproved)
		customEnvVars["ATLANTIS_PR_MERGEABLE"] = strconv.FormatBool(ctx.PullReqStatus.MergeableStatus.IsMergeable)
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range envs {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}

	runner := models.NewShellCommandRunner(shell, command, finalEnvVars, path, streamOutput, r.ProjectCmdOutputHandler)
	output, err := runner.Run(ctx)

	// Post-process before the error check so that a failing command never
	// carries output into the error that the step asked to hide or filter.
	output = postProcessRunOutput(output, tfVersion, postProcessOutput, postProcessFilterRegexes)

	if err != nil {
		err = runStepError{
			err:          err,
			command:      command,
			path:         path,
			output:       output,
			streamOutput: streamOutput,
		}
		if !ctx.CustomPolicyCheck {
			ctx.Log.Debug("error: %s", err)
		} else {
			ctx.Log.Debug("Treating custom policy tool error exit code as a policy failure.  Error output: %s", err)
		}
		return "", err
	}

	return output, nil
}

// postProcessRunOutput applies the step's output options to the raw command
// output. It is applied on both the success and the error path.
func postProcessRunOutput(
	output string,
	tfVersion *version.Version,
	postProcessOutput []valid.PostProcessRunOutputOption,
	postProcessFilterRegexes []*regexp.Regexp,
) string {
	for _, processOutput := range postProcessOutput {
		switch processOutput {
		case valid.PostProcessRunOutputHide:
			output = ""
		case valid.PostProcessRunOutputStripRefreshing:
			output = StripRefreshingFromPlanOutput(output, tfVersion)
		case valid.PostProcessRunOutputFilterRegexKey:
			for _, filterRegex := range postProcessFilterRegexes {
				output = FilterRegexFromPlanOutput(output, filterRegex)
			}
		}
	}
	return output
}

// hideRunStepErrorOutput returns err with the command output removed from any
// runStepError it carries. Callers that need the raw output on success but
// must not report it on failure use this instead of the hide option.
func hideRunStepErrorOutput(err error) error {
	if stepErr, ok := errors.AsType[runStepError](err); ok {
		stepErr.output = ""
		return stepErr
	}
	return err
}

type runStepError struct {
	err          error
	command      string
	path         string
	output       string
	streamOutput bool
}

func (e runStepError) Error() string {
	return fmt.Sprintf("%s: running %q in %q: \n%s", e.err, e.command, e.path, e.output)
}

func (e runStepError) JobMessage() string {
	if !e.streamOutput && e.output != "" {
		return e.Error()
	}
	return e.err.Error()
}
