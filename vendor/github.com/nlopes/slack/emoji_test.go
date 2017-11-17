package slack

import (
	"net/http"
	"reflect"
	"testing"
)

func getEmojiHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response := []byte(`{"ok": true, "emoji": {
			"bowtie": "https://my.slack.com/emoji/bowtie/46ec6f2bb0.png",
			"squirrel": "https://my.slack.com/emoji/squirrel/f35f40c0e0.png",
			"shipit": "alias:squirrel"
		}}`)
	rw.Write(response)
}

func TestGetEmoji(t *testing.T) {
	http.HandleFunc("/emoji.list", getEmojiHandler)

	once.Do(startServer)
	SLACK_API = "http://" + serverAddr + "/"
	api := New("testing-token")
	emojisResponse := map[string]string{
		"bowtie":   "https://my.slack.com/emoji/bowtie/46ec6f2bb0.png",
		"squirrel": "https://my.slack.com/emoji/squirrel/f35f40c0e0.png",
		"shipit":   "alias:squirrel",
	}

	emojis, err := api.GetEmoji()
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
		return
	}
	eq := reflect.DeepEqual(emojis, emojisResponse)
	if !eq {
		t.Errorf("got %v; want %v", emojis, emojisResponse)
	}
}
