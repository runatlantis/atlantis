package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetCommitStatuses(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/projects/1/repository/commits/b0b3a907f41409829b307a28b82fdbd552ee5a27/statuses", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1}]`)
	})

	opt := &GetCommitStatusesOptions{String("master"), String("test"), String("ci/jenkins"), Bool(true)}
	statuses, _, err := client.Commits.GetCommitStatuses("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", opt)

	if err != nil {
		t.Errorf("Commits.GetCommitStatuses returned error: %v", err)
	}

	want := []*CommitStatus{{ID: 1}}
	if !reflect.DeepEqual(want, statuses) {
		t.Errorf("Commits.GetCommitStatuses returned %+v, want %+v", statuses, want)
	}
}

func TestSetCommitStatus(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/projects/1/statuses/b0b3a907f41409829b307a28b82fdbd552ee5a27", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1}`)
	})

	opt := &SetCommitStatusOptions{Running, String("master"), String("ci/jenkins"), String(""), String("http://abc"), String("build")}
	status, _, err := client.Commits.SetCommitStatus("1", "b0b3a907f41409829b307a28b82fdbd552ee5a27", opt)

	if err != nil {
		t.Errorf("Commits.SetCommitStatus returned error: %v", err)
	}

	want := &CommitStatus{ID: 1}
	if !reflect.DeepEqual(want, status) {
		t.Errorf("Commits.SetCommitStatus returned %+v, want %+v", status, want)
	}
}
