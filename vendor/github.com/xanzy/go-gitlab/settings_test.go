package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetSettings(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1,    "default_projects_limit" : 100000}`)
	})

	settings, _, err := client.Settings.GetSettings()
	if err != nil {
		t.Fatal(err)
	}

	want := &Settings{ID: 1, DefaultProjectsLimit: 100000}
	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Settings.GetSettings returned %+v, want %+v", settings, want)
	}
}

func TestUpdateSettings(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/application/settings", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{"default_projects_limit" : 100}`)
	})

	options := &UpdateSettingsOptions{
		DefaultProjectsLimit: Int(100),
	}
	settings, _, err := client.Settings.UpdateSettings(options)
	if err != nil {
		t.Fatal(err)
	}

	want := &Settings{DefaultProjectsLimit: 100}
	if !reflect.DeepEqual(settings, want) {
		t.Errorf("Settings.UpdateSettings returned %+v, want %+v", settings, want)
	}
}
