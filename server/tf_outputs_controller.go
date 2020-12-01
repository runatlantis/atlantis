package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// TfOutputController holds the attributes for the terraform outputs controller
type TfOutputController struct {
	AtlantisVersion        string
	AtlantisURL            *url.URL
	Log                    *logging.SimpleLogger
	TfOutputHelper         terraform.OutputHelper
	WsUpgrader             websocket.Upgrader
	TfOutputDetailTemplate TemplateWriter
}

// requiredQuery type to describe all the required query strings
type requiredQuery string

const (
	createdAt          requiredQuery = "createdAt"
	createdAtFormatted requiredQuery = "createdAtFormatted"
	repoFullName       requiredQuery = "repoFullName"
	pullNum            requiredQuery = "pullNum"
	headCommit         requiredQuery = "headCommit"
	project            requiredQuery = "project"
	workspace          requiredQuery = "workspace"
	tfCommand          requiredQuery = "tfCommand"
)

// GetTfOutputDetail return the tf output detail page rendered.
func (t *TfOutputController) GetTfOutputDetail(w http.ResponseWriter, r *http.Request) {
	requiredQueries := []requiredQuery{createdAt, createdAtFormatted, repoFullName, pullNum, headCommit, project, workspace, tfCommand}
	queryValues := make(map[requiredQuery]string)
	for _, query := range requiredQueries {
		value, ok := r.URL.Query()[string(query)]
		// Verify if the query string exists in the request and only has one element.
		if !ok || len(query) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "missing %s query string", query)
			return
		}
		queryValues[query] = value[0]
	}

	pullNum, err := strconv.Atoi(queryValues[pullNum])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "error converting the pull request number, %v", err)
		return
	}

	viewData := TfOutputDetailData{
		AtlantisVersion:    t.AtlantisVersion,
		CleanedBasePath:    t.AtlantisURL.Path,
		CreatedAt:          queryValues[createdAt],
		CreatedAtFormatted: queryValues[createdAtFormatted],
		RepoFullName:       queryValues[repoFullName],
		PullNum:            pullNum,
		HeadCommit:         queryValues[headCommit],
		Project:            queryValues[project],
		Workspace:          queryValues[workspace],
		TfCommand:          queryValues[tfCommand],
	}

	err = t.TfOutputDetailTemplate.Execute(w, viewData)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error rendering tf output detail page, %v", err)
		return
	}
}

// GetTfOutputWebsocket opens the web socket to stream line the terraform output file being created or not by Atlantis.
func (t *TfOutputController) GetTfOutputWebsocket(w http.ResponseWriter, r *http.Request) {
	// Creates the websocket by upgrading the request
	c, err := t.WsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		t.Log.Err("fail to start websocket server, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Fail to start websocket server")
		return
	}
	defer c.Close()

	// Channel to close the method that is reading the tf output file continuously
	done := make(chan bool)
	c.SetCloseHandler(func(code int, text string) error {
		// Stopping reading the file (tail -f)
		done <- true
		// Close the tcp connection
		err = c.Close()
		if err != nil {
			return errors.Wrap(err, "can't close the web socket")
		}
		return nil
	})

	for {
		// Websocket start with the client sending a message
		mt, msg, err := c.ReadMessage()
		if err != nil {
			t.Log.Err("fail to read message from the websocket server, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Fail to read message from the websocket server")
			return
		}

		// The message contains the tf output file infos to stream the log
		tfOutputInfos := strings.Split(string(msg), "|")

		// Gets the tf output file name to stream
		tfOutputFileName, err := t.TfOutputHelper.FindOutputFile(
			tfOutputInfos[0],
			tfOutputInfos[1],
			tfOutputInfos[2],
			tfOutputInfos[3],
			tfOutputInfos[4],
			tfOutputInfos[5],
			tfOutputInfos[6],
		)
		if err != nil {
			t.Log.Err("can't find file %v", tfOutputInfos)
			err = c.WriteMessage(mt, []byte("file not found"))
			if err != nil {
				t.Log.Err("fail to write in the websocket server, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Fail to write in the websocket server")
			}
			return
		}

		// Call a go function to start the continue read file method, and as the channel will be receiving the messages
		// to reply back in the websocket
		fileLines := make(chan string)
		go func() {
			err := t.TfOutputHelper.ContinueReadFile(t.Log, tfOutputFileName, fileLines, done)
			if err != nil {
				err = c.WriteMessage(mt, []byte("failed to tail the file"))
				if err != nil {
					t.Log.Err("fail to write in the websocket server, %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, "Fail to write in the websocket server")
				}
				return
			}
		}()

		for {
			select {
			// For a new message in the channel, post it in the websocket
			case line := <-fileLines:
				err = c.WriteMessage(websocket.TextMessage, []byte(line))
				if err != nil {
					t.Log.Err("fail to write in the websocket, %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, "Fail to write in the websocket")
					return
				}
			}
		}
	}
}
