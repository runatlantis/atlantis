package server

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/logging"
)

type LogStreamingController struct {
	AtlantisVersion   string
	AtlantisURL       *url.URL
	Logger            logging.SimpleLogging
	LogStreamTemplate TemplateWriter
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
		fmt.Println(err) //j.respond(w, logging.Error, http.StatusInternalServerError, err.Error()) //come back to this***
		return
	}

	viewData := logStreamData{
		AtlantisVersion: j.AtlantisVersion,
		PullInfo:        pullInfo,
		CleanedBasePath: j.AtlantisURL.Path,
	}

	if err := j.LogStreamTemplate.Execute(w, viewData); err != nil {
		j.Logger.Err(err.Error()) //***Go through simplelogger to see if there is a better
	}
}

var upgrader = websocket.Upgrader{}

func (j *LogStreamingController) GetLogStreamWS(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := getPullInfo(r)
	if err != nil {
		fmt.Println(err)
		//j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	fmt.Println(pullInfo)

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err) //.Logger.Warn("Failed to upgrade websocket: %s", err)
		return
	}

	defer c.Close()

	ch := make(chan string, 1000)
	//Need to output something simple, add 1sec buffer
	for msg := range ch {
		_ = msg
		time.Sleep(1 * time.Second)
	}
}

/*
func (j *LogStreamingController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
*/
