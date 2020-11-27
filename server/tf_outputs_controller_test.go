package server

import (
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)

func TestTfOutputsController_GetTfOutput(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	tfOutputHelper, err := terraform.NewOutputHelper(tmp)
	Ok(t, err)

	log := logging.NewSimpleLogger("test", false, logging.Debug)

	controller := TfOutputsController{
		log:            log,
		tfOutputHelper: tfOutputHelper,
		wcUpgrader:     websocket.Upgrader{},
	}

	// Creates a test tf output file
	tfOutputFileName := "20201121175848-runatalntis_atlantis-1-test-1a2b3c4-default-init"
	file, err := os.OpenFile(filepath.Join(tmp, tfOutputFileName), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	Ok(t, err)
	defer file.Close()

	// Create a test server for the websocket
	s := httptest.NewServer(http.HandlerFunc(controller.GetTfOutput))
	defer s.Close()

	// Creates the URL for the websocket
	url, err := url.Parse(s.URL)
	Ok(t, err)
	url.Scheme = "ws"

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	Ok(t, err)
	defer ws.Close()

	// Writes the first message with the tf output file to "tail"
	err = ws.WriteMessage(websocket.TextMessage, []byte("20201121175848|runatalntis_atlantis|1|1a2b3c4|test|default|init"))
	Ok(t, err)

	testData := []string{"ab", "cd", "ef"}
	for _, data := range testData {
		log.Debug("writing test data")
		_, err := file.WriteString(data)
		Ok(t, err)

		// Reads the message from the websocket that was written in the function being tested.
		_, msg, err := ws.ReadMessage()
		Ok(t, err)

		Equals(t, data, string(msg))
	}
}
