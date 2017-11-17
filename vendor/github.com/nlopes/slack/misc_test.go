package slack

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"testing"
)

var (
	parseResponseOnce sync.Once
)

func parseResponseHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	token := r.FormValue("token")
	log.Println(token)
	if token == "" {
		rw.Write([]byte(`{"ok":false,"error":"not_authed"}`))
		return
	}
	if token != validToken {
		rw.Write([]byte(`{"ok":false,"error":"invalid_auth"}`))
		return
	}
	response := []byte(`{"ok": true}`)
	rw.Write(response)
}

func setParseResponseHandler() {
	http.HandleFunc("/parseResponse", parseResponseHandler)
}

func TestParseResponse(t *testing.T) {
	parseResponseOnce.Do(setParseResponseHandler)
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	values := url.Values{
		"token": {validToken},
	}
	responsePartial := &SlackResponse{}
	err := post("parseResponse", values, responsePartial, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestParseResponseNoToken(t *testing.T) {
	parseResponseOnce.Do(setParseResponseHandler)
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	values := url.Values{}
	responsePartial := &SlackResponse{}
	err := post("parseResponse", values, responsePartial, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if responsePartial.Ok == true {
		t.Errorf("Unexpected error: %s", err)
	} else if responsePartial.Error != "not_authed" {
		t.Errorf("got %v; want %v", responsePartial.Error, "not_authed")
	}
}

func TestParseResponseInvalidToken(t *testing.T) {
	parseResponseOnce.Do(setParseResponseHandler)
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	values := url.Values{
		"token": {"whatever"},
	}
	responsePartial := &SlackResponse{}
	err := post("parseResponse", values, responsePartial, false)
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	if responsePartial.Ok == true {
		t.Errorf("Unexpected error: %s", err)
	} else if responsePartial.Error != "invalid_auth" {
		t.Errorf("got %v; want %v", responsePartial.Error, "invalid_auth")
	}
}
