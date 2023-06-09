package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/metrics"
	tally "github.com/uber-go/tally/v4"
)

type JobIDKeyGenerator struct{}

func (g JobIDKeyGenerator) Generate(r *http.Request) (string, error) {
	jobID, ok := mux.Vars(r)["job-id"]
	if !ok {
		return "", fmt.Errorf("internal error: no job-id in route")
	}

	return jobID, nil
}

type JobsController struct {
	AtlantisVersion          string
	AtlantisURL              *url.URL
	Logger                   logging.SimpleLogging
	ProjectJobsTemplate      templates.TemplateWriter
	ProjectJobsErrorTemplate templates.TemplateWriter
	Backend                  locking.Backend
	WsMux                    *websocket.Multiplexor
	KeyGenerator             JobIDKeyGenerator
	StatsScope               tally.Scope
}

func (j *JobsController) getProjectJobs(w http.ResponseWriter, r *http.Request) error {
	jobID, err := j.KeyGenerator.Generate(r)

	if err != nil {
		j.respond(w, logging.Error, http.StatusBadRequest, err.Error())
		return err
	}

	viewData := templates.ProjectJobData{
		AtlantisVersion: j.AtlantisVersion,
		ProjectPath:     jobID,
		CleanedBasePath: j.AtlantisURL.Path,
	}

	return j.ProjectJobsTemplate.Execute(w, viewData)
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	errorCounter := j.StatsScope.SubScope("getprojectjobs").Counter(metrics.ExecutionErrorMetric)
	err := j.getProjectJobs(w, r)
	if err != nil {
		j.Logger.Err(err.Error())
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
