package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestRunPipeline(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/trigger/pipeline", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "status":"pending"}`)
	})

	opt := &RunPipelineTriggerOptions{Ref: String("master")}
	pipeline, _, err := client.PipelineTriggers.RunPipelineTrigger(1, opt)

	if err != nil {
		t.Errorf("PipelineTriggers.RunPipelineTrigger returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "pending"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("PipelineTriggers.RunPipelineTrigger returned %+v, want %+v", pipeline, want)
	}
}
