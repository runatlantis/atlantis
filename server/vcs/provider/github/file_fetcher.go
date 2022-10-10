package github

import (
	"context"
	"fmt"
	"net/http"

	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type RemoteFileFetcher struct {
	ClientCreator githubapp.ClientCreator
}

func (r *RemoteFileFetcher) GetModifiedFilesFromCommit(ctx context.Context, repo models.Repo, sha string, installationToken int64) ([]string, error) {
	var files []string
	nextPage := 0
	for {
		opts := gh.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			opts.Page = nextPage
		}
		client, err := r.ClientCreator.NewInstallationClient(installationToken)
		if err != nil {
			return files, errors.Wrap(err, "creating installation client")
		}
		repositoryCommit, resp, err := client.Repositories.GetCommit(ctx, repo.Owner, repo.Name, sha, &opts)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching repository commit")
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("not ok status fetching repository commit: %s", resp.Status)
		}
		for _, f := range repositoryCommit.Files {
			files = append(files, f.GetFilename())

			// If the file was renamed, we'll want to run plan in the directory
			// it was moved from as well.
			if f.GetStatus() == "renamed" {
				files = append(files, f.GetPreviousFilename())
			}
		}
		if resp.NextPage == 0 {
			break
		}
		nextPage = resp.NextPage
	}
	return files, nil
}
