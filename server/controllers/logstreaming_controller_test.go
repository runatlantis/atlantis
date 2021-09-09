package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/handlers/mocks"
	"github.com/runatlantis/atlantis/server/handlers/mocks/matchers"
)

func TestGetLogStream_WebSockets(t *testing.T) {
	t.Run("Should Group by Project Info", func(t *testing.T) {
		RegisterMockTestingT(t)
		websocketMock := mocks.NewMockWebsocketHandler()
		projectOutputHandler := mocks.NewMockProjectCommandOutputHandler()
		logger := logging.NewNoopLogger(t)
		webSocketWrapper := mocks.NewMockWebsocketConnectionWrapper()
		params := map[string]string{
			"org":     "test-org",
			"repo":    "test-repo",
			"pull":    "1",
			"project": "test-project",
		}
		request, _ := http.NewRequest(http.MethodGet, "/logStreaming/org/repo/1/project/ws", nil)
		request = mux.SetURLVars(request, params)
		response := httptest.NewRecorder()
		logStreamingController := &controllers.LogStreamingController{
			Logger:                      logger,
			WebsocketHandler:            websocketMock,
			ProjectCommandOutputHandler: projectOutputHandler,
		}

		When(websocketMock.Upgrade(matchers.AnyHttpResponseWriter(), matchers.AnyPtrToHttpRequest(), matchers.AnyHttpHeader())).ThenReturn(webSocketWrapper, nil)

		logStreamingController.GetLogStreamWS(response, request)

		webSocketWrapper.VerifyWasCalled(Once())
	})
}
