package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListFeatureFlags(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/features", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `
		[
			{
			  "name": "experimental_feature",
			  "state": "off",
			  "gates": [
				{
				  "key": "boolean",
				  "value": false
				}
			  ]
			},
			{
			  "name": "new_library",
			  "state": "on"
			}
		  ]
	`)
	})

	features, _, err := client.Features.ListFeatures()
	if err != nil {
		t.Errorf("Features.ListFeatures returned error: %v", err)
	}

	want := []*Feature{
		{Name: "experimental_feature", State: "off", Gates: []Gate{
			{Key: "boolean", Value: false},
		}},
		{Name: "new_library", State: "on"},
	}
	if !reflect.DeepEqual(want, features) {
		t.Errorf("Features.ListFeatures returned %+v, want %+v", features, want)
	}
}

func TestSetFeatureFlag(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/features/new_library", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `
		{
			"name": "new_library",
			"state": "conditional",
			"gates": [
			  {
				"key": "boolean",
				"value": false
			  },
			  {
				"key": "percentage_of_time",
				"value": 30
			  }
			]
		  }
		`)
	})

	feature, _, err := client.Features.SetFeatureFlag("new_library", "30")
	if err != nil {
		t.Errorf("Features.SetFeatureFlag returned error: %v", err)
	}

	want := &Feature{
		Name:  "new_library",
		State: "conditional",
		Gates: []Gate{
			{Key: "boolean", Value: false},
			{Key: "percentage_of_time", Value: 30.0},
		},
	}
	if !reflect.DeepEqual(want, feature) {
		t.Errorf("Features.SetFeatureFlag returned %+v, want %+v", feature, want)
	}
}
