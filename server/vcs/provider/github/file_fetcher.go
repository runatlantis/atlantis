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

type FileFetcherOptions struct {
	Sha   string
	PRNum int
}

func (r *RemoteFileFetcher) GetModifiedFiles(ctx context.Context, repo models.Repo, installationToken int64, fileFetcherOptions FileFetcherOptions) ([]string, error) {
	client, err := r.ClientCreator.NewInstallationClient(installationToken)
	if err != nil {
		return nil, errors.Wrap(err, "creating installation client")
	}

	var fileFetcher func(ctx context.Context, client *gh.Client, repo models.Repo, fileFetcherOptions FileFetcherOptions, listOptions gh.ListOptions) ([]*gh.CommitFile, *gh.Response, error)
	if fileFetcherOptions.Sha != "" {
		fileFetcher = GetCommit
	} else if fileFetcherOptions.PRNum != 0 {
		fileFetcher = ListFiles
	} else {
		return nil, errors.New("invalid fileFetcherOptions")
	}

	var files []string
	nextPage := 0
	for {
		listOptions := gh.ListOptions{
			PerPage: 300,
		}
		if nextPage != 0 {
			listOptions.Page = nextPage
		}

		pageFiles, resp, err := fileFetcher(ctx, client, repo, fileFetcherOptions, listOptions)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching files")
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("not ok status fetching files: %s", resp.Status)
		}
		for _, f := range pageFiles {
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

func GetCommit(ctx context.Context, client *gh.Client, repo models.Repo, fileFetcherOptions FileFetcherOptions, listOptions gh.ListOptions) ([]*gh.CommitFile, *gh.Response, error) {
	repositoryCommit, resp, err := client.Repositories.GetCommit(ctx, repo.Owner, repo.Name, fileFetcherOptions.Sha, &listOptions)
	if repositoryCommit != nil {
		return repositoryCommit.Files, resp, err
	}
	return nil, nil, errors.New("unable to retrieve commit files from GH commit")
}

func ListFiles(ctx context.Context, client *gh.Client, repo models.Repo, fileFetcherOptions FileFetcherOptions, listOptions gh.ListOptions) ([]*gh.CommitFile, *gh.Response, error) {
	return client.PullRequests.ListFiles(ctx, repo.Owner, repo.Name, fileFetcherOptions.PRNum, &listOptions)
}
