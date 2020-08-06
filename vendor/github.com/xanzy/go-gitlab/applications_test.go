package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestCreateApplication(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/applications",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			fmt.Fprint(w, `
{
	"id":1,
	"application_name":"testApplication"
}`)
		},
	)

	opt := &CreateApplicationOptions{
		Name: String("testApplication"),
	}
	app, _, err := client.Applications.CreateApplication(opt)
	if err != nil {
		t.Errorf("Applications.CreateApplication returned error: %v", err)
	}

	want := &Application{
		ID:              1,
		ApplicationName: "testApplication",
	}
	if !reflect.DeepEqual(want, app) {
		t.Errorf("Applications.CreateApplication returned %+v, want %+v", app, want)
	}
}

func TestListApplications(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/applications",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[
	{"id":1},
	{"id":2}
]`)
		},
	)

	apps, _, err := client.Applications.ListApplications(&ListApplicationsOptions{})
	if err != nil {
		t.Errorf("Applications.ListApplications returned error: %v", err)
	}

	want := []*Application{
		{ID: 1},
		{ID: 2},
	}
	if !reflect.DeepEqual(want, apps) {
		t.Errorf("Applications.ListApplications returned %+v, want %+v", apps, want)
	}
}

func TestDeleteApplication(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/applications/4",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "DELETE")
			w.WriteHeader(http.StatusAccepted)
		},
	)

	resp, err := client.Applications.DeleteApplication(4)
	if err != nil {
		t.Errorf("Applications.DeleteApplication returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("Applications.DeleteApplication returned status code %d, want %d", got, want)
	}
}
