package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/logging"
)

type LogStreamingController struct {
	AtlantisVersion   string
	AtlantisURL       *url.URL
	Logger            logging.SimpleLogging
	LogStreamTemplate templates.TemplateWriter
}

// Gets the PR information from the HTTP request params
func getPullInfo(r *http.Request) (string, error) {
	org, ok := mux.Vars(r)["org"]
	if !ok {
		return "", fmt.Errorf("Internal error: no org in route")
	}
	repo, ok := mux.Vars(r)["repo"]
	if !ok {
		return "", fmt.Errorf("Internal error: no repo in route")
	}
	pull, ok := mux.Vars(r)["pull"]
	if !ok {
		return "", fmt.Errorf("Internal error: no pull in route")
	}

	project, ok := mux.Vars(r)["project"]
	if !ok {
		return "", fmt.Errorf("Internal error: no project in route")
	}

	return fmt.Sprintf("%s/%s/%s/%s", org, repo, pull, project), nil
}

func (j *LogStreamingController) GetLogStream(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := getPullInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	viewData := templates.LogStreamData{
		AtlantisVersion: j.AtlantisVersion,
		PullInfo:        pullInfo,
		CleanedBasePath: j.AtlantisURL.Path,
	}

	err = j.LogStreamTemplate.Execute(w, viewData)
	if err != nil {
		j.Logger.Err(err.Error())
	}
}

var upgrader = websocket.Upgrader{}

func (j *LogStreamingController) GetLogStreamWS(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := getPullInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println(pullInfo)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		j.Logger.Warn("Failed to upgrade websocket: %s", err)
		return
	}

	defer c.Close()

	//for {
	if err := c.WriteMessage(websocket.BinaryMessage, []byte("Hello world"+"\r\n")); err != nil {
		j.Logger.Warn("Failed to write ws message: %s", err)
		return
	}
	//time.Sleep(1 * time.Second)
	//}
}

func (j *LogStreamingController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
