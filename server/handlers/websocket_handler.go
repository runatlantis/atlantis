package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_websocket_handler.go WebsocketHandler

type WebsocketHandler interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketConnectionWrapper, error)
	SetReadHandler(w WebsocketConnectionWrapper)
	SetCloseHandler(w WebsocketConnectionWrapper, receiver chan string)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_websocket_connection_wrapper.go WebsocketConnectionWrapper

type WebsocketConnectionWrapper interface {
	ReadMessage() (messageType int, p []byte, err error)
	WriteMessage(messageType int, data []byte) error
	SetCloseHandler(h func(code int, text string) error)
}

type DefaultWebsocketHandler struct {
	handler websocket.Upgrader
	Logger  logging.SimpleLogging
}

func NewWebsocketHandler(logger logging.SimpleLogging) WebsocketHandler {
	h := websocket.Upgrader{}
	h.CheckOrigin = func(r *http.Request) bool { return true }
	return &DefaultWebsocketHandler{
		handler: h,
		Logger:  logger,
	}
}

func (wh *DefaultWebsocketHandler) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketConnectionWrapper, error) {
	return wh.handler.Upgrade(w, r, responseHeader)
}

func (wh *DefaultWebsocketHandler) SetReadHandler(w WebsocketConnectionWrapper) {
	for {
		_, _, err := w.ReadMessage()
		if err != nil {
			// CloseGoingAway (1001) when a browser tab is closed.
			// Expected behaviour since we have a CloseHandler(), log warning if not a CloseGoingAway
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				wh.Logger.Warn("Failed to read WS message: %s", err)
			}
			return
		}
	}
}

func (wh *DefaultWebsocketHandler) SetCloseHandler(w WebsocketConnectionWrapper, receiver chan string) {
	w.SetCloseHandler(func(code int, text string) error {
		// Close the channnel after websocket connection closed.
		// Will gracefully exit the ProjectCommandOutputHandler.Receive() call and cleanup.
		wh.Logger.Info("Close handler called")
		close(receiver)
		return nil
	})
}
