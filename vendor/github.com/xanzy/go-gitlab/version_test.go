package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetVersion(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/version",
		func(w http.ResponseWriter, r *http.Request) {
			testMethod(t, r, "GET")
			fmt.Fprint(w, `{"version":"11.3.4-ee", "revision":"14d3a1d"}`)
		})

	version, _, err := client.Version.GetVersion()
	if err != nil {
		t.Errorf("Version.GetVersion returned error: %v", err)
	}

	want := &Version{Version: "11.3.4-ee", Revision: "14d3a1d"}
	if !reflect.DeepEqual(want, version) {
		t.Errorf("Version.GetVersion returned %+v, want %+v", version, want)
	}
}
