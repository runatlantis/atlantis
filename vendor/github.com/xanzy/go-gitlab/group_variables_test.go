package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListGroupVariabless(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `[{"key": "TEST_VARIABLE_1","value": "test1","protected": false,"masked": true}]`)
		})

	variables, _, err := client.GroupVariables.ListVariables(1, &ListGroupVariablesOptions{})
	if err != nil {
		t.Errorf("GroupVariables.ListVariables returned error: %v", err)
	}

	want := []*GroupVariable{
		{
			Key:       "TEST_VARIABLE_1",
			Value:     "test1",
			Protected: false,
			Masked:    true,
		},
	}

	if !reflect.DeepEqual(want, variables) {
		t.Errorf("GroupVariables.ListVariablesreturned %+v, want %+v", variables, want)
	}
}

func TestGetGroupVariable(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `{"key": "TEST_VARIABLE_1","value": "test1","protected": false,"masked": true}`)
		})

	variable, _, err := client.GroupVariables.GetVariable(1, "TEST_VARIABLE_1")

	if err != nil {
		t.Errorf("GroupVariables.GetVariable returned error: %v", err)
	}

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true}
	if !reflect.DeepEqual(want, variable) {
		t.Errorf("GroupVariables.GetVariable returned %+v, want %+v", variable, want)
	}
}

func TestCreateGroupVariable(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/variables",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "POST")
			fmt.Fprint(w, `{"key": "TEST_VARIABLE_1","value": "test1","protected": false,"masked": true}`)
		})

	opt := &CreateGroupVariableOptions{
		Key:       String("TEST_VARIABLE_1"),
		Value:     String("test1"),
		Protected: Bool(false),
		Masked:    Bool(true),
	}

	variable, _, err := client.GroupVariables.CreateVariable(1, opt, nil)
	if err != nil {
		t.Errorf("GroupVariables.CreateVariable returned error: %v", err)
	}

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true}
	if !reflect.DeepEqual(want, variable) {
		t.Errorf("GroupVariables.CreateVariable returned %+v, want %+v", variable, want)
	}
}

func TestDeleteGroupVariable(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "DELETE")
			w.WriteHeader(http.StatusAccepted)
		})

	resp, err := client.GroupVariables.RemoveVariable(1, "TEST_VARIABLE_1")
	if err != nil {
		t.Errorf("GroupVariables.RemoveVariable returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("GroupVariables.RemoveVariable returned %d, want %d", got, want)
	}
}

func TestUpdateGroupVariable(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/variables/TEST_VARIABLE_1",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "PUT")
			fmt.Fprint(w, `{"key": "TEST_VARIABLE_1","value": "test1","protected": false,"masked": true}`)
		})

	variable, _, err := client.GroupVariables.UpdateVariable(1, "TEST_VARIABLE_1", &UpdateGroupVariableOptions{})
	if err != nil {
		t.Errorf("GroupVariables.UpdateVariable returned error: %v", err)
	}

	want := &GroupVariable{Key: "TEST_VARIABLE_1", Value: "test1", Protected: false, Masked: true}
	if !reflect.DeepEqual(want, variable) {
		t.Errorf("Groups.UpdatedGroup returned %+v, want %+v", variable, want)
	}
}
