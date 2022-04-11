package converter

import (
	"fmt"
	"time"

	"github.com/google/go-github/v31/github"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/vcs/types/event"
)

type CommentEventConverter struct {
	RepoConverter RepoConverter
}

// Convert converts github issue comment events to our internal representation
// TODO: remove named methods and return CommentEvent directly /techdebt
func (e CommentEventConverter) Convert(comment *github.IssueCommentEvent) (event.Comment, error) {
	baseRepo, err := e.RepoConverter.Convert(comment.Repo)
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

	return event.Comment{
		BaseRepo:      baseRepo,
		MaybeHeadRepo: nil,
		MaybePull:     nil,
		User:          user,
		PullNum:       pullNum,
		VCSHost:       models.Github,
		Timestamp:     eventTimestamp,
		Comment:       comment.GetComment().GetBody(),
	}, nil
}
