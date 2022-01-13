package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally"
)

type JobsController struct {
	AtlantisVersion          string
	AtlantisURL              *url.URL
	Logger                   logging.SimpleLogging
	ProjectJobsTemplate      templates.TemplateWriter
	ProjectJobsErrorTemplate templates.TemplateWriter
	Db                       *db.BoltDB
	WsMux                    *websocket.Multiplexor
	StatsScope               tally.Scope
}

type ProjectInfoKeyGenerator struct{}

func (g ProjectInfoKeyGenerator) Generate(r *http.Request) (string, error) {
	projectInfo, err := newProjectInfo(r)

	if err != nil {
		return "", errors.Wrap(err, "creating project info")
	}

	return projectInfo.String(), nil
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
	workspace   string
	pullInfo
}

func (p *projectInfo) String() string {
	return fmt.Sprintf("%s/%s/%d/%s/%s", p.org, p.repo, p.pull, p.projectName, p.workspace)
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

	workspace, ok := mux.Vars(r)["workspace"]
	if !ok {
		return nil, fmt.Errorf("Internal error: no workspace in route")
	}

	return &projectInfo{
		pullInfo:    *pullInfo,
		projectName: project,
		workspace:   workspace,
	}, nil
}

func (j *JobsController) getProjectJobs(w http.ResponseWriter, r *http.Request) error {
	projectInfo, err := newProjectInfo(r)
	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return err
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
		return err
	}

	return nil
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	errorCounter := j.StatsScope.SubScope("getprojectjobs").Counter(metrics.ExecutionErrorMetric)
	err := j.getProjectJobs(w, r)
	if err != nil {
		errorCounter.Inc(1)
	}
}

func (j *JobsController) getProjectJobsWS(w http.ResponseWriter, r *http.Request) error {
	err := j.WsMux.Handle(w, r)

	if err != nil {
		j.respond(w, logging.Error, http.StatusInternalServerError, err.Error())
		return err
	}

	return nil
}

func (j *JobsController) GetProjectJobsWS(w http.ResponseWriter, r *http.Request) {
	jobsMetric := j.StatsScope.SubScope("getprojectjobs")
	errorCounter := jobsMetric.Counter(metrics.ExecutionErrorMetric)
	executionTime := jobsMetric.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	err := j.getProjectJobsWS(w, r)
	if err != nil {
		errorCounter.Inc(1)
	}
}

func (j *JobsController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
