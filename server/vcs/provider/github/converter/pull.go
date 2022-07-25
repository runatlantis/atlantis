package converter

import (
	"fmt"

	"github.com/google/go-github/v31/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// PullConverter converts a github pull request to our internal model
type PullConverter struct {
	RepoConverter RepoConverter
}

func (p *PullConverter) Convert(pull *github.PullRequest) (models.PullRequest, error) {
	commit := pull.Head.GetSHA()
	if commit == "" {
		return models.PullRequest{}, fmt.Errorf("head.sha is null")
	}
	url := pull.GetHTMLURL()
	if url == "" {
		return models.PullRequest{}, fmt.Errorf("html_url is null")
	}
	headBranch := pull.Head.GetRef()
	if headBranch == "" {
		return models.PullRequest{}, fmt.Errorf("head.ref is null")
	}
	baseBranch := pull.Base.GetRef()
	if baseBranch == "" {
		return models.PullRequest{}, fmt.Errorf("base.ref is null")
	}

	authorUsername := pull.User.GetLogin()
	if authorUsername == "" {
		return models.PullRequest{}, fmt.Errorf("user.login is null")
	}
	num := pull.GetNumber()
	if num == 0 {
		return models.PullRequest{}, fmt.Errorf("number is null")
	}

	baseRepo, err := p.RepoConverter.Convert(pull.Base.Repo)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "converting base repo")
	}
	headRepo, err := p.RepoConverter.Convert(pull.Head.Repo)
	if err != nil {
		return models.PullRequest{}, errors.Wrap(err, "converting head repo")
	}

	pullState := models.ClosedPullState
	closedAt := pull.GetClosedAt()
	updatedAt := pull.GetUpdatedAt()
	createdAt := pull.GetCreatedAt()
	if pull.GetState() == "open" {
		pullState = models.OpenPullState
	}

	return models.PullRequest{
		Author:     authorUsername,
		HeadBranch: headBranch,
		HeadCommit: commit,
		URL:        url,
		Num:        num,
		State:      pullState,
		BaseRepo:   baseRepo,
		HeadRepo:   headRepo,
		BaseBranch: baseBranch,
		ClosedAt:   closedAt,
		UpdatedAt:  updatedAt,
		CreatedAt:  createdAt,
	}, nil
}
