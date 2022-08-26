package vcs

import (
	"context"

	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GithubGraphQLClient struct {
	credentials GithubCredentials
	graphqlURL  string
}

func (gcl GithubGraphQLClient) Query(ctx context.Context, q interface{}, variables map[string]interface{}) error {
	token, err := gcl.credentials.GetToken()
	if err != nil {
		return errors.Wrap(err, "Failed to get GitHub token")
	}
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	// Use the client from shurcooL's githubv4 library for queries.
	v4QueryClient := githubv4.NewEnterpriseClient(gcl.graphqlURL, httpClient)

	err = v4QueryClient.Query(ctx, q, variables)
	if err != nil {
		return errors.Wrap(err, "making graphql query")
	}

	return nil
}
