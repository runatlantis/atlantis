package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/controllers/websocket"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/logging"
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
	Db                       *db.BoltDB
	WsMux                    *websocket.Multiplexor
	KeyGenerator             JobIDKeyGenerator
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	jobID, err := j.KeyGenerator.Generate(r)

	if err != nil {
		j.respond(w, logging.Error, http.StatusBadRequest, err.Error())
		return
	}

	viewData := templates.ProjectJobData{
		AtlantisVersion: j.AtlantisVersion,
		ProjectPath:     jobID,
		CleanedBasePath: j.AtlantisURL.Path,
	}

	if err = j.ProjectJobsTemplate.Execute(w, viewData); err != nil {
		j.Logger.Err(err.Error())
	}
}

func (j *JobsController) GetProjectJobsWS(w http.ResponseWriter, r *http.Request) {
	err := j.WsMux.Handle(w, r)

	if err != nil {
		j.respond(w, logging.Error, http.StatusBadRequest, err.Error())
		return
	}
}

func (j *JobsController) respond(w http.ResponseWriter, lvl logging.LogLevel, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Log(lvl, response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}
