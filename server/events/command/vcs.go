package command

import (
	"context"
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
)

// VCSStatusUpdater updates the status of a commit with the VCS host. We set
// the status to signify whether the plan/apply succeeds.
type VCSStatusUpdater struct {
	Client       vcs.Client
	TitleBuilder vcs.StatusTitleBuilder
}

func (d *VCSStatusUpdater) UpdateCombined(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName fmt.Stringer, checkRunId string) (string, error) {
	src := d.TitleBuilder.Build(cmdName.String())
	descrip := fmt.Sprintf("%s %s", strings.Title(cmdName.String()), d.statusDescription(status))

	request := types.UpdateStatusRequest{
		Repo:             repo,
		PullNum:          pull.Num,
		Ref:              pull.HeadCommit,
		StatusName:       src,
		State:            status,
		Description:      descrip,
		DetailsURL:       "",
		PullCreationTime: pull.CreatedAt,
		StatusId:         checkRunId,
	}
	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) UpdateCombinedCount(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.CommitStatus, cmdName fmt.Stringer, numSuccess int, numTotal int, checkRunId string) (string, error) {
	src := d.TitleBuilder.Build(cmdName.String())
	cmdVerb := "unknown"

	switch cmdName {
	case Plan:
		cmdVerb = "planned"
	case PolicyCheck:
		cmdVerb = "policies checked"
	case Apply:
		cmdVerb = "applied"
	}

	request := types.UpdateStatusRequest{
		Repo:             repo,
		PullNum:          pull.Num,
		Ref:              pull.HeadCommit,
		StatusName:       src,
		State:            status,
		Description:      fmt.Sprintf("%d/%d projects %s successfully.", numSuccess, numTotal, cmdVerb),
		DetailsURL:       "",
		PullCreationTime: pull.CreatedAt,
		StatusId:         checkRunId,
	}

	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) UpdateProject(ctx context.Context, projectCtx ProjectContext, cmdName fmt.Stringer, status models.CommitStatus, url string, checkRunId string) (string, error) {
	projectID := projectCtx.ProjectName
	if projectID == "" {
		projectID = fmt.Sprintf("%s/%s", projectCtx.RepoRelDir, projectCtx.Workspace)
	}
	statusName := d.TitleBuilder.Build(cmdName.String(), vcs.StatusTitleOptions{
		ProjectName: projectID,
	})

	description := fmt.Sprintf("%s %s", strings.Title(cmdName.String()), d.statusDescription(status))
	request := types.UpdateStatusRequest{
		Repo:             projectCtx.BaseRepo,
		PullNum:          projectCtx.Pull.Num,
		Ref:              projectCtx.Pull.HeadCommit,
		StatusName:       statusName,
		State:            status,
		Description:      description,
		DetailsURL:       url,
		PullCreationTime: projectCtx.Pull.CreatedAt,
		StatusId:         checkRunId,
	}

	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) statusDescription(status models.CommitStatus) string {
	var description string
	switch status {
	case models.PendingCommitStatus:
		description = "in progress..."
	case models.FailedCommitStatus:
		description = "failed."
	case models.SuccessCommitStatus:
		description = "succeeded."
	}

	return description
}
