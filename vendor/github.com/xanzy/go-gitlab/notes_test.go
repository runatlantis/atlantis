package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGetEpicNote(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/groups/1/epics/4329/notes/3", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprint(w, `{"id":3,"type":null,"body":"foo bar","attachment":null,"system":false,"noteable_id":4392,"noteable_type":"Epic","resolvable":false,"noteable_iid":null}`)
	})

	note, _, err := client.Notes.GetEpicNote("1", 4329, 3, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := &Note{
		ID:           3,
		Body:         "foo bar",
		Attachment:   "",
		Title:        "",
		FileName:     "",
		System:       false,
		NoteableID:   4392,
		NoteableType: "Epic",
	}

	if !reflect.DeepEqual(note, want) {
		t.Errorf("Notes.GetEpicNote want %#v, got %#v", note, want)
	}
}
