package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestListCustomUserAttributes(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/users/2/custom_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"key":"testkey1", "value":"testvalue1"}, {"key":"testkey2", "value":"testvalue2"}]`)
	})

	customAttributes, _, err := client.CustomAttribute.ListCustomUserAttributes(2)

	if err != nil {
		t.Errorf("CustomAttribute.ListCustomUserAttributes returned error: %v", err)
	}

	want := []*CustomAttribute{{Key: "testkey1", Value: "testvalue1"}, {Key: "testkey2", Value: "testvalue2"}}
	if !reflect.DeepEqual(want, customAttributes) {
		t.Errorf("CustomAttribute.ListCustomUserAttributes returned %+v, want %+v", customAttributes, want)
	}
}

func TestListCustomGroupAttributes(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/2/custom_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"key":"testkey1", "value":"testvalue1"}, {"key":"testkey2", "value":"testvalue2"}]`)
	})

	customAttributes, _, err := client.CustomAttribute.ListCustomGroupAttributes(2)

	if err != nil {
		t.Errorf("CustomAttribute.ListCustomGroupAttributes returned error: %v", err)
	}

	want := []*CustomAttribute{{Key: "testkey1", Value: "testvalue1"}, {Key: "testkey2", Value: "testvalue2"}}
	if !reflect.DeepEqual(want, customAttributes) {
		t.Errorf("CustomAttribute.ListCustomGroupAttributes returned %+v, want %+v", customAttributes, want)
	}
}

func TestListCustomProjectAttributes(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/2/custom_attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `[{"key":"testkey1", "value":"testvalue1"}, {"key":"testkey2", "value":"testvalue2"}]`)
	})

	customAttributes, _, err := client.CustomAttribute.ListCustomProjectAttributes(2)

	if err != nil {
		t.Errorf("CustomAttribute.ListCustomProjectAttributes returned error: %v", err)
	}

	want := []*CustomAttribute{{Key: "testkey1", Value: "testvalue1"}, {Key: "testkey2", Value: "testvalue2"}}
	if !reflect.DeepEqual(want, customAttributes) {
		t.Errorf("CustomAttribute.ListCustomProjectAttributes returned %+v, want %+v", customAttributes, want)
	}
}

func TestGetCustomUserAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/users/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"key":"testkey1", "value":"testvalue1"}`)
	})

	customAttribute, _, err := client.CustomAttribute.GetCustomUserAttribute(2, "testkey1")

	if err != nil {
		t.Errorf("CustomAttribute.GetCustomUserAttribute returned error: %v", err)
	}

	want := &CustomAttribute{Key: "testkey1", Value: "testvalue1"}
	if !reflect.DeepEqual(want, customAttribute) {
		t.Errorf("CustomAttribute.GetCustomUserAttribute returned %+v, want %+v", customAttribute, want)
	}
}

func TestGetCustomGropupAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"key":"testkey1", "value":"testvalue1"}`)
	})

	customAttribute, _, err := client.CustomAttribute.GetCustomGroupAttribute(2, "testkey1")

	if err != nil {
		t.Errorf("CustomAttribute.GetCustomGroupAttribute returned error: %v", err)
	}

	want := &CustomAttribute{Key: "testkey1", Value: "testvalue1"}
	if !reflect.DeepEqual(want, customAttribute) {
		t.Errorf("CustomAttribute.GetCustomGroupAttribute returned %+v, want %+v", customAttribute, want)
	}
}

func TestGetCustomProjectAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"key":"testkey1", "value":"testvalue1"}`)
	})

	customAttribute, _, err := client.CustomAttribute.GetCustomProjectAttribute(2, "testkey1")

	if err != nil {
		t.Errorf("CustomAttribute.GetCustomProjectAttribute returned error: %v", err)
	}

	want := &CustomAttribute{Key: "testkey1", Value: "testvalue1"}
	if !reflect.DeepEqual(want, customAttribute) {
		t.Errorf("CustomAttribute.GetCustomProjectAttribute returned %+v, want %+v", customAttribute, want)
	}
}

func TestSetCustomUserAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/users/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{"key":"testkey1", "value":"testvalue1"}`)
	})

	customAttribute, _, err := client.CustomAttribute.SetCustomUserAttribute(2, CustomAttribute{
		Key:   "testkey1",
		Value: "testvalue1",
	})

	if err != nil {
		t.Errorf("CustomAttribute.SetCustomUserAttributes returned error: %v", err)
	}

	want := &CustomAttribute{Key: "testkey1", Value: "testvalue1"}
	if !reflect.DeepEqual(want, customAttribute) {
		t.Errorf("CustomAttribute.SetCustomUserAttributes returned %+v, want %+v", customAttribute, want)
	}
}

func TestSetCustomGroupAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprint(w, `{"key":"testkey1", "value":"testvalue1"}`)
	})

	customAttribute, _, err := client.CustomAttribute.SetCustomGroupAttribute(2, CustomAttribute{
		Key:   "testkey1",
		Value: "testvalue1",
	})

	if err != nil {
		t.Errorf("CustomAttribute.SetCustomGroupAttributes returned error: %v", err)
	}

	want := &CustomAttribute{Key: "testkey1", Value: "testvalue1"}
	if !reflect.DeepEqual(want, customAttribute) {
		t.Errorf("CustomAttribute.SetCustomGroupAttributes returned %+v, want %+v", customAttribute, want)
	}
}

func TestDeleteCustomUserAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/users/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.CustomAttribute.DeleteCustomUserAttribute(2, "testkey1")
	if err != nil {
		t.Errorf("CustomAttribute.DeleteCustomUserAttribute returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("CustomAttribute.DeleteCustomUserAttribute returned %d, want %d", got, want)
	}
}

func TestDeleteCustomGroupAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.CustomAttribute.DeleteCustomGroupAttribute(2, "testkey1")
	if err != nil {
		t.Errorf("CustomAttribute.DeleteCustomGroupAttribute returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("CustomAttribute.DeleteCustomGroupAttribute returned %d, want %d", got, want)
	}
}

func TestDeleteCustomProjectAttribute(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/projects/2/custom_attributes/testkey1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusAccepted)
	})

	resp, err := client.CustomAttribute.DeleteCustomProjectAttribute(2, "testkey1")
	if err != nil {
		t.Errorf("CustomAttribute.DeleteCustomProjectAttribute returned error: %v", err)
	}

	want := http.StatusAccepted
	got := resp.StatusCode
	if got != want {
		t.Errorf("CustomAttribute.DeleteCustomProjectAttribute returned %d, want %d", got, want)
	}
}
