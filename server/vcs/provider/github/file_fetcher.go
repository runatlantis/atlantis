package github

import (
	"context"
	gh "github.com/google/go-github/v45/github"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

type RemoteFileFetcher struct {
	ClientCreator githubapp.ClientCreator
}

// FileFetcherOptions dictates the context of fetching modified files.
// If a sha is provided, we get the modified files within the context of said sha
// If a PR number is provided, we get the modified files for the whole PR
// If both, we prioritize PR number
type FileFetcherOptions struct {
	Sha   string
	PRNum int
}

func (r *RemoteFileFetcher) GetModifiedFiles(ctx context.Context, repo models.Repo, installationToken int64, fileFetcherOptions FileFetcherOptions) ([]string, error) {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}

	var run func(ctx context.Context, nextPage int) ([]*gh.CommitFile, *gh.Response, error)
	if fileFetcherOptions.PRNum != 0 {
		run = ListFiles(client, repo, fileFetcherOptions)
	} else if fileFetcherOptions.Sha != "" {
		run = GetCommit(client, repo, fileFetcherOptions)
	} else {
		return nil, errors.New("invalid fileFetcherOptions")
	}

	pageFiles, err := Iterate(ctx, run)
	if err != nil {
		return nil, errors.Wrap(err, "iterating through entries")
	}
	var renamed []string
	for _, f := range pageFiles {
		renamed = append(renamed, f.GetFilename())
		// If the file was renamed, we'll want to run plan in the directory
		// it was moved from as well.
		if f.GetStatus() == "renamed" {
			renamed = append(renamed, f.GetPreviousFilename())
		}
	}
	return renamed, nil
}

func GetCommit(client *gh.Client, repo models.Repo, fileFetcherOptions FileFetcherOptions) func(ctx context.Context, nextPage int) ([]*gh.CommitFile, *gh.Response, error) {
	return func(ctx context.Context, nextPage int) ([]*gh.CommitFile, *gh.Response, error) {
		listOptions := gh.ListOptions{
			PerPage: 100,
		}
		listOptions.Page = nextPage
		repositoryCommit, resp, err := client.Repositories.GetCommit(ctx, repo.Owner, repo.Name, fileFetcherOptions.Sha, &listOptions)
		if repositoryCommit != nil {
			return repositoryCommit.Files, resp, err
		}
		return nil, nil, errors.New("unable to retrieve commit files from GH commit")
	}
}

func ListFiles(client *gh.Client, repo models.Repo, fileFetcherOptions FileFetcherOptions) func(ctx context.Context, nextPage int) ([]*gh.CommitFile, *gh.Response, error) {
	return func(ctx context.Context, nextPage int) ([]*gh.CommitFile, *gh.Response, error) {
		listOptions := gh.ListOptions{
			PerPage: 100,
		}
		listOptions.Page = nextPage
		return client.PullRequests.ListFiles(ctx, repo.Owner, repo.Name, fileFetcherOptions.PRNum, &listOptions)
	}
}
