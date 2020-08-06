package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListProjectPipelines(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"id":1},{"id":2}]`)
	})

	opt := &ListProjectPipelinesOptions{Ref: String("master")}
	piplines, _, err := client.Pipelines.ListProjectPipelines(1, opt)
	if err != nil {
		t.Errorf("Pipelines.ListProjectPipelines returned error: %v", err)
	}

	want := []*PipelineInfo{{ID: 1}, {ID: 2}}
	if !reflect.DeepEqual(want, piplines) {
		t.Errorf("Pipelines.ListProjectPipelines returned %+v, want %+v", piplines, want)
	}
}

func TestGetPipeline(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":1,"status":"success"}`)
	})

	pipeline, _, err := client.Pipelines.GetPipeline(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.GetPipeline returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "success"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.GetPipeline returned %+v, want %+v", pipeline, want)
	}
}

func TestGetPipelineVariables(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/variables", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"key":"RUN_NIGHTLY_BUILD","variable_type":"env_var","value":"true"},{"key":"foo","value":"bar"}]`)
	})

	variables, _, err := client.Pipelines.GetPipelineVariables(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.GetPipelineVariables returned error: %v", err)
	}

	want := []*PipelineVariable{{Key: "RUN_NIGHTLY_BUILD", Value: "true", VariableType: "env_var"}, {Key: "foo", Value: "bar"}}
	if !reflect.DeepEqual(want, variables) {
		t.Errorf("Pipelines.GetPipelineVariables returned %+v, want %+v", variables, want)
	}
}

func TestCreatePipeline(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipeline", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprint(w, `{"id":1, "status":"pending"}`)
	})

	opt := &CreatePipelineOptions{Ref: String("master")}
	pipeline, _, err := client.Pipelines.CreatePipeline(1, opt)

	if err != nil {
		t.Errorf("Pipelines.CreatePipeline returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "pending"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.CreatePipeline returned %+v, want %+v", pipeline, want)
	}
}

func TestRetryPipelineBuild(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/retry", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprintln(w, `{"id":1, "status":"pending"}`)
	})

	pipeline, _, err := client.Pipelines.RetryPipelineBuild(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.RetryPipelineBuild returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "pending"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.RetryPipelineBuild returned %+v, want %+v", pipeline, want)
	}
}

func TestCancelPipelineBuild(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167/cancel", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprintln(w, `{"id":1, "status":"canceled"}`)
	})

	pipeline, _, err := client.Pipelines.CancelPipelineBuild(1, 5949167)
	if err != nil {
		t.Errorf("Pipelines.CancelPipelineBuild returned error: %v", err)
	}

	want := &Pipeline{ID: 1, Status: "canceled"}
	if !reflect.DeepEqual(want, pipeline) {
		t.Errorf("Pipelines.CancelPipelineBuild returned %+v, want %+v", pipeline, want)
	}
}

func TestDeletePipeline(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/1/pipelines/5949167", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.Pipelines.DeletePipeline("1", 5949167)

	if err != nil {
		t.Errorf("Pipelines.DeletePipeline returned error: %v", err)
	}

}
