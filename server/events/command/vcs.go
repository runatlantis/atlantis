package command

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/events/vcs/types"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// VCSStatusUpdater updates the status of a commit with the VCS host. We set
// the status to signify whether the plan/apply succeeds.
type VCSStatusUpdater struct {
	Client            vcs.Client
	TitleBuilder      vcs.StatusTitleBuilder
	DefaultDetailsURL string
}

func (d *VCSStatusUpdater) UpdateCombined(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.VCSStatus, cmdName fmt.Stringer, statusID string, output string) (string, error) {
	src := d.TitleBuilder.Build(cmdName.String())
	descrip := fmt.Sprintf("%s %s", cases.Title(language.English).String(cmdName.String()), d.statusDescription(status))

	request := types.UpdateStatusRequest{
		Repo:             repo,
		PullNum:          pull.Num,
		Ref:              pull.HeadCommit,
		StatusName:       src,
		State:            status,
		Description:      descrip,
		DetailsURL:       d.DefaultDetailsURL,
		PullCreationTime: pull.CreatedAt,
		StatusID:         statusID,
		CommandName:      titleString(cmdName),
		Output:           output,
	}
	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) UpdateCombinedCount(ctx context.Context, repo models.Repo, pull models.PullRequest, status models.VCSStatus, cmdName fmt.Stringer, numSuccess int, numTotal int, statusID string) (string, error) {
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
		DetailsURL:       d.DefaultDetailsURL,
		PullCreationTime: pull.CreatedAt,
		StatusID:         statusID,
		CommandName:      titleString(cmdName),

		// Additional fields for github checks rendering
		NumSuccess: strconv.FormatInt(int64(numSuccess), 10),
		NumTotal:   strconv.FormatInt(int64(numTotal), 10),
	}

	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) UpdateProject(ctx context.Context, projectCtx ProjectContext, cmdName fmt.Stringer, status models.VCSStatus, url string, statusID string) (string, error) {
	projectID := projectCtx.ProjectName
	if projectID == "" {
		projectID = fmt.Sprintf("%s/%s", projectCtx.RepoRelDir, projectCtx.Workspace)
	}
	statusName := d.TitleBuilder.Build(cmdName.String(), vcs.StatusTitleOptions{
		ProjectName: projectID,
	})

	description := fmt.Sprintf("%s %s", cases.Title(language.English).String(cmdName.String()), d.statusDescription(status))
	request := types.UpdateStatusRequest{
		Repo:             projectCtx.BaseRepo,
		PullNum:          projectCtx.Pull.Num,
		Ref:              projectCtx.Pull.HeadCommit,
		StatusName:       statusName,
		State:            status,
		Description:      description,
		DetailsURL:       url,
		PullCreationTime: projectCtx.Pull.CreatedAt,
		StatusID:         statusID,

		CommandName: titleString(cmdName),
		Project:     projectCtx.ProjectName,
		Workspace:   projectCtx.Workspace,
		Directory:   projectCtx.RepoRelDir,
	}

	return d.Client.UpdateStatus(ctx, request)
}

func (d *VCSStatusUpdater) statusDescription(status models.VCSStatus) string {
	var description string
	switch status {
	case models.QueuedVCSStatus:
		description = "queued."
	case models.PendingVCSStatus:
		description = "in progress..."
	case models.FailedVCSStatus:
		description = "failed."
	case models.SuccessVCSStatus:
		description = "succeeded."
	}

	return description
}

func titleString(cmdName fmt.Stringer) string {
	return cases.Title(language.English).String(strings.ReplaceAll(strings.ToLower(cmdName.String()), "_", " "))
}
