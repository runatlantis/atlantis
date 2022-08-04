package activities

import (
	"context"
	"strings"

	"github.com/google/go-github/v45/github"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
)

type VCSClientWrapper struct {
	client      *github.Client
	eventParser *events.EventParser
}

func NewVCSClientWrapper(githubUser string, githubToken string) *VCSClientWrapper {
	eventParser := &events.EventParser{
		GithubUser:  githubUser,
		GithubToken: githubToken,
	}

	transport := &github.BasicAuthTransport{
		Username: strings.TrimSpace(githubUser),
		Password: strings.TrimSpace(githubToken),
	}

	client := github.NewClient(transport.Client())

	return &VCSClientWrapper{
		client:      client,
		eventParser: eventParser,
	}
}

type GetRepositoryRequest struct {
	Owner string
	Repo  string
}

type GetRepositoryResponse struct {
	Repo models.Repo
}

func (r *VCSClientWrapper) GetRepository(ctx context.Context, request GetRepositoryRequest) (GetRepositoryResponse, error) {
	rawRepo, _, err := r.client.Repositories.Get(ctx, request.Owner, request.Repo)

	if err != nil {
		return GetRepositoryResponse{}, errors.Wrapf(err, "getting github repo %s/%s", request.Owner, request.Repo)
	}

	repository, err := r.eventParser.ParseGithubRepo(rawRepo)

	if err != nil {
		return GetRepositoryResponse{}, errors.Wrapf(err, "parsing github repo %s from response", *rawRepo.Name)
	}

	return GetRepositoryResponse{Repo: repository}, err
}
