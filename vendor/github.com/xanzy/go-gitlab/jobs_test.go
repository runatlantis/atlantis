package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListPipelineJobs(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/1/jobs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	jobs, _, err := client.Jobs.ListPipelineJobs(1, 1, nil)
	if err != nil {
		t.Errorf("Jobs.ListPipelineJobs returned error: %v", err)
	}

	want := []*Job{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(want, jobs) {
		t.Errorf("Jobs.ListPipelineJobs returned %+v, want %+v", jobs, want)
	}
}
