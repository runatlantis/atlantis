package activities

import (
	"context"
	"github.com/google/go-github/v45/github"
	"github.com/hashicorp/go-getter"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	internal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
	"net/http"
	"path/filepath"
	"time"
)

const deploymentsDirName = "deployments"

type githubActivities struct {
	ClientCreator githubapp.ClientCreator
	DataDir       string
	LinkBuilder   LinkBuilder
}

type CreateCheckRunRequest struct {
	Title      string
	Sha        string
	Repo       internal.Repo
	State      internal.CheckRunState
	Conclusion internal.CheckRunConclusion
	Summary    string
}

type UpdateCheckRunRequest struct {
	Title      string
	State      internal.CheckRunState
	Conclusion internal.CheckRunConclusion
	Repo       internal.Repo
	ID         int64
	Summary    string
}

type CreateCheckRunResponse struct {
	ID int64
}
type UpdateCheckRunResponse struct {
	ID int64
}

func (a *githubActivities) UpdateCheckRun(ctx context.Context, request UpdateCheckRunRequest) (UpdateCheckRunResponse, error) {

	output := github.CheckRunOutput{
		Title:   &request.Title,
		Text:    &request.Title,
		Summary: &request.Summary,
	}

	opts := github.UpdateCheckRunOptions{
		Name:   request.Title,
		Status: github.String(string(request.State)),
		Output: &output,
	}

	// Conclusion is required if status is Completed
	if request.State == internal.CheckRunComplete {
		opts.Conclusion = github.String(string(request.Conclusion))
	}

	client, err := a.ClientCreator.NewInstallationClient(request.Repo.Credentials.InstallationToken)

	if err != nil {
		return UpdateCheckRunResponse{}, errors.Wrap(err, "creating installation client")
	}

	run, _, err := client.Checks.UpdateCheckRun(ctx, request.Repo.Owner, request.Repo.Name, request.ID, opts)

	if err != nil {
		return UpdateCheckRunResponse{}, errors.Wrap(err, "creating check run")
	}

	return UpdateCheckRunResponse{
		ID: run.GetID(),
	}, nil
}

func (a *githubActivities) CreateCheckRun(ctx context.Context, request CreateCheckRunRequest) (CreateCheckRunResponse, error) {
	output := github.CheckRunOutput{
		Title:   &request.Title,
		Text:    &request.Title,
		Summary: &request.Summary,
	}

	opts := github.CreateCheckRunOptions{
		Name:    request.Title,
		HeadSHA: request.Sha,
		Status:  github.String("queued"),
		Output:  &output,
	}

	var state internal.CheckRunState
	if request.State == internal.CheckRunState("") {
		state = internal.CheckRunQueued
	} else {
		state = request.State
	}

	opts.Status = github.String(string(state))

	// Conclusion is required if status is Completed
	if state == internal.CheckRunComplete {
		opts.Conclusion = github.String(string(request.Conclusion))
	}

	client, err := a.ClientCreator.NewInstallationClient(request.Repo.Credentials.InstallationToken)

	if err != nil {
		return CreateCheckRunResponse{}, errors.Wrap(err, "creating installation client")
	}

	run, _, err := client.Checks.CreateCheckRun(ctx, request.Repo.Owner, request.Repo.Name, opts)

	if err != nil {
		return CreateCheckRunResponse{}, errors.Wrap(err, "creating check run")
	}

	return CreateCheckRunResponse{
		ID: run.GetID(),
	}, nil
}

type FetchRootRequest struct {
	Repo         internal.Repo
	Root         root.Root
	DeploymentId string
	Revision     string
}

type FetchRootResponse struct {
	LocalRoot *root.LocalRoot
}

// FetchRoot fetches a link to the archive URL using the GH client, processes that URL into a download URL that the
// go-getter library can use, and then go-getter to download/extract files/subdirs within the root path to the destinationPath.
func (a *githubActivities) FetchRoot(ctx context.Context, request FetchRootRequest) (FetchRootResponse, error) {
	ctx, cancel := temporal.StartHeartbeat(ctx, 10*time.Second)
	defer cancel()
	ref, err := request.Repo.HeadCommit.Ref.String()
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "processing request ref")
	}
	destinationPath := filepath.Join(a.DataDir, deploymentsDirName, request.DeploymentId)
	opts := &github.RepositoryContentGetOptions{
		Ref: ref,
	}
	client, err := a.ClientCreator.NewInstallationClient(request.Repo.Credentials.InstallationToken)
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "creating installation client")
	}
	// note: this link exists for 5 minutes when fetching a private repository archive
	archiveLink, resp, err := client.Repositories.GetArchiveLink(ctx, request.Repo.Owner, request.Repo.Name, github.Zipball, opts, true)
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "getting repo archive link")
	}
	// GH responds with a 302 + redirect link to where the archive exists
	if resp.StatusCode != http.StatusFound {
		return FetchRootResponse{}, errors.Errorf("getting repo archive link returns non-302 status %d", resp.StatusCode)
	}
	downloadLink := a.LinkBuilder.BuildDownloadLinkFromArchive(archiveLink, request.Root, request.Repo, request.Revision)
	err = getter.Get(destinationPath, downloadLink, getter.WithContext(ctx))
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "fetching and extracting zip")
	}
	localRoot := root.BuildLocalRoot(request.Root, request.Repo, destinationPath)
	return FetchRootResponse{
		LocalRoot: localRoot,
	}, nil
}
