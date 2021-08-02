package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
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

type LogStreamingController struct {
	AtlantisVersion        string
	AtlantisURL            *url.URL
	Logger                 logging.SimpleLogging
	LogStreamTemplate      templates.TemplateWriter
	LogStreamErrorTemplate templates.TemplateWriter
	Db                     *db.BoltDB
	TerraformOutputChan    chan *models.TerraformOutputLine

	logBuffers       map[string][]string
	wsChans          map[string]map[chan string]bool
	chanLock         sync.RWMutex
	WebsocketHandler WebsocketHandler
}

type PullInfo struct {
	Org         string
	Repo        string
	Pull        int
	ProjectName string
}

//Add channel to get Terraform output for current PR project
//Send output currently in buffer
func (j *LogStreamingController) addChan(pull string) chan string {
	ch := make(chan string, 1000)
	j.chanLock.Lock()
	for _, line := range j.logBuffers[pull] {
		ch <- line
	}
	if j.wsChans == nil {
		j.wsChans = map[string]map[chan string]bool{}
	}
	if j.wsChans[pull] == nil {
		j.wsChans[pull] = map[chan string]bool{}
	}
	j.wsChans[pull][ch] = true
	j.chanLock.Unlock()
	return ch
}

//Remove channel, so client no longer receives Terraform output
func (j *LogStreamingController) removeChan(pull string, ch chan string) {
	j.chanLock.Lock()
	delete(j.wsChans[pull], ch)
	j.chanLock.Unlock()
}

//Add log line to buffer and send to all current channels
func (j *LogStreamingController) writeLogLine(pull string, line string) {
	j.chanLock.Lock()
	if j.logBuffers == nil {
		j.logBuffers = map[string][]string{}
	}
	j.Logger.Info("Project info: %s, content: %s", pull, line)

	for ch := range j.wsChans[pull] {
		select {
		case ch <- line:
		default:
			delete(j.wsChans[pull], ch)
		}
	}
	if j.logBuffers[pull] == nil {
		j.logBuffers[pull] = []string{}
	}
	j.logBuffers[pull] = append(j.logBuffers[pull], line)
	j.chanLock.Unlock()
}

//Clear log lines in buffer
func (j *LogStreamingController) clearLogLines(pull string) {
	j.chanLock.Lock()
	delete(j.logBuffers, pull)
	j.chanLock.Unlock()
}

func (j *LogStreamingController) Listen() {
	for msg := range j.TerraformOutputChan {
		j.Logger.Info("Recieving message %s", msg.Line)
		if msg.ClearBuffBefore {
			j.clearLogLines(msg.ProjectInfo)
		}
		j.writeLogLine(msg.ProjectInfo, msg.Line)
		if msg.ClearBuffAfter {
			j.clearLogLines(msg.ProjectInfo)
		}
	}
}

func (p *PullInfo) String() string {
	return fmt.Sprintf("%s/%s/%d/%s", p.Org, p.Repo, p.Pull, p.ProjectName)
}

// Gets the PR information from the HTTP request params
func newPullInfo(r *http.Request) (*PullInfo, error) {
	fmt.Printf("%+v", r)
	fmt.Printf("\n%+v", mux.Vars(r))

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

	project, ok := mux.Vars(r)["project"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no project in route")
	}
	pullNum, err := strconv.Atoi(pull)
	if err != nil {
		return nil, err
	}

	return &PullInfo{
		Org:         org,
		Repo:        repo,
		Pull:        pullNum,
		ProjectName: project,
	}, nil
}

func (j *LogStreamingController) GetLogStream(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := newPullInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}
	//check db if it has pull project ret true or false(404)
	exist, err := j.RetrievePrStatus(pullInfo)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	if !exist {
		if err := j.LogStreamErrorTemplate.Execute(w, err); err != nil {
			j.Logger.Err(err.Error())
			j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
			return
		}
		j.respond(w, logging.Warn, http.StatusNotFound, "")
		return
	}

	viewData := templates.LogStreamData{
		AtlantisVersion: j.AtlantisVersion,
		PullInfo:        pullInfo.String(),
		CleanedBasePath: j.AtlantisURL.Path,
	}

	err = j.LogStreamTemplate.Execute(w, viewData)
	if err != nil {
		j.Logger.Err(err.Error())
	}
}

func (j *LogStreamingController) GetLogStreamWS(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := newPullInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return
	}

	c, err := j.WebsocketHandler.Upgrade(w, r, nil)

	if err != nil {
		j.Logger.Warn("Failed to upgrade websocket: %s", err)
		return
	}

	defer c.Close()

	pull := pullInfo.String()
	ch := j.addChan(pull)
	defer j.removeChan(pull, ch)

	for msg := range ch {
		j.Logger.Info(msg)
		if err := c.WriteMessage(websocket.BinaryMessage, []byte(msg+"\r\n\t")); err != nil {
			j.Logger.Warn("Failed to write ws message: %s", err)
			return
		}
	}
}

func (j *LogStreamingController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

//repo, pull num, project name moved to db
func (j *LogStreamingController) RetrievePrStatus(pullInfo *PullInfo) (bool, error) { //either implement new func in boltdb
	pull := models.PullRequest{
		Num: pullInfo.Pull,
		BaseRepo: models.Repo{
			FullName: fmt.Sprintf("%s/%s", pullInfo.Org, pullInfo.Repo),
			Owner:    pullInfo.Org,
			Name:     pullInfo.Repo,
			VCSHost: models.VCSHost{
				Hostname: "github.com",
				Type:     models.Github,
			},
		},
	}

	d, err := j.Db.GetPullStatus(pull)
	//Checks if pull request status exists
	if err != nil {
		//Checks if project array contains said project
		return false, err
	}
	if d != nil {
		for _, ps := range d.Projects {
			if ps.ProjectName == pullInfo.ProjectName {
				return true, nil
			}
		}
	}
	return false, nil
}
