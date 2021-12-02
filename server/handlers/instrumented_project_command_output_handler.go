package handlers

import (
	"fmt"

	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type InstrumentedProjectCommandOutputHandler struct {
	ProjectCommandOutputHandler
	numWSConnnections stats.Gauge
	logger            logging.SimpleLogging
}

func NewInstrumentedProjectCommandOutputHandler(projectCmdOutput chan *models.ProjectCmdOutputLine,
	projectStatusUpdater ProjectStatusUpdater,
	projectJobURLGenerator ProjectJobURLGenerator,
	logger logging.SimpleLogging,
	scope stats.Scope) ProjectCommandOutputHandler {
	prjCmdOutputHandler := NewAsyncProjectCommandOutputHandler(
		projectCmdOutput,
		projectStatusUpdater,
		projectJobURLGenerator,
		logger,
	)
	return &InstrumentedProjectCommandOutputHandler{
		ProjectCommandOutputHandler: prjCmdOutputHandler,
		numWSConnnections:           scope.Scope("getprojectjobs").Scope("websocket").NewGauge("connections"),
		logger:                      logger,
	}
}

func (p *InstrumentedProjectCommandOutputHandler) Register(projectInfo string, receiver chan string) {
	p.numWSConnnections.Inc()
	defer func() {
		// Log message to ensure numWSConnnections gauge is being updated properly.
		// [ORCA-955] TODO: Remove when removing the feature flag for log streaming.
		p.logger.Info(fmt.Sprintf("Decreasing num of ws connections for project: %s", projectInfo))
		p.numWSConnnections.Dec()
	}()
	p.ProjectCommandOutputHandler.Register(projectInfo, receiver)
}

func (p *InstrumentedProjectCommandOutputHandler) Deregister(projectInfo string, receiver chan string) {
	p.numWSConnnections.Inc()
	defer func() {
		// Log message to ensure numWSConnnections gauge is being updated properly.
		// [ORCA-955] TODO: Remove when removing the feature flag for log streaming.
		p.logger.Info(fmt.Sprintf("Decreasing num of ws connections for project: %s", projectInfo))
		p.numWSConnnections.Dec()
	}()
	p.ProjectCommandOutputHandler.Deregister(projectInfo, receiver)
}
