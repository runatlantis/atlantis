package fetchers

import (
	"errors"
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

// @TODO test GithubFetcherConfig
func TestGithubFetcherConfig_FetchConfig(t *testing.T) {

}

func TestParseRemoteReference(t *testing.T) {
	testCases := []struct {
		Name      string
		Reference string
		Token     string
		Owner     string
		Repo      string
		Query     string
		Error     error
	}{
		{
			Name:      "Normal Scenario",
			Reference: "https://github.com/runatlantis/atlantis/atlantis.yaml?ref=master",
			Token:     "my-token",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     nil,
		},
		{
			Name:      "Empty Token",
			Reference: "https://github.com/runatlantis/atlantis/atlantis.yaml?ref=master",
			Token:     "",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     errors.New("github token is not set"),
		},
		{
			Name:      "Empty Owner",
			Reference: "https://github.com//atlantis/atlantis.yaml?ref=master",
			Token:     "my-token",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     errors.New("github owner is not set"),
		},
		{
			Name:      "Empty Repo",
			Reference: "https://github.com/runatlantis//atlantis.yaml?ref=master",
			Token:     "my-token",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     errors.New("github repo is not set"),
		},
		{
			Name:      "Empty Path",
			Reference: "https://github.com/runatlantis/atlantis/?ref=master",
			Token:     "my-token",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     errors.New("github path is not set"),
		},
		{
			Name:      "Incomplete Remote Reference",
			Reference: "https://github.com/runatlantis/atlantis",
			Token:     "my-token",
			Owner:     "runatlantis",
			Repo:      "atlantis",
			Query:     "master",
			Error:     errors.New("remote reference incomplete"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := GithubFetcherConfig{}
			err := ParseGithubReference(&config, tc.Reference, tc.Token)
			if tc.Error == nil {
				Equals(t, tc.Token, config.Token)
				Equals(t, tc.Owner, config.Owner)
				Equals(t, tc.Repo, config.Repo)
				Equals(t, tc.Query, config.Query)
				Equals(t, nil, err)
			} else {
				ErrContains(t, tc.Error.Error(), err)
			}
		})
	}
}

func TestDecodeContent(t *testing.T) {
	testCases := []struct {
		Name          string
		EncodedString string
		DecodedString string
		Error         error
	}{
		{
			Name:          "Normal Scenario",
			EncodedString: "SGVsbG8sIHdvcmxkIQ==",
			DecodedString: "Hello, world!",
			Error:         nil,
		},
		{
			Name:          "Empty String",
			EncodedString: "",
			DecodedString: "",
			Error:         nil,
		},
		{
			Name:          "Corrupted String",
			EncodedString: "sdsdsd",
			DecodedString: "",
			Error:         errors.New("could not decode file content"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			content, err := decodeContent(tc.EncodedString)
			if tc.Error == nil {
				Equals(t, nil, err)
				Equals(t, tc.DecodedString, content)
			} else {
				ErrContains(t, tc.Error.Error(), err)
			}
		})
	}
}
