package gitlab

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

const exampleReleaseListRsp = `[
        {
          "tag_name": "v0.2",
          "description": "description",
          "name": "Awesome app v0.2 beta",
          "description_html": "html",
          "created_at": "2019-01-03T01:56:19.539Z",
          "author": {
            "id": 1,
            "name": "Administrator",
            "username": "root",
            "state": "active",
            "avatar_url": "https://www.gravatar.com/avatar",
            "web_url": "http://localhost:3000/root"
          },
          "commit": {
            "id": "079e90101242458910cccd35eab0e211dfc359c0",
            "short_id": "079e9010",
            "title": "Update README.md",
            "created_at": "2019-01-03T01:55:38.000Z",
            "parent_ids": [
              "f8d3d94cbd347e924aa7b715845e439d00e80ca4"
            ],
            "message": "Update README.md",
            "author_name": "Administrator",
            "author_email": "admin@example.com",
            "authored_date": "2019-01-03T01:55:38.000Z",
            "committer_name": "Administrator",
            "committer_email": "admin@example.com",
            "committed_date": "2019-01-03T01:55:38.000Z"
          },
          "assets": {
            "count": 4,
            "sources": [
              {
                "format": "zip",
                "url": "http://localhost:3000/archive/v0.2/awesome-app-v0.2.zip"
              },
              {
                "format": "tar.gz",
                "url": "http://localhost:3000/archive/v0.2/awesome-app-v0.2.tar.gz"
              }
            ],
            "links": [
              {
                "id": 2,
                "name": "awesome-v0.2.msi",
                "url": "http://192.168.10.15:3000/msi",
                "external": true
              },
              {
                "id": 1,
                "name": "awesome-v0.2.dmg",
                "url": "http://192.168.10.15:3000",
                "external": true
              }
            ]
          }
        },
        {
          "tag_name": "v0.1",
          "description": "description",
          "name": "Awesome app v0.1 alpha",
          "description_html": "description_html",
          "created_at": "2019-01-03T01:55:18.203Z",
          "author": {
            "id": 1,
            "name": "Administrator",
            "username": "root",
            "state": "active",
            "avatar_url": "https://www.gravatar.com/avatar",
            "web_url": "http://localhost:3000/root"
          },
          "commit": {
            "id": "f8d3d94cbd347e924aa7b715845e439d00e80ca4",
            "short_id": "f8d3d94c",
            "title": "Initial commit",
            "created_at": "2019-01-03T01:53:28.000Z",
            "parent_ids": [],
            "message": "Initial commit",
            "author_name": "Administrator",
            "author_email": "admin@example.com",
            "authored_date": "2019-01-03T01:53:28.000Z",
            "committer_name": "Administrator",
            "committer_email": "admin@example.com",
            "committed_date": "2019-01-03T01:53:28.000Z"
          },
          "assets": {
            "count": 2,
            "sources": [
              {
                "format": "zip",
                "url": "http://localhost:3000/archive/v0.1/awesome-app-v0.1.zip"
              },
              {
                "format": "tar.gz",
                "url": "http://localhost:3000/archive/v0.1/awesome-app-v0.1.tar.gz"
              }
            ],
            "links": []
          }
        }
]`

func TestReleasesService_ListReleases(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, exampleReleaseListRsp)
		})

	opt := &ListReleasesOptions{}
	releases, _, err := client.Releases.ListReleases(1, opt)
	if err != nil {
		t.Error(err)
	}
	if len(releases) != 2 {
		t.Error("expected 2 releases")
	}
}

const exampleReleaseRsp = `{
        "tag_name": "v0.1",
        "description": "description",
        "name": "Awesome app v0.1 alpha",
        "description_html": "description_html",
        "created_at": "2019-01-03T01:55:18.203Z",
        "author": {
          "id": 1,
          "name": "Administrator",
          "username": "root",
          "state": "active",
          "avatar_url": "https://www.gravatar.com/avatar/",
          "web_url": "http://localhost:3000/root"
        },
        "commit": {
          "id": "f8d3d94cbd347e924aa7b715845e439d00e80ca4",
          "short_id": "f8d3d94c",
          "title": "Initial commit",
          "created_at": "2019-01-03T01:53:28.000Z",
          "parent_ids": [],
          "message": "Initial commit",
          "author_name": "Administrator",
          "author_email": "admin@example.com",
          "authored_date": "2019-01-03T01:53:28.000Z",
          "committer_name": "Administrator",
          "committer_email": "admin@example.com",
          "committed_date": "2019-01-03T01:53:28.000Z"
        },
        "assets": {
          "count": 2,
          "sources": [
            {
              "format": "zip",
              "url": "http://localhost:3000/archive/v0.1/awesome-app-v0.1.zip"
            },
            {
              "format": "tar.gz",
              "url": "http://localhost:3000/archive/v0.1/awesome-app-v0.1.tar.gz"
            }
          ],
          "links": []
        }
}`

func TestReleasesService_GetRelease(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, exampleReleaseRsp)
		})

	release, _, err := client.Releases.GetRelease(1, "v0.1")
	if err != nil {
		t.Error(err)
	}
	if release.TagName != "v0.1" {
		t.Errorf("expected tag v0.1, got %s", release.TagName)
	}
}

func TestReleasesService_CreateRelease(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("unable to read request body")
			}
			if !strings.Contains(string(b), "v0.1") {
				t.Errorf("expected request body to contain v0.1, got %s",
					string(b))
			}
			if strings.Contains(string(b), "assets") {
				t.Errorf("expected request body not to have assets, got %s",
					string(b))
			}
			fmt.Fprint(w, exampleReleaseRsp)
		})

	opts := &CreateReleaseOptions{
		Name:        String("name"),
		TagName:     String("v0.1"),
		Description: String("Description"),
	}

	release, _, err := client.Releases.CreateRelease(1, opts)
	if err != nil {
		t.Error(err)
	}
	if release.TagName != "v0.1" {
		t.Errorf("expected tag v0.1, got %s", release.TagName)
	}
}

func TestReleasesService_CreateReleaseWithAsset(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			b, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("unable to read request body")
			}
			if !strings.Contains(string(b), "v0.1") {
				t.Errorf("expected request body to contain v0.1, got %s",
					string(b))
			}
			if !strings.Contains(string(b), "assets") {
				t.Errorf("expected request body to have assets, got %s",
					string(b))
			}
			fmt.Fprint(w, exampleReleaseRsp)
		})

	opts := &CreateReleaseOptions{
		Name:        String("name"),
		TagName:     String("v0.1"),
		Description: String("Description"),
		Assets: &ReleaseAssets{
			Links: []*ReleaseAssetLink{
				{"sldkf", "sldkfj"},
			},
		},
	}

	release, _, err := client.Releases.CreateRelease(1, opts)
	if err != nil {
		t.Error(err)
	}
	if release.TagName != "v0.1" {
		t.Errorf("expected tag v0.1, got %s", release.TagName)
	}
}

func TestReleasesService_UpdateRelease(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "PUT")
			fmt.Fprint(w, exampleReleaseRsp)
		})

	opts := &UpdateReleaseOptions{
		Name:        String("name"),
		Description: String("Description"),
	}

	release, _, err := client.Releases.UpdateRelease(1, "v0.1", opts)
	if err != nil {
		t.Error(err)
	}
	if release.TagName != "v0.1" {
		t.Errorf("expected tag v0.1, got %s", release.TagName)
	}

}

func TestReleasesService_DeleteRelease(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "DELETE")
			fmt.Fprint(w, exampleReleaseRsp)
		})

	release, _, err := client.Releases.DeleteRelease(1, "v0.1")
	if err != nil {
		t.Error(err)
	}
	if release.TagName != "v0.1" {
		t.Errorf("expected tag v0.1, got %s", release.TagName)
	}

}
