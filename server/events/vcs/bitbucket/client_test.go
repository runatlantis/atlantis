package bitbucket_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs/bitbucket"
	. "github.com/runatlantis/atlantis/testing"
)

// Should follow pagination properly.
func TestClient_GetModifiedFilesPagination(t *testing.T) {
	respTemplate := `
{
    "pagelen": 1,
    "values": [
        {
            "type": "diffstat",
            "status": "modified",
            "lines_removed": 1,
            "lines_added": 2,
            "old": {
                "path": "%s",
                "type": "commit_file",
                "links": {
                    "self": {
                        "href": "https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/e1749643d655d7c7014001a6c0f58abaf42ad850/setup.py"
                    }
                }
            },
            "new": {
                "path": "%s",
                "type": "commit_file",
                "links": {
                    "self": {
                        "href": "https://api.bitbucket.org/2.0/repositories/bitbucket/geordi/src/d222fa235229c55dad20b190b0b571adf737d5a6/setup.py"
                    }
                }
            }
        }
    ],
    "page": 1,
    "size": 1
`
	firstResp := fmt.Sprintf(respTemplate, "file1.txt", "file2.txt")
	secondResp := fmt.Sprintf(respTemplate, "file2.txt", "file3.txt")
	var serverURL string

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/2.0/repositories/owner/repo/pullrequests/1/diffstat":
			resp := firstResp + fmt.Sprintf(`,"next": "%s/2.0/repositories/owner/repo/pullrequests/1/diffstat?page=2"}`, serverURL)
			w.Write([]byte(resp)) // nolint: errcheck
			return
			// The second should hit this URL.
		case "/2.0/repositories/owner/repo/pullrequests/1/diffstat?page=2":
			w.Write([]byte(secondResp + "}")) // nolint: errcheck
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	serverURL = testServer.URL
	client, err := bitbucket.NewClient(http.DefaultClient, "user", "pass", serverURL, "runatlantis.io")
	Ok(t, err)

	files, err := client.GetModifiedFiles(models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Bitbucket,
			Hostname: "bitbucket.org",
		},
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"file1.txt", "file2.txt", "file3.txt"}, files)
}

// If the "old" key in the list of files is nil we shouldn't error.
func TestClient_GetModifiedFilesOldNil(t *testing.T) {
	resp := `
{
  "pagelen": 500,
  "values": [
    {
      "status": "added",
      "old": null,
      "lines_removed": 0,
      "lines_added": 2,
      "new": {
        "path": "parent/child/file1.txt",
        "type": "commit_file",
        "links": {
          "self": {
            "href": "https://api.bitbucket.org/2.0/repositories/lkysow/atlantis-example/src/1ed8205eec00dab4f1c0a8c486a4492c98c51f8e/main.tf"
          }
        }
      },
      "type": "diffstat"
    }
  ],
  "page": 1,
  "size": 1
}`

	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		// The first request should hit this URL.
		case "/2.0/repositories/owner/repo/pullrequests/1/diffstat":
			w.Write([]byte(resp)) // nolint: errcheck
			return
		default:
			t.Errorf("got unexpected request at %q", r.RequestURI)
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
	}))
	defer testServer.Close()

	client, err := bitbucket.NewClient(http.DefaultClient, "user", "pass", testServer.URL, "runatlantis.io")
	Ok(t, err)

	files, err := client.GetModifiedFiles(models.Repo{
		FullName:          "owner/repo",
		Owner:             "owner",
		Name:              "repo",
		CloneURL:          "",
		SanitizedCloneURL: "",
		VCSHost: models.VCSHost{
			Type:     models.Bitbucket,
			Hostname: "bitbucket.org",
		},
	}, models.PullRequest{
		Num: 1,
	})
	Ok(t, err)
	Equals(t, []string{"parent/child/file1.txt"}, files)
}
