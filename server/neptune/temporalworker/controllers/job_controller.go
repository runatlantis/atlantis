package controllers

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/logging"
	neptune "github.com/runatlantis/atlantis/server/neptune/temporalworker/config"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/controllers/websocket"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/job"
	"github.com/uber-go/tally/v4"
)

type multiplexor interface {
	Handle(w http.ResponseWriter, r *http.Request) error
}

type receiverRegistry interface {
	AddReceiver(jobID string, ch chan string)
}

type store interface {
	Get(jobID string) (*job.Job, error)
}

type JobsController struct {
	AtlantisVersion string
	AtlantisURL     *url.URL

	WsMux               multiplexor
	JobsProjectTemplate templates.TemplateWriter
	KeyGenerator        JobKeyGenerator

	StatsScope tally.Scope
	Logger     logging.Logger
}

func NewJobsController(
	store store,
	receiverRegistry receiverRegistry,
	serverCfg neptune.ServerConfig,
	scope tally.Scope,
	logger logging.Logger,
) *JobsController {
	jobPartitionRegistry := job.PartitionRegistry{
		ReceiverRegistry: receiverRegistry,
		Store:            store,
		Logger:           logger,
	}

	keyGenerator := JobKeyGenerator{}

	wsMux := websocket.NewInstrumentedMultiplexor(
		websocket.NewMultiplexor(
			logger,
			keyGenerator,
			jobPartitionRegistry,
		),
		scope.SubScope("http.getjob"),
	)

	return &JobsController{
		AtlantisVersion:     serverCfg.Version,
		AtlantisURL:         serverCfg.URL,
		KeyGenerator:        keyGenerator,
		StatsScope:          scope,
		Logger:              logger,
		JobsProjectTemplate: templates.ProjectJobsTemplate,
		WsMux:               wsMux,
	}
}

func (j *JobsController) getProjectJobs(w http.ResponseWriter, r *http.Request) error {
	jobID, err := j.KeyGenerator.Generate(r)
	if err != nil {
		j.respond(w, http.StatusBadRequest, err.Error())
		return err
	}

	viewData := templates.ProjectJobData{
		AtlantisVersion: j.AtlantisVersion,
		ProjectPath:     jobID,
		CleanedBasePath: j.AtlantisURL.Path,
	}
	if err = j.JobsProjectTemplate.Execute(w, viewData); err != nil {
		j.Logger.Error(err.Error())
		return err
	}

	return nil
}

func (j *JobsController) GetProjectJobs(w http.ResponseWriter, r *http.Request) {
	errorCounter := j.StatsScope.Counter(metrics.ExecutionErrorMetric)
	err := j.getProjectJobs(w, r)
	if err != nil {
		errorCounter.Inc(1)
	}
}

func (j *JobsController) getProjectJobsWS(w http.ResponseWriter, r *http.Request) error {
	err := j.WsMux.Handle(w, r)
	if err != nil {
		j.respond(w, http.StatusBadRequest, err.Error())
		return err
	}
	return nil
}

func (j *JobsController) GetProjectJobsWS(w http.ResponseWriter, r *http.Request) {
	jobsMetric := j.StatsScope.SubScope("api")
	errorCounter := jobsMetric.Counter(metrics.ExecutionErrorMetric)
	executionTime := jobsMetric.Timer(metrics.ExecutionTimeMetric).Start()
	defer executionTime.Stop()

	err := j.getProjectJobsWS(w, r)
	if err != nil {
		errorCounter.Inc(1)
	}
}

func (j *JobsController) respond(w http.ResponseWriter, responseCode int, format string, args ...interface{}) {
	response := fmt.Sprintf(format, args...)
	j.Logger.Error(response)
	w.WriteHeader(responseCode)
	fmt.Fprintln(w, response)
}

// JobKeyGenerator generates job id from the request
type JobKeyGenerator struct{}

func (g JobKeyGenerator) Generate(r *http.Request) (string, error) {
	jobID, ok := mux.Vars(r)["job-id"]
	if !ok {
		return "", fmt.Errorf("internal error: no job-id in route")
	}

	return jobID, nil
}
