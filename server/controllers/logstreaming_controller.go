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
	"github.com/runatlantis/atlantis/server/logging"
)

type LogStreamingController struct {
	AtlantisVersion        string
	AtlantisURL            *url.URL
	Logger                 logging.SimpleLogging
	LogStreamTemplate      templates.TemplateWriter
	LogStreamErrorTemplate templates.TemplateWriter
	Db                     *db.BoltDB
	//TerraformOutputChan <-chan *models.TerraformOutputLine
}

type PullInfo struct {
	Org         string
	Repo        string
	Pull        int
	ProjectName string
}

func (p *PullInfo) String() string {
	return fmt.Sprintf("%s/%s/%d/%s", p.Org, p.Repo, p.Pull, p.ProjectName)
}

// Gets the PR information from the HTTP request params
func newPullInfo(r *http.Request) (*PullInfo, error) {
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
		if err = j.LogStreamErrorTemplate.Execute(w, err); err != nil {
			j.Logger.Err(err.Error())
		}
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

var upgrader = websocket.Upgrader{}

func (j *LogStreamingController) GetLogStreamWS(w http.ResponseWriter, r *http.Request) {
	pullInfo, err := newPullInfo(r)
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
	if err := c.WriteMessage(websocket.BinaryMessage, []byte("Hello World"+"\r\n")); err != nil {
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

//var pullObj, err = LogStreamingController.Vcs.GetPullRequestFromName() //Do I still need to create this object?
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
