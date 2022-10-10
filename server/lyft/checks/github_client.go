package checks

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/uber-go/tally/v4"
)

var projectCommandTemplateWithLogs = `
| **Command Name** | **Project** | **Workspace** | **Status** | **Logs** |  
| - | - | - | - | - |
| %s  | {%s}  | {%s}  | {%s}  | %s  | 
`

var projectCommandTemplate = `
| **Command Name**  | **Project** | **Workspace** | **Status** | 
| - | - | - | - |
| %s  | {%s}  | {%s}  | {%s}  |
`

var commandTemplate = `
| **Command Name**  |  **Status** |  
| - | - |
| %s  | {%s}  | 
`

var commandTemplateWithCount = `
| **Command Name** | **Num Total** | **Num Success** | **Status** |  
| - | - | - | - |
| %s | {%s} | {%s} | {%s} |  
`

const (
	// Reference: https://github.com/github/docs/issues/3765
	maxChecksOutputLength = 65535
)

// github checks conclusion
type ChecksConclusion int //nolint:golint // avoiding refactor while adding linter action

const (
	Neutral ChecksConclusion = iota
	TimedOut
	ActionRequired
	Cancelled
	Failure
	Success
)

func (e ChecksConclusion) String() string {
	switch e {
	case Neutral:
		return "neutral"
	case TimedOut:
		return "timed_out"
	case ActionRequired:
		return "action_required"
	case Cancelled:
		return "cancelled"
	case Failure:
		return "failure"
	case Success:
		return "success"
	}
	return ""
}

// github checks status
type CheckStatus int

const (
	Queued CheckStatus = iota
	InProgress
	Completed
)

func (e CheckStatus) String() string {
	switch e {
	case Queued:
		return "queued"
	case InProgress:
		return "in_progress"
	case Completed:
		return "completed"
	}
	return ""
}

// [WENGINES-4643] TODO: Remove this wrapper and add checks implementation to UpdateStatus() directly after github checks is stable
type ChecksClientWrapper struct { //nolint:golint // avoiding refactor while adding linter action
	*vcs.GithubClient
	FeatureAllocator feature.Allocator
	Logger           logging.Logger

	// Adds metric on checks vs commit status api usage to get an estimate of when all PRs have cut over to using the new status updates
	Scope tally.Scope
}

func (c *ChecksClientWrapper) UpdateStatus(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	shouldAllocate, err := c.FeatureAllocator.ShouldAllocate(feature.GithubChecks, feature.FeatureContext{
		RepoName:         request.Repo.FullName,
		PullCreationTime: request.PullCreationTime,
	})
	if err != nil {
		c.Logger.ErrorContext(ctx, fmt.Sprintf("unable to allocate for feature: %s", feature.GithubChecks), map[string]interface{}{
			"error": err.Error(),
		})
	}

	if !shouldAllocate {
		c.Scope.Counter("commit_status").Inc(1)
		return c.GithubClient.UpdateStatus(ctx, request)
	}
	c.Scope.Counter("checks").Inc(1)

	// Empty status ID means we create a new check run
	if request.StatusID == "" {
		return c.createCheckRun(ctx, request)
	}
	return request.StatusID, c.updateCheckRun(ctx, request, request.StatusID)
}

func (c *ChecksClientWrapper) createCheckRun(ctx context.Context, request types.UpdateStatusRequest) (string, error) {
	status, conclusion := c.resolveChecksStatus(request.State)
	createCheckRunOpts := github.CreateCheckRunOptions{
		Name:    request.StatusName,
		HeadSHA: request.Ref,
		Status:  &status,
		Output:  c.createCheckRunOutput(request),
	}

	if request.DetailsURL != "" {
		createCheckRunOpts.DetailsURL = &request.DetailsURL
	}

	// Conclusion is required if status is Completed
	if status == Completed.String() {
		createCheckRunOpts.Conclusion = &conclusion
	}

	return c.GithubClient.CreateCheckRun(ctx, request.Repo.Owner, request.Repo.Name, createCheckRunOpts)
}

func (c *ChecksClientWrapper) updateCheckRun(ctx context.Context, request types.UpdateStatusRequest, checkRunID string) error {
	status, conclusion := c.resolveChecksStatus(request.State)
	updateCheckRunOpts := github.UpdateCheckRunOptions{
		Name:   request.StatusName,
		Status: &status,
		Output: c.createCheckRunOutput(request),
	}

	if request.DetailsURL != "" {
		updateCheckRunOpts.DetailsURL = &request.DetailsURL
	}

	// Conclusion is required if status is Completed
	if status == Completed.String() {
		updateCheckRunOpts.Conclusion = &conclusion
	}

	checkRunIDInt, err := strconv.ParseInt(checkRunID, 10, 64)
	if err != nil {
		return err
	}

	return c.GithubClient.UpdateCheckRun(ctx, request.Repo.Owner, request.Repo.Name, checkRunIDInt, updateCheckRunOpts)
}

func (c *ChecksClientWrapper) resolveState(state models.CommitStatus) string {
	switch state {
	case models.PendingCommitStatus:
		return "In Progress"
	case models.SuccessCommitStatus:
		return "Success"
	case models.FailedCommitStatus:
		return "Failed"
	}
	return "Failed"
}

func (c *ChecksClientWrapper) createCheckRunOutput(request types.UpdateStatusRequest) *github.CheckRunOutput {

	var summary string

	// Project command
	if strings.Contains(request.StatusName, ":") {

		// plan/apply command
		if request.DetailsURL != "" {
			summary = fmt.Sprintf(projectCommandTemplateWithLogs,
				request.CommandName,
				request.Project,
				request.Workspace,
				c.resolveState(request.State),
				fmt.Sprintf("[Logs](%s)", request.DetailsURL),
			)
		} else {
			summary = fmt.Sprintf(projectCommandTemplate,
				request.CommandName,
				request.Project,
				request.Workspace,
				c.resolveState(request.State),
			)
		}
	} else {
		if request.NumSuccess != "" && request.NumTotal != "" {
			summary = fmt.Sprintf(commandTemplateWithCount,
				request.CommandName,
				request.NumTotal,
				request.NumSuccess,
				c.resolveState(request.State))
		} else {
			summary = fmt.Sprintf(commandTemplate,
				request.CommandName,
				c.resolveState(request.State))
		}

	}

	// Add formatting to summary
	summary = strings.ReplaceAll(strings.ReplaceAll(summary, "{", "`"), "}", "`")

	checkRunOutput := github.CheckRunOutput{
		Title:   &request.StatusName,
		Summary: &summary,
	}

	if request.Output != "" {
		checkRunOutput.Text = c.capCheckRunOutput(request.Output)
	}

	return &checkRunOutput
}

// Cap the output string if it exceeds the max checks output length
func (c *ChecksClientWrapper) capCheckRunOutput(output string) *string {
	cappedOutput := output
	if len(output) > maxChecksOutputLength {
		cappedOutput = output[:maxChecksOutputLength]
	}
	return &cappedOutput
}

// Github Checks uses Status and Conclusion to report status of the check run. Need to map models.CommitStatus to Status and Conclusion
// Status -> queued, in_progress, completed
// Conclusion -> failure, neutral, cancelled, timed_out, or action_required. (Optional. Required if you provide a status of "completed".)
func (c *ChecksClientWrapper) resolveChecksStatus(state models.CommitStatus) (string, string) {
	status := Queued
	conclusion := Neutral

	switch state {
	case models.SuccessCommitStatus:
		status = Completed
		conclusion = Success

	case models.PendingCommitStatus:
		status = InProgress

	case models.FailedCommitStatus:
		status = Completed
		conclusion = Failure
	}

	return status.String(), conclusion.String()
}
