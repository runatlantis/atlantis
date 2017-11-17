package slack

import (
	"log"
	"net/http/httptest"
	"sync"
)

const (
	validToken = "testing-token"
)

var (
	serverAddr string
	once       sync.Once
)

func startServer() {
	server := httptest.NewServer(nil)
	serverAddr = server.Listener.Addr().String()
	log.Print("Test WebSocket server listening on ", serverAddr)
}
