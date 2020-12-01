package server

import (
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// mockThfOutputHelper mocks terraform.OutputHelper to simulate reading the file as the pagemock framework could not
// deal with this scenario with channels.
type mockThfOutputHelper struct {
	mockFileLines chan string
}

func (m *mockThfOutputHelper) List() ([]terraform.TfOutputFile, error) {
	return nil, nil
}

func (m *mockThfOutputHelper) ParseFileName(fileName string) (terraform.TfOutputFile, error) {
	return terraform.TfOutputFile{}, nil
}

func (m *mockThfOutputHelper) CreateFileName(fullRepoName string, pullRequestNr int, headCommit string, project string, workspace string, tfCommand string) string {
	return ""
}

func (m *mockThfOutputHelper) ContinueReadFile(Log *logging.SimpleLogger, fileName string, fileLines chan<- string, done chan bool) error {
	for {
		select {
		// Gets the mock file line from the test
		case mockFileLine := <- m.mockFileLines:
			// Sends the mocked file line to the fileLines channel
			fileLines <- mockFileLine
		}
	}
}

func (m *mockThfOutputHelper) FindOutputFile(createdAt, fullRepoName, pullNr, headCommit, project, workspace, tfCommand string) (string, error) {
	return "", nil
}

func TestTfOutputsController_GetTfOutputWebsocket(t *testing.T) {
	mockFileLines := make(chan string)
	mockTfOutputHelper := &mockThfOutputHelper{
		mockFileLines: mockFileLines,
	}

	log := logging.NewSimpleLogger("test", false, logging.Debug)

	controller := TfOutputController{
		Log:            log,
		TfOutputHelper: mockTfOutputHelper,
		WsUpgrader:     websocket.Upgrader{},
	}

	// Create a test server for the websocket
	s := httptest.NewServer(http.HandlerFunc(controller.GetTfOutputWebsocket))
	defer s.Close()

	// Creates the URL for the websocket
	url, err := url.Parse(s.URL)
	Ok(t, err)
	url.Scheme = "ws"

	// Connect to the web socket server
	ws, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	Ok(t, err)
	defer ws.Close()

	// Writes the first message with the tf output file to "tail"
	// Here could be any message as the client should start the "conversation" in the websocket.
	err = ws.WriteMessage(websocket.TextMessage, []byte("20201121175848|runatalntis_atlantis|1|1a2b3c4|test|default|init"))
	Ok(t, err)

	testData := []string{"ab", "cd", "ef"}
	for _, data := range testData {
		// Write the test data into the channel that mocks the file being read
		mockFileLines <- data

		// Reads the message from the websocket that was written in the function being tested.
		_, msg, err := ws.ReadMessage()
		Ok(t, err)

		Equals(t, data, string(msg))
	}
}
