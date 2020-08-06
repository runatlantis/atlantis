package gitlab

import (
	"fmt"
	"net/http"
	"testing"
)

const exampleReleaseLinkList = `[
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
]`

func TestReleaseLinksService_ListReleaseLinks(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1/assets/links",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, exampleReleaseLinkList)
		})

	releaseLinks, _, err := client.ReleaseLinks.ListReleaseLinks(
		1, "v0.1", &ListReleaseLinksOptions{},
	)
	if err != nil {
		t.Error(err)
	}
	if len(releaseLinks) != 2 {
		t.Error("expected 2 links")
	}
	if releaseLinks[0].Name != "awesome-v0.2.msi" {
		t.Errorf("release link name, expected '%s', got '%s'", "awesome-v0.2.msi",
			releaseLinks[0].Name)
	}
}

const exampleReleaseLink = `{
        "id":1,
        "name":"awesome-v0.2.dmg",
        "url":"http://192.168.10.15:3000",
        "external":true
 }`

func TestReleaseLinksService_CreateReleaseLink(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1/assets/links",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			fmt.Fprint(w, exampleReleaseLink)
		})

	releaseLink, _, err := client.ReleaseLinks.CreateReleaseLink(
		1, "v0.1",
		&CreateReleaseLinkOptions{
			Name: String("awesome-v0.2.dmg"),
			URL:  String("http://192.168.10.15:3000"),
		})
	if err != nil {
		t.Error(err)
	}
	if releaseLink.Name != "awesome-v0.2.dmg" {
		t.Errorf("release link name, expected '%s', got '%s'", "awesome-v0.2.dmg",
			releaseLink.Name)
	}
}

func TestReleaseLinksService_GetReleaseLink(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1/assets/links/1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, exampleReleaseLink)
		})

	releaseLink, _, err := client.ReleaseLinks.GetReleaseLink(1, "v0.1", 1)
	if err != nil {
		t.Error(err)
	}
	if releaseLink.Name != "awesome-v0.2.dmg" {
		t.Errorf("release link name, expected '%s', got '%s'", "awesome-v0.2.dmg",
			releaseLink.Name)
	}
}

func TestReleaseLinksService_UpdateReleaseLink(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1/assets/links/1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "PUT")
			fmt.Fprint(w, exampleReleaseLink)
		})

	releaseLink, _, err := client.ReleaseLinks.UpdateReleaseLink(
		1, "v0.1", 1,
		&UpdateReleaseLinkOptions{
			Name: String("awesome-v0.2.dmg"),
		})
	if err != nil {
		t.Error(err)
	}
	if releaseLink.Name != "awesome-v0.2.dmg" {
		t.Errorf("release link name, expected '%s', got '%s'", "awesome-v0.2.dmg",
			releaseLink.Name)
	}
}

func TestReleaseLinksService_DeleteReleaseLink(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/releases/v0.1/assets/links/1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "DELETE")
			fmt.Fprint(w, exampleReleaseLink)
		})

	releaseLink, _, err := client.ReleaseLinks.DeleteReleaseLink(1, "v0.1", 1)
	if err != nil {
		t.Error(err)
	}
	if releaseLink.Name != "awesome-v0.2.dmg" {
		t.Errorf("release link name, expected '%s', got '%s'", "awesome-v0.2.dmg",
			releaseLink.Name)
	}
}
