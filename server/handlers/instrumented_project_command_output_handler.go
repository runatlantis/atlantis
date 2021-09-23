package handlers

import (
	"fmt"

	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/logging"
)

type InstrumentedProjectCommandOutputHandler struct {
	ProjectCommandOutputHandler
	numWSConnnections stats.Gauge
	logger            logging.SimpleLogging
}

func NewInstrumentedProjectCommandOutputHandler(prjCmdOutputHandler ProjectCommandOutputHandler, statsScope stats.Scope, logger logging.SimpleLogging) ProjectCommandOutputHandler {
	return &InstrumentedProjectCommandOutputHandler{
		ProjectCommandOutputHandler: prjCmdOutputHandler,
		numWSConnnections:           statsScope.Scope("getprojectjobs").Scope("websocket").NewGauge("connections"),
		logger:                      logger,
	}
}

func (p *InstrumentedProjectCommandOutputHandler) Receive(projectInfo string, receiver chan string, callback func(msg string) error) error {
	p.numWSConnnections.Inc()
	defer func() {
		// Log message to ensure numWSConnnections gauge is being updated properly.
		// [ORCA-955] TODO: Remove when removing the feature flag for log streaming.
		p.logger.Info(fmt.Sprintf("Decreasing num of ws connections for project: %s", projectInfo))
		p.numWSConnnections.Dec()
	}()
	return p.ProjectCommandOutputHandler.Receive(projectInfo, receiver, callback)
}
