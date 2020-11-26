package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	. "github.com/runatlantis/atlantis/testing"
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
	tfOutputFileName := "20201121175848-runatalntis_atlantis-1-1a2b3c4-test-default-init"
	file, err := os.Create(filepath.Join(tmp, tfOutputFileName))
	Ok(t, err)

	// Create a test server for the websocket
	s := httptest.NewServer(http.HandlerFunc(controller.GetTfOutput))
	defer s.Close()

	// Creates the URL with all the parameters required
	url := fmt.Sprintf("%s%s?createdAt=%s&fullRepoName=%s&pullNr=%s&project=%s&headCommit=%s&workspace=%s&tfCommand=%s",
		"ws",
		strings.TrimPrefix(s.URL, "http"),
		"20201121175848",
		"runatalntis_atlantis",
		"1",
		"1a2b3c4",
		"test",
		"default",
		"init",
	)

	fmt.Println(url)

	// Connect to the server
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	Ok(t, err)
	defer ws.Close()

	testData := []string{"ab", "cd", "ef"}
	for _, data := range testData {
		_, err := file.WriteString(data)
		Ok(t, err)

		// Reads the message from the websocket that was written in the function being tested.
		_, msg, err := ws.ReadMessage()
		Ok(t, err)

		Equals(t, data, string(msg))
	}
}
