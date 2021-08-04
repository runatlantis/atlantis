package controllers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/handlers"
	"github.com/runatlantis/atlantis/server/logging"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/handlers/mocks"
	"github.com/runatlantis/atlantis/server/handlers/mocks/matchers"
)

func TestGetLogStream_WebSockets(t *testing.T) {
	t.Run("Should Group by Project Info", func(t *testing.T) {
		RegisterMockTestingT(t)
		tempchan := make(chan *models.ProjectCmdOutputLine)
		websocketMock := mocks.NewMockWebsocketHandler()
		projectOutputHandler := &handlers.DefaultProjectCommandOutputHandler{}
		logger := logging.NewNoopLogger(t)
		websocketWriterMock := mocks.NewMockWebsocketResponseWriter()
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
			Logger:              logger,
			WebsocketHandler:    websocketMock,
			ProjectCommandOutputHandler: projectOutputHandler,
		}

		When(websocketMock.Upgrade(matchers.AnyHttpResponseWriter(), matchers.AnyPtrToHttpRequest(), matchers.AnyHttpHeader())).ThenReturn(websocketWriterMock, nil)

		go func() {
			tempchan <- &models.ProjectCmdOutputLine{
				ProjectInfo: "test-org/test-repo/1/test-project",
				Line:        "Test Terraform Output",
			}
		}()

		go func() {
			projectOutputHandler.Handle()
		}()

		go func() {
			logStreamingController.GetLogStreamWS(response, request)
		}()

		time.Sleep(1 * time.Second)

		websocketWriterMock.VerifyWasCalled(Once())
	})
}
