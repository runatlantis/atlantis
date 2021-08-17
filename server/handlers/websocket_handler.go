package handlers

import (
	"net/http"

	"github.com/gorilla/websocket"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_websocket_handler.go WebsocketHandler

type WebsocketHandler interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketResponseWriter, error)
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_websocket_response_writer.go WebsocketResponseWriter

type WebsocketResponseWriter interface {
	WriteMessage(messageType int, data []byte) error
	Close() error
}

type DefaultWebsocketHandler struct {
	handler websocket.Upgrader
}

func NewWebsocketHandler() WebsocketHandler {
	return &DefaultWebsocketHandler{
		handler: websocket.Upgrader{},
	}
}

func (wh *DefaultWebsocketHandler) Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (WebsocketResponseWriter, error) {
	return wh.handler.Upgrade(w, r, responseHeader)
}
