package converter

import (
	"fmt"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
)

// PullGetter makes API calls to get pull requests.
type PullGetter interface {
	// GetPullRequest gets the pull request with id pullNum for the repo.
	GetPullRequest(repo models.Repo, pullNum int) (*github.PullRequest, error)
}

type CommentEventConverter struct {
	PullConverter PullConverter
	PullGetter    PullGetter
}

// Convert converts github issue comment events to our internal representation
func (e CommentEventConverter) Convert(comment *github.IssueCommentEvent) (event.Comment, error) {
	baseRepo, err := e.PullConverter.RepoConverter.Convert(comment.Repo)
	if err != nil {
		return event.Comment{}, err
	}
	if comment.Comment == nil || comment.Comment.User.GetLogin() == "" {
		return event.Comment{}, fmt.Errorf("comment.user.login is null")
	}
	commenterUsername := comment.Comment.User.GetLogin()
	user := models.User{
		Username: commenterUsername,
	}
	pullNum := comment.Issue.GetNumber()
	if pullNum == 0 {
		return event.Comment{}, fmt.Errorf("issue.number is null")
	}

	eventTimestamp := time.Now()
	if comment.Comment.CreatedAt != nil {
		eventTimestamp = *comment.Comment.CreatedAt
	}

	ghPull, err := e.PullGetter.GetPullRequest(baseRepo, pullNum)
	if err != nil {
		return event.Comment{}, errors.Wrap(err, "getting pull from github")
	}

	pull, err := e.PullConverter.Convert(ghPull)
	if err != nil {
		return event.Comment{}, errors.Wrap(err, "converting pull request type")
	}

	return event.Comment{
		BaseRepo:  pull.BaseRepo,
		HeadRepo:  pull.HeadRepo,
		Pull:      pull,
		User:      user,
		PullNum:   pullNum,
		VCSHost:   models.Github,
		Timestamp: eventTimestamp,
		Comment:   comment.GetComment().GetBody(),
	}, nil
}
