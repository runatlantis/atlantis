package slack

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

type reactionsHandler struct {
	gotParams map[string]string
	response  string
}

func newReactionsHandler() *reactionsHandler {
	return &reactionsHandler{
		gotParams: make(map[string]string),
		response:  `{ "ok": true }`,
	}
}

func (rh *reactionsHandler) accumulateFormValue(k string, r *http.Request) {
	if v := r.FormValue(k); v != "" {
		rh.gotParams[k] = v
	}
}

func (rh *reactionsHandler) handler(w http.ResponseWriter, r *http.Request) {
	rh.accumulateFormValue("channel", r)
	rh.accumulateFormValue("count", r)
	rh.accumulateFormValue("file", r)
	rh.accumulateFormValue("file_comment", r)
	rh.accumulateFormValue("full", r)
	rh.accumulateFormValue("name", r)
	rh.accumulateFormValue("page", r)
	rh.accumulateFormValue("timestamp", r)
	rh.accumulateFormValue("user", r)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(rh.response))
}

func TestSlack_AddReaction(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	tests := []struct {
		name       string
		ref        ItemRef
		wantParams map[string]string
	}{
		{
			"thumbsup",
			NewRefToMessage("ChannelID", "123"),
			map[string]string{
				"name":      "thumbsup",
				"channel":   "ChannelID",
				"timestamp": "123",
			},
		},
		{
			"thumbsup",
			NewRefToFile("FileID"),
			map[string]string{
				"name": "thumbsup",
				"file": "FileID",
			},
		},
		{
			"thumbsup",
			NewRefToComment("FileCommentID"),
			map[string]string{
				"name":         "thumbsup",
				"file_comment": "FileCommentID",
			},
		},
	}
	var rh *reactionsHandler
	http.HandleFunc("/reactions.add", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	for i, test := range tests {
		rh = newReactionsHandler()
		err := api.AddReaction(test.name, test.ref)
		if err != nil {
			t.Fatalf("%d: Unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(rh.gotParams, test.wantParams) {
			t.Errorf("%d: Got params %#v, want %#v", i, rh.gotParams, test.wantParams)
		}
	}
}

func TestSlack_RemoveReaction(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	tests := []struct {
		name       string
		ref        ItemRef
		wantParams map[string]string
	}{
		{
			"thumbsup",
			NewRefToMessage("ChannelID", "123"),
			map[string]string{
				"name":      "thumbsup",
				"channel":   "ChannelID",
				"timestamp": "123",
			},
		},
		{
			"thumbsup",
			NewRefToFile("FileID"),
			map[string]string{
				"name": "thumbsup",
				"file": "FileID",
			},
		},
		{
			"thumbsup",
			NewRefToComment("FileCommentID"),
			map[string]string{
				"name":         "thumbsup",
				"file_comment": "FileCommentID",
			},
		},
	}
	var rh *reactionsHandler
	http.HandleFunc("/reactions.remove", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	for i, test := range tests {
		rh = newReactionsHandler()
		err := api.RemoveReaction(test.name, test.ref)
		if err != nil {
			t.Fatalf("%d: Unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(rh.gotParams, test.wantParams) {
			t.Errorf("%d: Got params %#v, want %#v", i, rh.gotParams, test.wantParams)
		}
	}
}

func TestSlack_GetReactions(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	tests := []struct {
		ref           ItemRef
		params        GetReactionsParameters
		wantParams    map[string]string
		json          string
		wantReactions []ItemReaction
	}{
		{
			NewRefToMessage("ChannelID", "123"),
			GetReactionsParameters{},
			map[string]string{
				"channel":   "ChannelID",
				"timestamp": "123",
			},
			`{"ok": true,
    "type": "message",
    "message": {
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
    }}`,
			[]ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		},
		{
			NewRefToFile("FileID"),
			GetReactionsParameters{Full: true},
			map[string]string{
				"file": "FileID",
				"full": "true",
			},
			`{"ok": true,
    "type": "file",
    "file": {
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
    }}`,
			[]ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		},
		{

			NewRefToComment("FileCommentID"),
			GetReactionsParameters{},
			map[string]string{
				"file_comment": "FileCommentID",
			},
			`{"ok": true,
    "type": "file_comment",
    "file": {},
    "comment": {
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
    }}`,
			[]ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		},
	}
	var rh *reactionsHandler
	http.HandleFunc("/reactions.get", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
	for i, test := range tests {
		rh = newReactionsHandler()
		rh.response = test.json
		got, err := api.GetReactions(test.ref, test.params)
		if err != nil {
			t.Fatalf("%d: Unexpected error: %s", i, err)
		}
		if !reflect.DeepEqual(got, test.wantReactions) {
			t.Errorf("%d: Got reaction %#v, want %#v", i, got, test.wantReactions)
		}
		if !reflect.DeepEqual(rh.gotParams, test.wantParams) {
			t.Errorf("%d: Got params %#v, want %#v", i, rh.gotParams, test.wantParams)
		}
	}
}

func TestSlack_ListReactions(t *testing.T) {
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	rh := newReactionsHandler()
	http.HandleFunc("/reactions.list", func(w http.ResponseWriter, r *http.Request) { rh.handler(w, r) })
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
	want := []ReactedItem{
		ReactedItem{
			Item: NewMessageItem("C1", &Message{Msg: Msg{
				Text: "hello",
				Reactions: []ItemReaction{
					ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
					ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
				},
			}}),
			Reactions: []ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		},
		ReactedItem{
			Item: NewFileItem(&File{Name: "toy"}),
			Reactions: []ItemReaction{
				ItemReaction{Name: "clock1", Count: 3, Users: []string{"U1", "U2"}},
			},
		},
		ReactedItem{
			Item: NewFileCommentItem(&File{Name: "toy"}, &Comment{Comment: "cool toy"}),
			Reactions: []ItemReaction{
				ItemReaction{Name: "astonished", Count: 3, Users: []string{"U1", "U2", "U3"}},
			},
		},
	}
	wantParams := map[string]string{
		"user":  "User",
		"count": "200",
		"page":  "2",
		"full":  "true",
	}
	params := NewListReactionsParameters()
	params.User = "User"
	params.Count = 200
	params.Page = 2
	params.Full = true
	got, paging, err := api.ListReactions(params)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got reaction %#v, want %#v", got, want)
		for i, item := range got {
			fmt.Printf("Item %d, Type: %s\n", i, item.Type)
			fmt.Printf("Message  %#v\n", item.Message)
			fmt.Printf("File     %#v\n", item.File)
			fmt.Printf("Comment  %#v\n", item.Comment)
			fmt.Printf("Reactions %#v\n", item.Reactions)
		}
	}
	if !reflect.DeepEqual(rh.gotParams, wantParams) {
		t.Errorf("Got params %#v, want %#v", rh.gotParams, wantParams)
	}
	if reflect.DeepEqual(paging, Paging{}) {
		t.Errorf("Want paging data, got empty struct")
	}
}
