// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ApplyStepRunner runs `terraform apply`.
type ApplyStepRunner struct {
	TerraformExecutor     TerraformExec          `validate:"required"`
	DefaultTFDistribution terraform.Distribution `validate:"required"`
	DefaultTFVersion      *version.Version       `validate:"required"`
	CommitStatusUpdater   StatusUpdater          `validate:"required"`
	AsyncTFExec           AsyncTFExec            `validate:"required"`
	PlanStore             PlanStore              `validate:"required"`
}

func (a *ApplyStepRunner) Run(ctx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	// extra_args comes from configuration, so environment variable references
	// in it may be expanded. Marked on this copy of the context; everything
	// else, including comment args, stays literal.
	if len(extraArgs) > 0 {
		ctx.ExpandableArgs = extraArgs
	}
	if a.hasTargetFlag(ctx, extraArgs) {
		return "", errors.New("cannot run apply with -target because we are applying an already generated plan. Instead, run -target with atlantis plan")
	}

	planPath := GetPlanFilePath(ctx, path)
	if loadErr := a.PlanStore.Load(ctx, planPath); loadErr != nil {
		return "", fmt.Errorf("loading plan: %w", loadErr)
	}
	contents, err := os.ReadFile(planPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", ctx.RepoRelDir, ctx.Workspace)
	}
	if err != nil {
		return "", fmt.Errorf("unable to read planfile: %w", err)
	}

	// This runner is itself a built-in apply step, irrespective of the
	// context's managed-plan classification. Never fall back to mutable bytes.
	if ctx.ExpectedPlanHash == "" {
		return "", fmt.Errorf("expected plan hash is missing for dir %q workspace %q project %q; run `atlantis plan` before apply", ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)
	}
	digest := sha256.Sum256(contents)
	if hex.EncodeToString(digest[:]) != ctx.ExpectedPlanHash {
		return "", fmt.Errorf("plan file changed for dir %q workspace %q project %q; run `atlantis plan` before apply", ctx.RepoRelDir, ctx.Workspace, ctx.ProjectName)
	}

	ctx.Log.Info("starting apply")
	var out string
	tfDistribution := a.DefaultTFDistribution
	if ctx.TerraformDistribution != nil {
		tfDistribution = terraform.NewDistribution(*ctx.TerraformDistribution)
	}
	tfVersion := a.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	// TODO: Leverage PlanTypeStepRunnerDelegate here
	if IsRemotePlan(contents) {
		args := append(append([]string{"apply", "-input=false", "-no-color"}, extraArgs...), ctx.CommentArgs...)
		out, err = a.runRemoteApply(ctx, args, path, contents, tfDistribution, tfVersion, envs)
		if err == nil {
			out = a.cleanRemoteApplyOutput(out)
		}
	} else {
		executionPlanPath, cleanup, snapshotErr := writeValidatedPlanSnapshot(contents)
		if snapshotErr != nil {
			return "", snapshotErr
		}
		defer cleanup()
		// NOTE: we need to quote the plan path because Bitbucket Server can
		// have spaces in its repo owner names which is part of the path.
		// planPath is passed as its own argument, so a path containing a space
		// needs no quoting; quoting it would make the quotes part of the path.
		args := append(append(append([]string{"apply", "-input=false"}, extraArgs...), ctx.CommentArgs...), executionPlanPath)
		out, err = a.TerraformExecutor.RunCommandWithVersion(ctx, path, args, envs, tfDistribution, tfVersion, ctx.Workspace)
	}

	// If the apply was successful, delete the plan.
	if err == nil {
		ctx.Log.Info("apply successful, deleting planfile")
		if removeErr := a.PlanStore.Remove(ctx, planPath); removeErr != nil {
			ctx.Log.Warn("failed to delete planfile after successful apply: %s", removeErr)
		}
	}
	return out, err
}

func (a *ApplyStepRunner) hasTargetFlag(ctx command.ProjectContext, extraArgs []string) bool {
	isTargetFlag := func(s string) bool {
		if s == "-target" {
			return true
		}
		split := strings.Split(s, "=")
		return split[0] == "-target"
	}

	// CommentArgs, not EscapedCommentArgs: the escaped form has a backslash
	// before every byte, so it never compares equal to "-target".
	if slices.ContainsFunc(ctx.CommentArgs, isTargetFlag) {
		return true
	}
	return slices.ContainsFunc(extraArgs, isTargetFlag)
}

// cleanRemoteApplyOutput removes unneeded output like the refresh and plan
// phases to make the final comment cleaner.
func (a *ApplyStepRunner) cleanRemoteApplyOutput(out string) string {
	applyStartText := `  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: 
`
	_, after, found := strings.Cut(out, applyStartText)
	if !found {
		return out
	}
	return after
}

// runRemoteApply handles running the apply and performing actions in real-time
// as we get the output from the command.
// Specifically, we set commit statuses with links to Terraform Enterprise's
// UI to view real-time output.
// We also check if the plan that's about to be applied matches the one we
// printed to the pull request.
// We need to do this because remote plan doesn't support -out, so we do a
// manual diff.
// It also writes "yes" or "no" to the process to confirm the apply.
func (a *ApplyStepRunner) runRemoteApply(
	ctx command.ProjectContext,
	applyArgs []string,
	path string,
	planfileBytes []byte,
	tfDistribution terraform.Distribution,
	tfVersion *version.Version,
	envs map[string]string) (string, error) {
	var err error

	// updateStatusF will update the commit status and log any error.
	updateStatusF := func(status models.CommitStatus, url string) {
		if err := a.CommitStatusUpdater.UpdateProject(ctx, command.Apply, status, url, nil); err != nil {
			ctx.Log.Err("unable to update status: %s", err)
		}
	}

	// Start the async command execution.
	ctx.Log.Debug("starting async tf remote operation")
	inCh, outCh := a.AsyncTFExec.RunCommandAsync(ctx, filepath.Clean(path), applyArgs, envs, tfDistribution, tfVersion, ctx.Workspace)
	var lines []string
	nextLineIsRunURL := false
	var runURL string
	var planChangedErr error

	for line := range outCh {
		if line.Err != nil {
			err = line.Err
			break
		}
		lines = append(lines, line.Line)

		// Here we're checking for the run url and updating the status
		// if found.
		if line.Line == lineBeforeRunURL {
			nextLineIsRunURL = true
		} else if nextLineIsRunURL {
			runURL = strings.TrimSpace(line.Line)
			if ctx.RemoteApplyRunURL != nil {
				*ctx.RemoteApplyRunURL = runURL
			}
			ctx.Log.Debug("remote run url found, updating commit status")
			updateStatusF(models.PendingCommitStatus, runURL)
			nextLineIsRunURL = false
		}

		// If the plan is complete and it's waiting for us to verify the apply,
		// check if the plan is the same and if so, input "yes".
		if a.atConfirmApplyPrompt(lines) {
			ctx.Log.Debug("remote apply is waiting for confirmation")

			// Check if the plan is as expected.
			planChangedErr = a.remotePlanChanged(string(planfileBytes), strings.Join(lines, "\n"), tfVersion)
			if planChangedErr != nil {
				ctx.Log.Err("plan generated during apply does not match expected plan, aborting")
				inCh <- "no\n"
				// Need to continue so we read all the lines, otherwise channel
				// sender (in TerraformClient) will block indefinitely waiting
				// for us to read.
				continue
			}

			ctx.Log.Debug("plan generated during apply matches expected plan, continuing")
			inCh <- "yes\n"
		}
	}

	ctx.Log.Debug("async tf remote operation complete")
	output := strings.Join(lines, "\n")
	if planChangedErr != nil {
		updateStatusF(models.FailedCommitStatus, runURL)
		// The output isn't important if the plans don't match so we just
		// discard it.
		return "", planChangedErr
	}

	if err != nil {
		updateStatusF(models.FailedCommitStatus, runURL)
	} else {
		ctx.Log.Debug("remote apply succeeded; deferring success status until final apply freshness validation")
	}
	return output, err
}

// remotePlanChanged checks if the plan generated during the plan phase matches
// the one we're about to apply in the apply phase.
// If the plans don't match, it returns an error with a diff of the two plans
// that can be printed to the pull request.
func (a *ApplyStepRunner) remotePlanChanged(planfileContents string, applyOut string, tfVersion *version.Version) error {
	output := StripRefreshingFromPlanOutput(applyOut, tfVersion)

	// Strip plan output after the prompt to execute the plan.
	before, _, found := strings.Cut(output, "Do you want to perform these actions in workspace \"")
	if !found {
		return fmt.Errorf("couldn't find plan end when parsing apply output:\n%q", applyOut)
	}
	currPlan := strings.TrimSpace(before)

	// Ensure we strip the remoteOpsHeader from the plan contents so the
	// comparison is fair. We add this header in the plan phase so we can
	// identify that this planfile came from a remote plan.
	expPlan := strings.TrimSpace(planfileContents[len(remoteOpsHeader):])

	if currPlan != expPlan {
		return fmt.Errorf(planChangedErrFmt, expPlan, currPlan)
	}
	return nil
}

// atConfirmApplyPrompt returns true if the apply is at the "confirm this apply" step.
// This is determined by looking at the current command output provided by
// applyLines.
func (a *ApplyStepRunner) atConfirmApplyPrompt(applyLines []string) bool {
	waitingMatchLines := strings.Split(waitingForConfirmation, "\n")
	return len(applyLines) >= len(waitingMatchLines) && reflect.DeepEqual(applyLines[len(applyLines)-len(waitingMatchLines):], waitingMatchLines)
}

// planChangedErrFmt is the error we print to pull requests when the plan changed
// between remote terraform plan and apply phases.
var planChangedErrFmt = `Plan generated during apply phase did not match plan generated during plan phase.
Aborting apply.

Expected Plan:

%s
**************************************************

Actual Plan:

%s
**************************************************

This likely occurred because someone applied a change to this state in-between
your plan and apply commands.
To resolve, re-run plan.`

// waitingForConfirmation is what is printed during a remote apply when
// terraform is waiting for confirmation to apply the plan.
var waitingForConfirmation = `  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.`

// Keep execution copies in a private temporary directory without changing the
// convention PLANFILE exposed to custom steps. The system temporary directory
// is normally outside the checkout, but operators can override it with TMPDIR.
// Terraform consumes only these verified bytes.
func writeValidatedPlanSnapshot(contents []byte) (string, func(), error) {
	dir, err := os.MkdirTemp("", "atlantis-validated-plan-")
	if err != nil {
		return "", nil, fmt.Errorf("creating validated plan directory: %w", err)
	}
	cleanup := func() { _ = os.RemoveAll(dir) }
	root, err := os.OpenRoot(dir)
	if err != nil {
		cleanup()
		return "", nil, fmt.Errorf("opening validated plan directory: %w", err)
	}
	defer root.Close()
	path := filepath.Join(dir, "plan.tfplan")
	if err := root.WriteFile("plan.tfplan", contents, 0400); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("writing validated plan snapshot: %w", err)
	}
	return path, cleanup, nil
}
