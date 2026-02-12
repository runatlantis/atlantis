package websocket

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

func NewWriter(log logging.SimpleLogging, checkOrigin bool) *Writer {
	upgrader := websocket.Upgrader{
		CheckOrigin: checkOriginFunc(checkOrigin),
	}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &Writer{
		upgrader: upgrader,
		log:      log,
	}
}

type Writer struct {
	upgrader websocket.Upgrader
	log      logging.SimpleLogging
}

func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, input chan string) error {
	conn, err := w.upgrader.Upgrade(rw, r, nil)
	if err != nil {
		return errors.Wrap(err, "upgrading websocket connection")
	}

	var connMu sync.Mutex
	done := make(chan struct{})

	// read from the socket to detect a disconnection
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// block on reading our input channel
	for {
		select {
		case msg, ok := <-input:
			if !ok {
				// input channel is closed, close the connection
				if err := conn.Close(); err != nil {
					w.log.Warn("Failed to close ws connection: %s", err)
				}
				return nil
			}

			connMu.Lock()
			err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n"))
			connMu.Unlock()
			if err != nil {
				w.log.Warn("Failed to write ws message: %s", err)
				return err
			}
		case <-done:
			// client disconnected
			return nil
		}
	}
}
