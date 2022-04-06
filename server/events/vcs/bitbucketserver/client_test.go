package bitbucketserver_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucketserver"
	. "github.com/runatlantis/atlantis/testing"
)

// Test that we include the base path in our base url.
func TestClient_BasePath(t *testing.T) {
	cases := []struct {
		inputURL string
		expURL   string
		expErr   string
	}{
		{
			inputURL: "mycompany.com",
			expErr:   `must have 'http://' or 'https://' in base url "mycompany.com"`,
		},
		{
			inputURL: "https://mycompany.com",
			expURL:   "https://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com",
			expURL:   "http://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com:7990",
			expURL:   "http://mycompany.com:7990",
		},
		{
			inputURL: "http://mycompany.com/",
			expURL:   "http://mycompany.com",
		},
		{
			inputURL: "http://mycompany.com:7990/",
			expURL:   "http://mycompany.com:7990",
		},
		{
			inputURL: "http://mycompany.com/basepath/",
			expURL:   "http://mycompany.com/basepath",
		},
		{
			inputURL: "http://mycompany.com:7990/basepath/",
			expURL:   "http://mycompany.com:7990/basepath",
		},
	}

	for _, c := range cases {
		t.Run(c.inputURL, func(t *testing.T) {
			client, err := bitbucketserver.NewClient(nil, "u", "p", c.inputURL, "atlantis-url")
			if c.expErr != "" {
				ErrEquals(t, c.expErr, err)
			} else {
				Ok(t, err)
				Equals(t, c.expURL, client.BaseURL)
			}
		})
	}
}

// Should follow pagination properly.
func TestClient_GetModifiedFilesPagination(t *testing.T) {
	respTemplate := `
{
  "values": [
    {
      "path": {
        "toString": "%s"
      }
    },
    {
      "path": {
        "toString": "%s"
      }
    }
  ],
  "size": 2,
  "isLastPage": true,
  "start": 0,
  "limit": 2,
  "nextPageStart": null
}
`
	firstResp := fmt.Sprintf(respTemplate, "file1.txt", "file2.txt")
	secondResp := fmt.Sprintf(respTemplate, "file2.txt", "file3.txt")
	var serverURL string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/changes?start=0":
			resp := strings.Replace(firstResp, `"isLastPage": true`, `"isLastPage": false`, -1)
			resp = strings.Replace(resp, `"nextPageStart": null`, `"nextPageStart": 3`, -1)
			w.Write([]byte(resp)) // nolint: errcheck
			return
			// The second should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/changes?start=3":
			w.Write([]byte(secondResp)) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	serverURL = testServer.URL
	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", serverURL, "runatlantis.io")
	Ok(t, err)

	files, err := client.GetModifiedFiles(models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", serverURL),
		VCSHost: models.VCSHost{
			Type:     models.BitbucketCloud,
			Hostname: "bitbucket.org",
		},
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"file1.txt", "file2.txt", "file3.txt"}, files)
}

func TestClient_MarkdownPullLink(t *testing.T) {
	client, err := bitbucketserver.NewClient(nil, "u", "p", "https://base-url", "atlantis-url")
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "#1"
	Equals(t, exp, s)
}
