package gitlab

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"testing"
)

func TestCreateLabel(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "name": "My Label", "color" : "#11FF22"}`)
	})

	// Create new label
	l := &CreateLabelOptions{
		Name:  String("My Label"),
		Color: String("#11FF22"),
	}
	label, _, err := client.Labels.CreateLabel("1", l)

	if err != nil {
		log.Fatal(err)
	}
	want := &Label{ID: 1, Name: "My Label", Color: "#11FF22"}
	if !reflect.DeepEqual(want, label) {
		t.Errorf("Labels.CreateLabel returned %+v, want %+v", label, want)
	}
}

func TestDeleteLabel(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	// Delete label
	label := &DeleteLabelOptions{
		Name: String("My Label"),
	}

	_, err := client.Labels.DeleteLabel("1", label)

	if err != nil {
		log.Fatal(err)
	}
}

func TestUpdateLabel(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{"id":1, "name": "New Label", "color" : "#11FF23" , "description":"This is updated label"}`)
	})

	// Update label
	l := &UpdateLabelOptions{
		Name:        String("My Label"),
		NewName:     String("New Label"),
		Color:       String("#11FF23"),
		Description: String("This is updated label"),
	}

	label, resp, err := client.Labels.UpdateLabel("1", l)

	if resp == nil {
		log.Fatal(err)
	}
	if err != nil {
		log.Fatal(err)
	}

	want := &Label{ID: 1, Name: "New Label", Color: "#11FF23", Description: "This is updated label"}

	if !reflect.DeepEqual(want, label) {
		t.Errorf("Labels.UpdateLabel returned %+v, want %+v", label, want)
	}
}

func TestSubscribeToLabel(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels/5/subscribe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{  "id" : 5, "name" : "bug", "color" : "#d9534f", "description": "Bug reported by user", "open_issues_count": 1, "closed_issues_count": 0, "open_merge_requests_count": 1, "subscribed": true,"priority": null}`)
	})

	label, _, err := client.Labels.SubscribeToLabel("1", "5")
	if err != nil {
		log.Fatal(err)
	}
	want := &Label{ID: 5, Name: "bug", Color: "#d9534f", Description: "Bug reported by user", OpenIssuesCount: 1, ClosedIssuesCount: 0, OpenMergeRequestsCount: 1, Subscribed: true}
	if !reflect.DeepEqual(want, label) {
		t.Errorf("Labels.SubscribeToLabel returned %+v, want %+v", label, want)
	}
}

func TestUnsubscribeFromLabel(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels/5/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
	})

	_, err := client.Labels.UnsubscribeFromLabel("1", "5")
	if err != nil {
		log.Fatal(err)
	}
}

func TestListLabels(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/labels", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{  "id" : 5, "name" : "bug", "color" : "#d9534f", "description": "Bug reported by user", "open_issues_count": 1, "closed_issues_count": 0, "open_merge_requests_count": 1, "subscribed": true,"priority": null}]`)
	})

	o := &ListLabelsOptions{
		Page:    1,
		PerPage: 10,
	}
	label, _, err := client.Labels.ListLabels("1", o)
	if err != nil {
		t.Log(err.Error() == "invalid ID type 1.1, the ID must be an int or a string")

	}
	want := []*Label{{ID: 5, Name: "bug", Color: "#d9534f", Description: "Bug reported by user", OpenIssuesCount: 1, ClosedIssuesCount: 0, OpenMergeRequestsCount: 1, Subscribed: true}}
	if !reflect.DeepEqual(want, label) {
		t.Errorf("Labels.ListLabels returned %+v, want %+v", label, want)
	}
}
