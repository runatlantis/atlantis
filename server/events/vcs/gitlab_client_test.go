package vcs

import (
	. "github.com/runatlantis/atlantis/testing"
	"testing"
)

// Test that the base url gets set properly.
func TestNewGitlabClient_BaseURL(t *testing.T) {
	cases := []struct {
		Hostname   string
		ExpBaseURL string
	}{
		{
			"gitlab.com",
			"https://gitlab.com/api/v4/",
		},
		{
			"http://custom.domain",
			"http://custom.domain/api/v4/",
		},
		{
			"https://custom.domain",
			"https://custom.domain/api/v4/",
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
