package activities

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
)

const (
	SetRevisionReason     = "new changes deployed to root modified in this PR"
	SetRevisionMethodType = "POST"
	SetRevisionEndpoint   = "set_minimum_service_pr_revision"
)

type revisionSetterClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type NoopClient struct{}

func (n *NoopClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		Body:       http.NoBody,
		StatusCode: 200,
	}, nil
}

type prRevisionSetterActivities struct {
	client revisionSetterClient

	url       string
	basicAuth valid.BasicAuth
}

type SetPRRevisionRequest struct {
	Repository  github.Repo
	PullRequest github.PullRequest
	Revision    string
}

func generateURL(url string, request SetPRRevisionRequest) string {
	return fmt.Sprintf(
		"https://%s/%s/%s/%d/%s/%s",
		url,
		SetRevisionEndpoint,
		request.Repository.Name,
		request.PullRequest.Number,
		request.Revision,
		SetRevisionReason,
	)
}

func (b *prRevisionSetterActivities) SetPRRevision(ctx context.Context, request SetPRRevisionRequest) error {
	url := generateURL(b.url, request)
	req, err := http.NewRequestWithContext(ctx, SetRevisionMethodType, url, nil)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}

	// add basic auth credentials
	req.SetBasicAuth(b.basicAuth.Username, b.basicAuth.Password)
	response, err := b.client.Do(req)
	if err != nil {
		return errors.Wrap(err, "setting PR revision")
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		bytes, err := io.ReadAll(response.Body)
		if err != nil {
			return errors.Wrap(err, "reading response body")
		}

		return fmt.Errorf("setting PR revision: %s", string(bytes))
	}
	return nil
}
