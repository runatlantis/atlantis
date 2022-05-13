package websocket

import (
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/logging"
)

func NewWriter(log logging.Logger) *Writer {
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
	log logging.Logger
}

func (w *Writer) Write(rw http.ResponseWriter, r *http.Request, input chan string) error {
	conn, err := w.upgrader.Upgrade(rw, r, nil)

	if err != nil {
		return errors.Wrap(err, "upgrading websocket connection")
	}

	// block on reading our input channel
	for msg := range input {
		if err := conn.WriteMessage(websocket.BinaryMessage, []byte("\r"+msg+"\n")); err != nil {
			w.log.Warn(fmt.Sprintf("Failed to write ws message: %s", err))
			return err
		}
	}

	// close ws conn after input channel is closed
	if err = conn.Close(); err != nil {
		w.log.Warn(fmt.Sprintf("Failed to close ws connection: %s", err))
	}
	return nil
}
