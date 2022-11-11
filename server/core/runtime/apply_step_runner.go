package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

// ApplyStepRunner runs `terraform apply`.
type ApplyStepRunner struct {
	TerraformExecutor TerraformExec
	VCSStatusUpdater  StatusUpdater
	AsyncTFExec       AsyncTFExec
}

func (a *ApplyStepRunner) Run(ctx context.Context, prjCtx command.ProjectContext, extraArgs []string, path string, envs map[string]string) (string, error) {
	if a.hasTargetFlag(prjCtx, extraArgs) {
		return "", errors.New("cannot run apply with -target because we are applying an already generated plan. Instead, run -target with atlantis plan")
	}

	planPath := filepath.Join(path, GetPlanFilename(prjCtx.Workspace, prjCtx.ProjectName))
	contents, err := os.ReadFile(planPath)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("no plan found at path %q and workspace %q–did you run plan?", prjCtx.RepoRelDir, prjCtx.Workspace)
	}
	if err != nil {
		return "", errors.Wrap(err, "unable to read planfile")
	}

	prjCtx.Log.InfoContext(prjCtx.RequestCtx, "starting apply")
	var out string

	// TODO: Leverage PlanTypeStepRunnerDelegate here
	if IsRemotePlan(contents) {
		args := append(append([]string{"apply", "-input=false"}, extraArgs...), prjCtx.EscapedCommentArgs...)
		out, err = a.runRemoteApply(ctx, prjCtx, args, path, planPath, prjCtx.TerraformVersion, envs)
		if err == nil {
			out = a.cleanRemoteApplyOutput(out)
		}
	} else {
		// NOTE: we need to quote the plan path because Bitbucket Server can
		// have spaces in its repo owner names which is part of the path.
		args := append(append(append([]string{"apply", "-input=false"}, extraArgs...), prjCtx.EscapedCommentArgs...), fmt.Sprintf("%q", planPath))
		out, err = a.TerraformExecutor.RunCommandWithVersion(ctx, prjCtx, path, args, envs, prjCtx.TerraformVersion, prjCtx.Workspace)
	}

	// If the apply was successful, delete the plan.
	if err == nil {
		prjCtx.Log.InfoContext(prjCtx.RequestCtx, "apply successful, deleting planfile")
		if removeErr := os.Remove(planPath); removeErr != nil {
			prjCtx.Log.WarnContext(prjCtx.RequestCtx, fmt.Sprintf("failed to delete planfile after successful apply: %s", removeErr))
		}
	}
	return out, err
}

func (a *ApplyStepRunner) hasTargetFlag(prjCtx command.ProjectContext, extraArgs []string) bool {
	isTargetFlag := func(s string) bool {
		if s == "-target" {
			return true
		}
		split := strings.Split(s, "=")
		return split[0] == "-target"
	}

	for _, arg := range prjCtx.EscapedCommentArgs {
		if isTargetFlag(arg) {
			return true
		}
	}
	for _, arg := range extraArgs {
		if isTargetFlag(arg) {
			return true
		}
	}
	return false
}

// cleanRemoteApplyOutput removes unneeded output like the refresh and plan
// phases to make the final comment cleaner.
func (a *ApplyStepRunner) cleanRemoteApplyOutput(out string) string {
	applyStartText := `  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: 
`
	applyStartIdx := strings.Index(out, applyStartText)
	if applyStartIdx < 0 {
		return out
	}
	return out[applyStartIdx+len(applyStartText):]
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
	ctx context.Context,
	prjCtx command.ProjectContext,
	applyArgs []string,
	path string,
	absPlanPath string,
	tfVersion *version.Version,
	envs map[string]string) (string, error) {

	// The planfile contents are needed to ensure that the plan didn't change
	// between plan and apply phases.
	planfileBytes, err := os.ReadFile(absPlanPath)
	if err != nil {
		return "", errors.Wrap(err, "reading planfile")
	}

	// updateStatusF will update the commit status and log any error.
	updateStatusF := func(status models.VCSStatus, url string, statusID string) {
		if _, err := a.VCSStatusUpdater.UpdateProject(ctx, prjCtx, command.Apply, status, url, statusID); err != nil {
			prjCtx.Log.ErrorContext(prjCtx.RequestCtx, fmt.Sprintf("unable to update status: %s", err))
		}
	}

	// Start the async command execution.
	inCh := make(chan string)
	defer close(inCh)
	outCh := a.AsyncTFExec.RunCommandAsyncWithInput(ctx, prjCtx, filepath.Clean(path), applyArgs, envs, tfVersion, prjCtx.Workspace, inCh)
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
			updateStatusF(models.PendingVCSStatus, runURL, prjCtx.StatusID)
			nextLineIsRunURL = false
		}

		// If the plan is complete and it's waiting for us to verify the apply,
		// check if the plan is the same and if so, input "yes".
		if a.atConfirmApplyPrompt(lines) {
			// Check if the plan is as expected.
			planChangedErr = a.remotePlanChanged(string(planfileBytes), strings.Join(lines, "\n"), tfVersion)
			if planChangedErr != nil {
				prjCtx.Log.ErrorContext(prjCtx.RequestCtx, "plan generated during apply does not match expected plan, aborting")
				inCh <- "no\n"
				// Need to continue so we read all the lines, otherwise channel
				// sender (in TerraformClient) will block indefinitely waiting
				// for us to read.
				continue
			}

			inCh <- "yes\n"
		}
	}

	output := strings.Join(lines, "\n")
	if planChangedErr != nil {
		updateStatusF(models.FailedVCSStatus, runURL, prjCtx.StatusID)
		// The output isn't important if the plans don't match so we just
		// discard it.
		return "", planChangedErr
	}

	if err != nil {
		updateStatusF(models.FailedVCSStatus, runURL, prjCtx.StatusID)
	} else {
		updateStatusF(models.SuccessVCSStatus, runURL, prjCtx.StatusID)
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
	planEndIdx := strings.Index(output, "Do you want to perform these actions in workspace \"")
	if planEndIdx < 0 {
		return fmt.Errorf("Couldn't find plan end when parsing apply output:\n%q", applyOut)
	}
	currPlan := strings.TrimSpace(output[:planEndIdx])

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
