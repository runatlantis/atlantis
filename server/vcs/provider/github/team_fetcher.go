package github

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
)

type TeamMemberFetcher struct {
	ClientCreator githubapp.ClientCreator
	Org           string
}

func (t *TeamMemberFetcher) ListTeamMembers(ctx context.Context, installationToken int64, teamSlug string) ([]string, error) {
	client, err := t.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}
	var usernames []string
	run := func(ctx context.Context, nextPage int) ([]*gh.User, *gh.Response, error) {
		listOptions := &gh.TeamListTeamMembersOptions{
			ListOptions: gh.ListOptions{
				PerPage: 100,
			},
		}
		listOptions.Page = nextPage
		return client.Teams.ListTeamMembersBySlug(ctx, t.Org, teamSlug, listOptions)
	}
	users, err := Iterate(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "iterating through entries")
	}
	for _, user := range users {
		usernames = append(usernames, user.GetLogin())
	}
	return usernames, nil
}
