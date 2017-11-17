package slack

import (
	"net/http"
	"testing"
)

func getBotInfo(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response := []byte(`{"ok": true, "bot": {
			"id":"B02875YLA",
			"deleted":false,
			"name":"github",
			"icons": {
              "image_36":"https:\/\/a.slack-edge.com\/2fac\/plugins\/github\/assets\/service_36.png",
              "image_48":"https:\/\/a.slack-edge.com\/2fac\/plugins\/github\/assets\/service_48.png",
              "image_72":"https:\/\/a.slack-edge.com\/2fac\/plugins\/github\/assets\/service_72.png"
            }
        }}`)
	rw.Write(response)
}

func TestGetBotInfo(t *testing.T) {
	http.HandleFunc("/bots.info", getBotInfo)

	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")

	bot, err := api.GetBotInfo("B02875YLA")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}

	if bot.ID != "B02875YLA" {
		t.Fatal("Incorrect ID")
	}
	if bot.Name != "github" {
		t.Fatal("Incorrect Name")
	}
	if len(bot.Icons.Image36) == 0 {
		t.Fatal("Missing Image36")
	}
	if len(bot.Icons.Image48) == 0 {
		t.Fatal("Missing Image38")
	}
	if len(bot.Icons.Image72) == 0 {
		t.Fatal("Missing Image72")
	}
}
