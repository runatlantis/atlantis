package converter

import (
	"testing"

	"github.com/google/go-github/v31/github"
	"github.com/mohae/deepcopy"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestCommentEvent_Convert(t *testing.T) {
	repoConverter := RepoConverter{
		GithubUser:  "github-user",
		GithubToken: "github-token",
	}
	subject := CommentEventConverter{
		RepoConverter: repoConverter,
	}

	comment := github.IssueCommentEvent{
		Repo: &Repo,
		Issue: &github.Issue{
			Number:  github.Int(1),
			User:    &github.User{Login: github.String("issue_user")},
			HTMLURL: github.String("https://github.com/runatlantis/atlantis/issues/1"),
		},
		Comment: &github.IssueComment{
			User: &github.User{Login: github.String("comment_user")},
		},
	}

	testComment := deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment = nil
	_, err := subject.Convert(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User = nil
	_, err = subject.Convert(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Comment.User.Login = nil
	_, err = subject.Convert(&testComment)
	ErrEquals(t, "comment.user.login is null", err)

	testComment = deepcopy.Copy(comment).(github.IssueCommentEvent)
	testComment.Issue = nil
	_, err = subject.Convert(&testComment)
	ErrEquals(t, "issue.number is null", err)

	// this should be successful
	commentEvent, err := subject.Convert(&comment)
	Ok(t, err)
	Equals(t, models.Repo{
		Owner:             *comment.Repo.Owner.Login,
		FullName:          *comment.Repo.FullName,
		CloneURL:          "https://github-user:github-token@github.com/owner/repo.git",
		SanitizedCloneURL: "https://github-user:<redacted>@github.com/owner/repo.git",
		Name:              "repo",
		VCSHost: models.VCSHost{
			Hostname: "github.com",
			Type:     models.Github,
		},
	}, commentEvent.BaseRepo)
	Equals(t, models.User{
		Username: *comment.Comment.User.Login,
	}, commentEvent.User)
	Equals(t, *comment.Issue.Number, commentEvent.PullNum)
}
