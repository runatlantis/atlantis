package fetchers

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

type GithubFetcherConfig struct {
	Repo  string
	Owner string
	Path  string
	Query string
	Token string
}

func (c *GithubFetcherConfig) FetchConfig() (string, error) {
	ctx := context.Background()

	// Get an authenticated github client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: c.Token},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// construct the get options from the query segment of the fetcher config
	opt := github.RepositoryContentGetOptions{Ref: c.Query}

	// Get the actual file contents.
	// We are only interested in a file content, not the result of a directory listing.
	fileContent, _, _, err := client.Repositories.GetContents(ctx, c.Owner, c.Repo, c.Path, &opt)
	if err != nil {
		return "", errors.Wrap(err, "could not fetch content")
	}

	// Bae64 decode the response's body.
	configString, err := decodeContent(*fileContent.Content)
	if err != nil {
		return "", errors.New("error decoding config file content")
	}

	// We got config!
	return configString, nil
}

func ParseGithubReference(c *GithubFetcherConfig, remoteReference string, token string) error {
	u, err := url.Parse(remoteReference)
	if err != nil {
		return errors.Wrap(err, "remote reference syntax error")
	}

	// Parse the path part of the URL.
	// The most minimal Remote Reference for github would look like this:
	// `http://github.com/owner/repo/atlantis.yaml`
	//
	// When it is parsed using url.Parse(), the Path() function returns this
	// `/owner/repo/atlantis.yaml`
	//
	// If we split it by `/`, then the first element in the slice is going to be empty, given that the first character
	// is `/`, and the minimum number of elements we should get is 4, that includes the owner, the repo, and the
	// atlantis file name.
	p := strings.Split(u.Path, "/")
	if len(p) < 4 {
		return errors.New("remote reference incomplete")
	}

	c.Token = token
	c.Owner = p[1]
	c.Repo = p[2]
	c.Path = strings.Join(p[3:], "/")

	// Pick the first ref, if it exists. Users should be passing more than one ref anyway.
	refs, ok := u.Query()["ref"]
	if ok {
		c.Query = refs[0]
	}

	// Validate the fetcher config before moving forward
	return validateGithubFetcherConfig(c)
}

func validateGithubFetcherConfig(config *GithubFetcherConfig) error {
	// GithubFetcherConfig.Query field can be empty, so we don't validate it here.

	if config.Owner == "" {
		return errors.New("github owner is not set")
	}

	if config.Repo == "" {
		return errors.New("github repo is not set")
	}

	if config.Path == "" {
		return errors.New("github path is not set")
	}

	if config.Token == "" {
		return errors.New("github token is not set")
	}

	return nil
}

func decodeContent(content interface{}) (string, error) {
	decodedContent, err := base64.StdEncoding.DecodeString(content.(string))
	if err != nil {
		err = errors.New("could not decode file content")
		return "", err
	}

	return string(decodedContent), nil
}
