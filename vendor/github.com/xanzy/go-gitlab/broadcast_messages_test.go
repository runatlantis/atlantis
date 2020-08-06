package gitlab

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"
)

func TestListBroadcastMessages(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/broadcast_messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `[{
			"message": "Some Message",
			"starts_at": "2017-06-26T06:00:00.000Z",
			"ends_at": "2017-06-27T12:59:00.000Z",
			"color": "#E75E40",
			"font": "#FFFFFF",
			"id": 1,
			"active": false
		},{
			"message": "SomeMessage2",
			"starts_at": "2015-04-27T06:43:00.000Z",
			"ends_at": "2015-04-28T20:43:00.000Z",
			"color": "#AA33EE",
			"font": "#224466",
			"id": 2,
			"active": true
		}]`)
	})

	got, _, err := client.BroadcastMessage.ListBroadcastMessages(nil, nil)
	if err != nil {
		t.Errorf("ListBroadcastMessages returned error: %v", err)
	}

	wantedFirstStartsAt := time.Date(2017, 06, 26, 6, 0, 0, 0, time.UTC)
	wantedFirstEndsAt := time.Date(2017, 06, 27, 12, 59, 0, 0, time.UTC)

	wantedSecondStartsAt := time.Date(2015, 04, 27, 6, 43, 0, 0, time.UTC)
	wantedSecondEndsAt := time.Date(2015, 04, 28, 20, 43, 0, 0, time.UTC)

	want := []*BroadcastMessage{{
		Message:  "Some Message",
		StartsAt: &wantedFirstStartsAt,
		EndsAt:   &wantedFirstEndsAt,
		Color:    "#E75E40",
		Font:     "#FFFFFF",
		ID:       1,
		Active:   false,
	}, {
		Message:  "SomeMessage2",
		StartsAt: &wantedSecondStartsAt,
		EndsAt:   &wantedSecondEndsAt,
		Color:    "#AA33EE",
		Font:     "#224466",
		ID:       2,
		Active:   true,
	}}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListBroadcastMessages returned \ngot:\n%v\nwant:\n%v", Stringify(got), Stringify(want))
	}
}

func TestGetBroadcastMessages(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/broadcast_messages/1/", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		fmt.Fprintf(w, `{
			"message": "Some Message",
			"starts_at": "2017-06-26T06:00:00.000Z",
			"ends_at": "2017-06-27T12:59:00.000Z",
			"color": "#E75E40",
			"font": "#FFFFFF",
			"id": 1,
			"active": false
		}`)
	})

	got, _, err := client.BroadcastMessage.GetBroadcastMessage(1)
	if err != nil {
		t.Errorf("GetBroadcastMessage returned error: %v", err)
	}

	wantedStartsAt := time.Date(2017, time.June, 26, 6, 0, 0, 0, time.UTC)
	wantedEndsAt := time.Date(2017, time.June, 27, 12, 59, 0, 0, time.UTC)

	want := &BroadcastMessage{
		Message:  "Some Message",
		StartsAt: &wantedStartsAt,
		EndsAt:   &wantedEndsAt,
		Color:    "#E75E40",
		Font:     "#FFFFFF",
		ID:       1,
		Active:   false,
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetBroadcastMessage returned \ngot:\n%v\nwant:\n%v", Stringify(got), Stringify(want))
	}
}

func TestCreateBroadcastMessages(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	wantedStartsAt := time.Date(2017, time.June, 26, 6, 0, 0, 0, time.UTC)
	wantedEndsAt := time.Date(2017, time.June, 27, 12, 59, 0, 0, time.UTC)

	mux.HandleFunc("/api/v4/broadcast_messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "POST")
		fmt.Fprintf(w, `{
			"message": "Some Message",
			"starts_at": "2017-06-26T06:00:00.000Z",
			"ends_at": "2017-06-27T12:59:00.000Z",
			"color": "#E75E40",
			"font": "#FFFFFF",
			"id": 42,
			"active": false
		}`)
	})

	opt := &CreateBroadcastMessageOptions{
		Message:  String("Some Message"),
		StartsAt: &wantedStartsAt,
		EndsAt:   &wantedEndsAt,
		Color:    String("#E75E40"),
		Font:     String("#FFFFFF"),
	}

	got, _, err := client.BroadcastMessage.CreateBroadcastMessage(opt)
	if err != nil {
		t.Errorf("CreateBroadcastMessage returned error: %v", err)
	}

	want := &BroadcastMessage{
		Message:  "Some Message",
		StartsAt: &wantedStartsAt,
		EndsAt:   &wantedEndsAt,
		Color:    "#E75E40",
		Font:     "#FFFFFF",
		ID:       42,
		Active:   false,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("CreateBroadcastMessage returned \ngot:\n%v\nwant:\n%v", Stringify(got), Stringify(want))
	}
}

func TestUpdateBroadcastMessages(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	wantedStartsAt := time.Date(2017, time.June, 26, 6, 0, 0, 0, time.UTC)
	wantedEndsAt := time.Date(2017, time.June, 27, 12, 59, 0, 0, time.UTC)

	mux.HandleFunc("/api/v4/broadcast_messages/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "PUT")
		fmt.Fprintf(w, `{
			"message": "Some Message Updated",
			"starts_at": "2017-06-26T06:00:00.000Z",
			"ends_at": "2017-06-27T12:59:00.000Z",
			"color": "#E75E40",
			"font": "#FFFFFF",
			"id": 42,
			"active": false
		}`)
	})

	opt := &UpdateBroadcastMessageOptions{
		Message:  String("Some Message Updated"),
		StartsAt: &wantedStartsAt,
		EndsAt:   &wantedEndsAt,
		Color:    String("#E75E40"),
		Font:     String("#FFFFFF"),
	}

	got, _, err := client.BroadcastMessage.UpdateBroadcastMessage(1, opt)
	if err != nil {
		t.Errorf("UpdateBroadcastMessage returned error: %v", err)
	}

	want := &BroadcastMessage{
		Message:  "Some Message Updated",
		StartsAt: &wantedStartsAt,
		EndsAt:   &wantedEndsAt,
		Color:    "#E75E40",
		Font:     "#FFFFFF",
		ID:       42,
		Active:   false,
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("UpdateBroadcastMessage returned \ngot:\n%v\nwant:\n%v", Stringify(got), Stringify(want))
	}
}

func TestDeleteBroadcastMessages(t *testing.T) {
	mux, server, client := setup()
	defer teardown(server)

	mux.HandleFunc("/api/v4/broadcast_messages/1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
	})

	_, err := client.BroadcastMessage.DeleteBroadcastMessage(1)
	if err != nil {
		t.Errorf("UpdateBroadcastMessage returned error: %v", err)
	}
}
