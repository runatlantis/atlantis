package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

func NewWriter(log logging.SimpleLogging) *Writer {
	upgrader := websocket.Upgrader{}
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	return &Writer{
		upgrader: upgrader,
		log:      log,
	}
}

type Writer struct {
	upgrader websocket.Upgrader

	//TODO: Remove dependency on atlantis logger here if we upstream this.
	log logging.SimpleLogging
}

func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, input chan string) error {
	conn, err := w.upgrader.Upgrade(rw, r, nil)

	if err != nil {
		return errors.Wrap(err, "upgrading websocket connection")
	}

	conn.SetCloseHandler(func(code int, text string) error {
		// Close the channnel after websocket connection closed.
		// Will gracefully exit the ProjectCommandOutputHandler.Register() call and cleanup.
		// is it good practice to close at the receiver? Probably not, we should figure out a better
		// way to handle this case
		close(input)
		return nil
	})

	// Add a reader goroutine to listen for socket.close() events.
	go w.setReadHandler(conn)

	// block on reading our input channel
	for msg := range input {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
			w.log.Warn("Failed to write ws message: %s", err)
			return err
		}
	}

	return nil
}

func (w *Writer) setReadHandler(c *websocket.Conn) {
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			// CloseGoingAway (1001) when a browser tab is closed.
			// Expected behaviour since we have a CloseHandler(), log warning if not a CloseGoingAway
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				w.log.Warn("Failed to read WS message: %s", err)
			}
			return
		}
	}

}
