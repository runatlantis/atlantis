package activities

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	key "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/google/go-github/v45/github"
	"github.com/hashicorp/go-getter"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/logger"
	internal "github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/temporal"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

const PullRequestOpenState = "open"

type ClientContext struct {
	InstallationToken int64
	context.Context
}

var HashiGetter = func(ctx context.Context, dst, src string) error {
	return getter.Get(dst, src, getter.WithContext(ctx))
}

// wraps hashicorp's go getter to allow for testing
type gogetter func(ctx context.Context, dst, src string) error

type githubClient interface {
	CreateCheckRun(ctx internal.Context, owner, repo string, opts github.CreateCheckRunOptions) (*github.CheckRun, *github.Response, error)
	UpdateCheckRun(ctx internal.Context, owner, repo string, checkRunID int64, opts github.UpdateCheckRunOptions) (*github.CheckRun, *github.Response, error)
	GetArchiveLink(ctx internal.Context, owner, repo string, archiveformat github.ArchiveFormat, opts *github.RepositoryContentGetOptions, followRedirects bool) (*url.URL, *github.Response, error)
	CompareCommits(ctx internal.Context, owner, repo string, base, head string, opts *github.ListOptions) (*github.CommitsComparison, *github.Response, error)
	ListModifiedFiles(ctx internal.Context, owner, repo string, pullNumber int) ([]*github.CommitFile, error)
	ListPullRequests(ctx internal.Context, owner, repo, base, state string) ([]*github.PullRequest, error)
}

type DiffDirection string

const (
	DirectionAhead     DiffDirection = "ahead"
	DirectionBehind    DiffDirection = "behind"
	DirectionIdentical DiffDirection = "identical"
	DirectionDiverged  DiffDirection = "diverged"
)

const deploymentsDirName = "deployments"

type githubActivities struct {
	Client      githubClient
	DataDir     string
	LinkBuilder LinkBuilder
	Getter      gogetter
}

type CreateCheckRunRequest struct {
	Title      string
	Sha        string
	Repo       internal.Repo
	State      internal.CheckRunState
	Actions    []internal.CheckRunAction
	Summary    string
	ExternalID string
}

type UpdateCheckRunRequest struct {
	Title      string
	State      internal.CheckRunState
	Actions    []internal.CheckRunAction
	Repo       internal.Repo
	ID         int64
	Summary    string
	ExternalID string
}

type CreateCheckRunResponse struct {
	ID     int64
	Status string
}

type UpdateCheckRunResponse struct {
	ID     int64
	Status string
}

func (a *githubActivities) GithubUpdateCheckRun(ctx context.Context, request UpdateCheckRunRequest) (UpdateCheckRunResponse, error) {
	output := github.CheckRunOutput{
		Title:   &request.Title,
		Text:    &request.Title,
		Summary: &request.Summary,
	}

	state, conclusion := getCheckStateAndConclusion(request.State)

	opts := github.UpdateCheckRunOptions{
		Name:       request.Title,
		Status:     github.String(state),
		Output:     &output,
		ExternalID: &request.ExternalID,
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

	run, _, err := a.Client.UpdateCheckRun(
		internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken),
		request.Repo.Owner, request.Repo.Name, request.ID, opts,
	)

	if err != nil {
		return UpdateCheckRunResponse{}, errors.Wrap(err, "creating check run")
	}

	return UpdateCheckRunResponse{
		ID:     run.GetID(),
		Status: run.GetStatus(),
	}, nil
}

func (a *githubActivities) GithubCreateCheckRun(ctx context.Context, request CreateCheckRunRequest) (CreateCheckRunResponse, error) {
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

	// update with any actions
	if len(request.Actions) != 0 {
		var actions []*github.CheckRunAction

		for _, a := range request.Actions {
			actions = append(actions, a.ToGithubAction())
		}

		opts.Actions = actions
	}

	if conclusion != "" {
		opts.Conclusion = &conclusion
	}

	run, _, err := a.Client.CreateCheckRun(
		internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken),
		request.Repo.Owner, request.Repo.Name, opts,
	)

	if err != nil {
		return CreateCheckRunResponse{}, errors.Wrap(err, "creating check run")
	}

	return CreateCheckRunResponse{
		ID:     run.GetID(),
		Status: run.GetStatus(),
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
	case internal.CheckRunTimeout:
		state = "completed"
		conclusion = "timed_out"
	case internal.CheckRunActionRequired:
		state = "completed"
		conclusion = "action_required"
	default:
		state = string(internalState)
	}

	return state, conclusion
}

type FetchRootRequest struct {
	Repo         internal.Repo
	Root         terraform.Root
	DeploymentID string
	Revision     string
}

type FetchRootResponse struct {
	LocalRoot *terraform.LocalRoot

	// if we do more with this, we can consider moving generation into it's own activity
	DeployDirectory string
}

// FetchRoot fetches a link to the archive URL using the GH client, processes that URL into a download URL that the
// go-getter library can use, and then go-getter to download/extract files/subdirs within the root path to the destinationPath.
func (a *githubActivities) GithubFetchRoot(ctx context.Context, request FetchRootRequest) (FetchRootResponse, error) {
	cancel := temporal.StartHeartbeat(ctx, temporal.HeartbeatTimeout)
	defer cancel()

	deployBasePath := filepath.Join(a.DataDir, deploymentsDirName, request.DeploymentID)
	repositoryPath := filepath.Join(deployBasePath, "repo")
	opts := &github.RepositoryContentGetOptions{
		Ref: request.Revision,
	}
	// note: this link exists for 5 minutes when fetching a private repository archive
	archiveLink, resp, err := a.Client.GetArchiveLink(internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken), request.Repo.Owner, request.Repo.Name, github.Zipball, opts, true)
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "getting repo archive link")
	}
	// GH responds with a 302 + redirect link to where the archive exists
	if resp.StatusCode != http.StatusFound {
		return FetchRootResponse{}, errors.Errorf("getting repo archive link returns non-302 status %d", resp.StatusCode)
	}
	downloadLink := a.LinkBuilder.BuildDownloadLinkFromArchive(archiveLink, request.Root, request.Repo, request.Revision)
	err = a.Getter(ctx, repositoryPath, downloadLink)
	if err != nil {
		return FetchRootResponse{}, errors.Wrap(err, "fetching and extracting zip")
	}
	rootPath := filepath.Join(repositoryPath, request.Root.Path)

	// let's drop a symlink to the root in the deploy base path to make navigation easier
	rootSymlink := filepath.Join(deployBasePath, "root")
	err = os.Symlink(rootPath, rootSymlink)
	if err != nil {
		logger.Warn(ctx, "unable to symlink to terraform root", key.ErrKey, err)
	}

	localRoot := terraform.BuildLocalRoot(request.Root, request.Repo, rootPath)
	return FetchRootResponse{
		LocalRoot:       localRoot,
		DeployDirectory: deployBasePath,
	}, nil
}

type CompareCommitRequest struct {
	DeployRequestRevision  string
	LatestDeployedRevision string
	Repo                   internal.Repo
}

type CompareCommitResponse struct {
	CommitComparison DiffDirection
}

func (a *githubActivities) GithubCompareCommit(ctx context.Context, request CompareCommitRequest) (CompareCommitResponse, error) {
	comparison, resp, err := a.Client.CompareCommits(internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken), request.Repo.Owner, request.Repo.Name, request.LatestDeployedRevision, request.DeployRequestRevision, &github.ListOptions{})

	if err != nil {
		return CompareCommitResponse{}, errors.Wrap(err, "comparing commits")
	}

	if comparison.GetStatus() == "" || resp.StatusCode != http.StatusOK {
		return CompareCommitResponse{}, fmt.Errorf("invalid commit comparison status: %s, Status Code: %d", comparison.GetStatus(), resp.StatusCode)
	}

	return CompareCommitResponse{
		CommitComparison: DiffDirection(comparison.GetStatus()),
	}, nil
}

type ListPRsRequest struct {
	Repo  internal.Repo
	State internal.PullRequestState
}

type ListPRsResponse struct {
	PullRequests []internal.PullRequest
}

func (a *githubActivities) ListPRs(ctx context.Context, request ListPRsRequest) (ListPRsResponse, error) {
	prs, err := a.Client.ListPullRequests(
		internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken),
		request.Repo.Owner,
		request.Repo.Name,
		request.Repo.DefaultBranch,
		PullRequestOpenState,
	)
	if err != nil {
		return ListPRsResponse{}, errors.Wrap(err, "listing open pull requests")
	}

	pullRequests := []internal.PullRequest{}
	for _, pullRequest := range prs {
		pullRequests = append(pullRequests, internal.PullRequest{
			Number: pullRequest.GetNumber(),
		})
	}

	return ListPRsResponse{
		PullRequests: pullRequests,
	}, nil
}

type ListModifiedFilesRequest struct {
	Repo        internal.Repo
	PullRequest internal.PullRequest
}

type ListModifiedFilesResponse struct {
	FilePaths []string
}

func (a *githubActivities) ListModifiedFiles(ctx context.Context, request ListModifiedFilesRequest) (ListModifiedFilesResponse, error) {
	files, err := a.Client.ListModifiedFiles(
		internal.ContextWithInstallationToken(ctx, request.Repo.Credentials.InstallationToken),
		request.Repo.Owner,
		request.Repo.Name,
		request.PullRequest.Number,
	)
	if err != nil {
		return ListModifiedFilesResponse{}, errors.Wrap(err, "listing modified files in pr")
	}

	filepaths := []string{}
	for _, file := range files {
		filepaths = append(filepaths, file.GetFilename())

		// account for previous file name as well if the file has moved across roots
		if file.GetStatus() == "renamed" {
			filepaths = append(filepaths, file.GetPreviousFilename())
		}
	}

	// strings are utf-8 encoded of size 1 to 4 bytes, assuming each file path is of length 100, max size of a filepath = 4 * 100 = 400 bytes
	// upper limit of 2Mb can accomodate (2*1024*1024)/400 = 524k filepaths which is >> max number of results supported by the GH API 3000.
	return ListModifiedFilesResponse{
		FilePaths: filepaths,
	}, nil
}
