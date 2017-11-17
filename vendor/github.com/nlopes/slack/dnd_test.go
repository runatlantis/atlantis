package slack

import (
	"net/http"
	"reflect"
	"testing"
)

func TestSlack_EndDND(t *testing.T) {
	http.HandleFunc("/dnd.endDnd", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{ "ok": true }`))
	})
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	err := api.EndDND()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestSlack_EndSnooze(t *testing.T) {
	http.HandleFunc("/dnd.endSnooze", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{ "ok": true,
                          "dnd_enabled": true,
                          "next_dnd_start_ts": 1450418400,
                          "next_dnd_end_ts": 1450454400,
                          "snooze_enabled": false }`))
	})
	state := DNDStatus{
		Enabled:            true,
		NextStartTimestamp: 1450418400,
		NextEndTimestamp:   1450454400,
		SnoozeInfo:         SnoozeInfo{SnoozeEnabled: false},
	}
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	snoozeState, err := api.EndSnooze()
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	eq := reflect.DeepEqual(snoozeState, &state)
	if !eq {
		t.Errorf("got %v; want %v", snoozeState, &state)
	}
}

func TestSlack_GetDNDInfo(t *testing.T) {
	http.HandleFunc("/dnd.info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
            "ok": true,
            "dnd_enabled": true,
            "next_dnd_start_ts": 1450416600,
            "next_dnd_end_ts": 1450452600,
            "snooze_enabled": true,
            "snooze_endtime": 1450416600,
            "snooze_remaining": 1196
        }`))
	})
	userDNDInfo := DNDStatus{
		Enabled:            true,
		NextStartTimestamp: 1450416600,
		NextEndTimestamp:   1450452600,
		SnoozeInfo: SnoozeInfo{
			SnoozeEnabled:   true,
			SnoozeEndTime:   1450416600,
			SnoozeRemaining: 1196,
		},
	}
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	userDNDInfoResponse, err := api.GetDNDInfo(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	eq := reflect.DeepEqual(userDNDInfoResponse, &userDNDInfo)
	if !eq {
		t.Errorf("got %v; want %v", userDNDInfoResponse, &userDNDInfo)
	}
}

func TestSlack_GetDNDTeamInfo(t *testing.T) {
	http.HandleFunc("/dnd.teamInfo", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
            "ok": true,
            "users": {
                "U023BECGF": {
                    "dnd_enabled": true,
                    "next_dnd_start_ts": 1450387800,
                    "next_dnd_end_ts": 1450423800
                },
                "U058CJVAA": {
                    "dnd_enabled": false,
                    "next_dnd_start_ts": 1,
                    "next_dnd_end_ts": 1
                }
            }
        }`))
	})
	usersDNDInfo := map[string]DNDStatus{
		"U023BECGF": DNDStatus{
			Enabled:            true,
			NextStartTimestamp: 1450387800,
			NextEndTimestamp:   1450423800,
		},
		"U058CJVAA": DNDStatus{
			Enabled:            false,
			NextStartTimestamp: 1,
			NextEndTimestamp:   1,
		},
	}
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	usersDNDInfoResponse, err := api.GetDNDTeamInfo(nil)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	eq := reflect.DeepEqual(usersDNDInfoResponse, usersDNDInfo)
	if !eq {
		t.Errorf("got %v; want %v", usersDNDInfoResponse, usersDNDInfo)
	}
}

func TestSlack_SetSnooze(t *testing.T) {
	http.HandleFunc("/dnd.setSnooze", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
            "ok": true,
            "dnd_enabled": true,
            "snooze_endtime": 1450373897,
            "snooze_remaining": 60
        }`))
	})
	snooze := DNDStatus{
		Enabled: true,
		SnoozeInfo: SnoozeInfo{
			SnoozeEndTime:   1450373897,
			SnoozeRemaining: 60,
		},
	}
	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	snoozeResponse, err := api.SetSnooze(60)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	eq := reflect.DeepEqual(snoozeResponse, &snooze)
	if !eq {
		t.Errorf("got %v; want %v", snoozeResponse, &snooze)
	}
}
