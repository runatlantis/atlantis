package activities

import (
	"context"
	"net/http"
	"path/filepath"
	"time"

	"github.com/google/go-github/v45/github"
	"github.com/hashicorp/go-getter"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	internal "github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/temporal"
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
	Summary    string
	ExternalID string
}

type UpdateCheckRunRequest struct {
	Title   string
	State   internal.CheckRunState
	Actions []internal.CheckRunAction
	Repo    internal.Repo
	ID      int64
	Summary string
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

	state, conclusion := getCheckStateAndConclusion(request.State)

	opts := github.UpdateCheckRunOptions{
		Name:   request.Title,
		Status: github.String(state),
		Output: &output,
	}

	// update with any actions
	if len(request.Actions) != 0 {
		var actions []*github.CheckRunAction

		for _, a := range request.Actions {
			actions = append(actions, a.ToGithubAction())
		}

		opts.Actions = actions
	}

	if conclusion != "" {
		opts.Conclusion = github.String(conclusion)
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

	state, conclusion := getCheckStateAndConclusion(request.State)

	opts := github.CreateCheckRunOptions{
		Name:       request.Title,
		HeadSHA:    request.Sha,
		Status:     &state,
		Output:     &output,
		ExternalID: &request.ExternalID,
	}

	if conclusion != "" {
		opts.Conclusion = &conclusion
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

func getCheckStateAndConclusion(internalState internal.CheckRunState) (string, string) {
	var state string
	var conclusion string
	// checks are weird in that success and failure are defined in the conclusion, and the state
	// is just marked as complete, let's just deal with that stuff here because it's not intuitive for
	// callers
	switch internalState {

	// default to queued if we have nothing
	case internal.CheckRunUnknown:
		state = string(internal.CheckRunQueued)
	case internal.CheckRunFailure:
		state = "completed"
		conclusion = "failure"
	case internal.CheckRunSuccess:
		state = "completed"
		conclusion = "success"
	default:
		state = string(internalState)
	}

	return state, conclusion
}

type FetchRootRequest struct {
	Repo         internal.Repo
	Root         root.Root
	DeploymentID string
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
	ref, err := request.Repo.Ref.String()
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "processing request ref")
	}
	destinationPath := filepath.Join(a.DataDir, deploymentsDirName, request.DeploymentID)
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
