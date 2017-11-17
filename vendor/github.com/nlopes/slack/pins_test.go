package slack

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type pinsHandler struct {
	gotParams map[string]string
	response  string
}

func newPinsHandler() *pinsHandler {
	return &pinsHandler{
		gotParams: make(map[string]string),
		response:  `{ "ok": true }`,
	}
}

func (rh *pinsHandler) accumulateFormValue(k string, r *http.Request) {
	if v := r.FormValue(k); v != "" {
		rh.gotParams[k] = v
	}
}

func (rh *pinsHandler) handler(w http.ResponseWriter, r *http.Request) {
	rh.accumulateFormValue("channel", r)
	rh.accumulateFormValue("count", r)
	rh.accumulateFormValue("file", r)
	rh.accumulateFormValue("file_comment", r)
	rh.accumulateFormValue("page", r)
	rh.accumulateFormValue("timestamp", r)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(rh.response))
}

func TestSlack_AddPin(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	tests := []struct {
		channel    string
		ref        ItemRef
		wantParams map[string]string
	}{
		{
			"ChannelID",
			NewRefToMessage("ChannelID", "123"),
			map[string]string{
				"channel":   "ChannelID",
				"timestamp": "123",
			},
		},
		{
			"ChannelID",
			NewRefToFile("FileID"),
			map[string]string{
				"channel": "ChannelID",
				"file":    "FileID",
			},
		},
		{
			"ChannelID",
			NewRefToComment("FileCommentID"),
			map[string]string{
				"channel":      "ChannelID",
				"file_comment": "FileCommentID",
			},
		},
	}
	var rh *pinsHandler
	http.HandleFunc("/pins.add", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	for i, test := range tests {
		rh = newPinsHandler()
		err := api.AddPin(test.channel, test.ref)
		if err != nil {
			t.Fatalf("%d: Unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(rh.gotParams, test.wantParams) {
			t.Errorf("%d: Got params %#v, want %#v", i, rh.gotParams, test.wantParams)
		}
	}
}

func TestSlack_RemovePin(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	tests := []struct {
		channel    string
		ref        ItemRef
		wantParams map[string]string
	}{
		{
			"ChannelID",
			NewRefToMessage("ChannelID", "123"),
			map[string]string{
				"channel":   "ChannelID",
				"timestamp": "123",
			},
		},
		{
			"ChannelID",
			NewRefToFile("FileID"),
			map[string]string{
				"channel": "ChannelID",
				"file":    "FileID",
			},
		},
		{
			"ChannelID",
			NewRefToComment("FileCommentID"),
			map[string]string{
				"channel":      "ChannelID",
				"file_comment": "FileCommentID",
			},
		},
	}
	var rh *pinsHandler
	http.HandleFunc("/pins.remove", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	for i, test := range tests {
		rh = newPinsHandler()
		err := api.RemovePin(test.channel, test.ref)
		if err != nil {
			t.Fatalf("%d: Unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(rh.gotParams, test.wantParams) {
			t.Errorf("%d: Got params %#v, want %#v", i, rh.gotParams, test.wantParams)
		}
	}
}

func TestSlack_ListPins(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	rh := newPinsHandler()
	http.HandleFunc("/pins.list", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	rh.response = `{"ok": true,
    "items": [
        {
            "type": "message",
            "channel": "C1",
            "message": {
                "text": "hello",
                "reactions": [
                    {
                        "name": "astonished",
                        "count": 3,
                        "users": [ "U1", "U2", "U3" ]
                    },
                    {
                        "name": "clock1",
                        "count": 3,
                        "users": [ "U1", "U2" ]
                    }
                ]
            }
        },
        {
            "type": "file",
            "file": {
                "name": "toy",
                "reactions": [
                    {
                        "name": "clock1",
                        "count": 3,
                        "users": [ "U1", "U2" ]
                    }
                ]
            }
        },
        {
            "type": "file_comment",
            "file": {
                "name": "toy"
            },
            "comment": {
                "comment": "cool toy",
                "reactions": [
                    {
                        "name": "astonished",
                        "count": 3,
                        "users": [ "U1", "U2", "U3" ]
                    }
                ]
            }
        }
    ],
    "paging": {
        "count": 100,
        "total": 4,
        "page": 1,
        "pages": 1
    }}`
	want := []Item{
		NewMessageItem("C1", &Message{Msg: Msg{
			Text: "hello",
			Reactions: []ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		}}),
		NewFileItem(&File{Name: "toy"}),
		NewFileCommentItem(&File{Name: "toy"}, &Comment{Comment: "cool toy"}),
	}
	wantParams := map[string]string{
		"channel": "ChannelID",
	}
	got, paging, err := api.ListPins("ChannelID")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got Pins %#v, want %#v", got, want)
		for i, item := range got {
			fmt.Printf("Item %d, Type: %s\n", i, item.Type)
			fmt.Printf("Message  %#v\n", item.Message)
			fmt.Printf("File     %#v\n", item.File)
			fmt.Printf("Comment  %#v\n", item.Comment)
		}
	}
	if !reflect.DeepEqual(rh.gotParams, wantParams) {
		t.Errorf("Got params %#v, want %#v", rh.gotParams, wantParams)
	}
	if reflect.DeepEqual(paging, Paging{}) {
		t.Errorf("Want paging data, got empty struct")
	}
}
