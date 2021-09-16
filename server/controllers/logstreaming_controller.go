package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/handlers"
	"github.com/runatlantis/atlantis/server/logging"
)

type JobsController struct {
	AtlantisVersion          string
	AtlantisURL              *url.URL
	Logger                   logging.SimpleLogging
	ProjectJobsTemplate      templates.TemplateWriter
	ProjectJobsErrorTemplate templates.TemplateWriter
	Db                       *db.BoltDB

	WebsocketHandler            handlers.WebsocketHandler
	ProjectCommandOutputHandler handlers.ProjectCommandOutputHandler
}

type pullInfo struct {
	org  string
	repo string
	pull int
}

func (p *pullInfo) String() string {
	return fmt.Sprintf("%s/%s/%d", p.org, p.repo, p.pull)
}

type projectInfo struct {
	projectName string
	pullInfo
}

func (p *projectInfo) String() string {
	return fmt.Sprintf("%s/%s/%d/%s", p.org, p.repo, p.pull, p.projectName)
}

func newPullInfo(r *http.Request) (*pullInfo, error) {
	org, ok := mux.Vars(r)["org"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no org in route")
	}
	repo, ok := mux.Vars(r)["repo"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no repo in route")
	}
	pull, ok := mux.Vars(r)["pull"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no pull in route")
	}
	pullNum, err := strconv.Atoi(pull)
	if err != nil {
		return nil, err
	}

	return &pullInfo{
		org:  org,
		repo: repo,
		pull: pullNum,
	}, nil
}

// Gets the PR information from the HTTP request params
func newProjectInfo(r *http.Request) (*projectInfo, error) {
	pullInfo, err := newPullInfo(r)
	if err != nil {
		return nil, err
	}

	project, ok := mux.Vars(r)["project"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no project in route")
	}

	return &projectInfo{
		pullInfo:    *pullInfo,
		projectName: project,
	}, nil
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	projectInfo, err := newProjectInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	viewData := templates.ProjectJobData{
		AtlantisVersion: j.AtlantisVersion,
		ProjectPath:     projectInfo.String(),
		CleanedBasePath: j.AtlantisURL.Path,
		ClearMsg:        models.LogStreamingClearMsg,
	}

	err = j.ProjectJobsTemplate.Execute(w, viewData)
	if err != nil {
		j.Logger.Err(err.Error())
	}
}

func (j *JobsController) GetProjectJobsWS(w http.ResponseWriter, r *http.Request) {
	projectInfo, err := newProjectInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	c, err := j.WebsocketHandler.Upgrade(w, r, nil)
	if err != nil {
		j.Logger.Warn("Failed to upgrade websocket: %s", err)
		return
	}

	// Buffer size set to 1000 to ensure messages get queued (upto 1000) if the receiverCh is not ready to
	// receive messages before the channel is closed and resources cleaned up.
	receiver := make(chan string, 1000)
	j.WebsocketHandler.SetCloseHandler(c, receiver)

	// Add a reader goroutine to listen for socket.close() events.
	go j.WebsocketHandler.SetReadHandler(c)

	pull := projectInfo.String()
	err = j.ProjectCommandOutputHandler.Receive(pull, receiver, func(msg string) error {
		if err := c.WriteMessage(websocket.BinaryMessage, []byte(msg+"\r\n\t")); err != nil {
			j.Logger.Warn("Failed to write ws message: %s", err)
			return err
		}
		return nil
	})

	if err != nil {
		j.Logger.Warn("Failed to receive message: %s", err)
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}
}

func (j *JobsController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
