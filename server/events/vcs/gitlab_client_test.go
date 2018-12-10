package vcs

import (
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that the base url gets set properly.
func TestNewGitlabClient_BaseURL(t *testing.T) {
	gitlabClientUnderTest = true
	defer func() { gitlabClientUnderTest = false }()
	cases := []struct {
		Hostname   string
		ExpBaseURL string
	}{
		{
			"gitlab.com",
			"https://gitlab.com/api/v4/",
		},
		{
			"custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"http://custom.domain",
			"http://custom.domain/api/v4/",
		},
		{
			"http://custom.domain:8080",
			"http://custom.domain:8080/api/v4/",
		},
		{
			"https://custom.domain",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/",
			"https://custom.domain/api/v4/",
		},
		{
			"https://custom.domain/basepath/",
			"https://custom.domain/basepath/api/v4/",
		},
	}

	for _, c := range cases {
		t.Run(c.Hostname, func(t *testing.T) {
			client, err := NewGitlabClient(c.Hostname, "token", nil)
			Ok(t, err)
			Equals(t, c.ExpBaseURL, client.Client.BaseURL().String())
		})
	}
}

// This function gets called even if GitlabClient is nil
// so we need to test that.
func TestGitlabClient_SupportsCommonMarkNil(t *testing.T) {
	var gl *GitlabClient
	Equals(t, false, gl.SupportsCommonMark())
}

func TestGitlabClient_SupportsCommonMark(t *testing.T) {
	cases := []struct {
		version string
		exp     bool
	}{
		{
			"11.0",
			false,
		},
		{
			"11.1",
			true,
		},
		{
			"11.2",
			true,
		},
		{
			"12.0",
			true,
		},
	}

	for _, c := range cases {
		t.Run(c.version, func(t *testing.T) {
			vers, err := version.NewVersion(c.version)
			Ok(t, err)
			gl := GitlabClient{
				Version: vers,
			}
			Equals(t, c.exp, gl.SupportsCommonMark())
		})
	}
}
