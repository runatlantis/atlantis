package websocket

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/websocket"
)

func wsHandler(t *testing.T, checkOrigin bool) http.HandlerFunc {
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOriginFunc(checkOrigin),
	}
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Log("upgrade:", err)
			return
		}
		defer c.Close()
	}
}

func TestCheckOriginFunc(t *testing.T) {

	tests := []struct {
		name        string
		checkOrigin bool
		origin      string
		host        string
		wantErr     bool
	}{
		{"same origin", true, "http://example.com/", "example.com", false},
		{"same origin with port", true, "http://example.com:8080/", "example.com:8080", false},
		{"fail with different origin", true, "http://example.net/", "example.com", true},
		{"success with same origin without check", false, "http://example.com/", "example.com", false},
		{"success with different origin without check", false, "http://example.net/", "example.com", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := httptest.NewServer(wsHandler(t, tt.checkOrigin))
			u, _ := url.Parse(s.URL)
			u.Path = "/"
			u.Scheme = "ws"
			header := http.Header{
				"Origin": []string{tt.origin},
				"Host":   []string{tt.host},
			}
			c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
			if err == nil {
				defer c.Close()
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("websocket dial error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}
