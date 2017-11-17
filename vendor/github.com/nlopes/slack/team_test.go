package slack

import (
	"errors"
	"net/http"
	"testing"
	"strings"
)

var (
	ErrIncorrectResponse = errors.New("Response is incorrect")
)

func getTeamInfo(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response := []byte(`{"ok": true, "team": {
			"id": "F0UWHUX",
			"name": "notalar",
			"domain": "notalar",
			"icon": {
              "image_34": "https://slack.global.ssl.fastly.net/66f9/img/avatars-teams/ava_0002-34.png",
              "image_44": "https://slack.global.ssl.fastly.net/66f9/img/avatars-teams/ava_0002-44.png",
              "image_55": "https://slack.global.ssl.fastly.net/66f9/img/avatars-teams/ava_0002-55.png",
              "image_default": true
          }
		}}`)
	rw.Write(response)
}

func TestGetTeamInfo(t *testing.T) {
	http.HandleFunc("/team.info", getTeamInfo)

	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")

	teamInfo, err := api.GetTeamInfo()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	// t.Fatal refers to -> t.Errorf & return
	if teamInfo.ID != "F0UWHUX" {
		t.Fatal(ErrIncorrectResponse)
	}
	if teamInfo.Domain != "notalar" {
		t.Fatal(ErrIncorrectResponse)
	}
	if teamInfo.Name != "notalar" {
		t.Fatal(ErrIncorrectResponse)
	}
	if teamInfo.Icon == nil {
		t.Fatal(ErrIncorrectResponse)
	}
}

func getTeamAccessLogs(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response := []byte(`{"ok": true, "logins": [{
			"user_id": "F0UWHUX",
			"username": "notalar",
			"date_first": 1475684477,
			"date_last": 1475684645,
			"count": 8,
			"ip": "127.0.0.1",
			"user_agent": "SlackWeb/3abb0ae2380d48a9ae20c58cc624ebcd Mozilla/5.0 (Windows NT 6.1; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Slack/1.2.6 Chrome/45.0.2454.85 AtomShell/0.34.3 Safari/537.36 Slack_SSB/1.2.6",
			"isp": "AT&T U-verse",
                        "country": "US",
                        "region": "IN"
                        },
                        {
                        "user_id": "XUHWU0F",
			"username": "ralaton",
			"date_first": 1447395893,
			"date_last": 1447395965,
			"count": 5,
			"ip": "192.168.0.1",
			"user_agent": "com.tinyspeck.chatlyio/2.60 (iPhone; iOS 9.1; Scale/3.00)",
			"isp": null,
                        "country": null,
                        "region": null
                        }],
                        "paging": {
    			"count": 2,
    			"total": 2,
    			"page": 1,
    			"pages": 1
    			}
  }`)
	rw.Write(response)
}

func TestGetAccessLogs(t *testing.T) {
	http.HandleFunc("/team.accessLogs", getTeamAccessLogs)

	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")

	logins, paging, err := api.GetAccessLogs(NewAccessLogParameters())
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if len(logins) != 2 {
		t.Fatal("Should have been 2 logins")
	}

	// test the first login
	login1 := logins[0]
	login2 := logins[1]

	if (login1.UserID != "F0UWHUX") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.Username != "notalar") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.DateFirst != 1475684477) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.DateLast != 1475684645) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.Count != 8) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.IP != "127.0.0.1") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (!strings.HasPrefix(login1.UserAgent, "SlackWeb")) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.ISP != "AT&T U-verse") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.Country != "US") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login1.Region != "IN") {
		t.Fatal(ErrIncorrectResponse)
	}

	// test that the null values from login2 are coming across correctly
	if (login2.ISP != "") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login2.Country != "") {
		t.Fatal(ErrIncorrectResponse)
	}
	if (login2.Region != "") {
		t.Fatal(ErrIncorrectResponse)
	}

	// test the paging
	if (paging.Count != 2) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (paging.Total != 2) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (paging.Page != 1) {
		t.Fatal(ErrIncorrectResponse)
	}
	if (paging.Pages != 1) {
		t.Fatal(ErrIncorrectResponse)
	}
}

