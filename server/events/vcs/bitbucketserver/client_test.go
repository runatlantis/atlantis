package bitbucketserver_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

// Test that we use the correct version parameter in our call to merge the pull
// request.
func TestClient_MergePull(t *testing.T) {
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-request.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge?version=3":
			Equals(t, "POST", r.Method)
			w.Write(pullRequest) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)

	err = client.MergePull(models.PullRequest{
		Num:        1,
		HeadCommit: "",
		URL:        "",
		HeadBranch: "",
		BaseBranch: "",
		Author:     "",
		State:      0,
		BaseRepo: models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
			VCSHost: models.VCSHost{
				Type:     models.BitbucketCloud,
				Hostname: "bitbucket.org",
			},
		},
	}, models.PullRequestOptions{
		DeleteSourceBranchOnMerge: false,
	})
	Ok(t, err)
}

// Test that we delete the source branch in our call to merge the pull
// request.
func TestClient_MergePullDeleteSourceBranch(t *testing.T) {
	pullRequest, err := os.ReadFile(filepath.Join("testdata", "pull-request.json"))
	Ok(t, err)
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1":
			w.Write(pullRequest) // nolint: errcheck
			return
		case "/rest/api/1.0/projects/ow/repos/repo/pull-requests/1/merge?version=3":
			Equals(t, "POST", r.Method)
			w.Write(pullRequest) // nolint: errcheck
		case "/rest/branch-utils/1.0/projects/ow/repos/repo/branches":
			Equals(t, "DELETE", r.Method)
			defer r.Body.Close()
			b, err := io.ReadAll(r.Body)
			Ok(t, err)
			var payload bitbucketserver.DeleteSourceBranch
			err = json.Unmarshal(b, &payload)
			Ok(t, err)
			Equals(t, "refs/heads/foo", payload.Name)
			w.WriteHeader(http.StatusNoContent) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucketserver.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)

	err = client.MergePull(models.PullRequest{
		Num:        1,
		HeadCommit: "",
		URL:        "",
		HeadBranch: "foo",
		BaseBranch: "",
		Author:     "",
		State:      0,
		BaseRepo: models.Repo{
			FullName:          "owner/repo",
			Owner:             "owner",
			Name:              "repo",
			SanitizedCloneURL: fmt.Sprintf("%s/scm/ow/repo.git", testServer.URL),
			VCSHost: models.VCSHost{
				Type:     models.BitbucketServer,
				Hostname: "bitbucket.org",
			},
		},
	}, models.PullRequestOptions{
		DeleteSourceBranchOnMerge: true,
	})
	Ok(t, err)
}

func TestClient_MarkdownPullLink(t *testing.T) {
	client, err := bitbucketserver.NewClient(nil, "u", "p", "https://base-url", "atlantis-url")
	Ok(t, err)
	pull := models.PullRequest{Num: 1}
	s, _ := client.MarkdownPullLink(pull)
	exp := "#1"
	Equals(t, exp, s)
}
