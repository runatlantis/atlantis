package handlers

import (
	stats "github.com/lyft/gostats"
	"github.com/runatlantis/atlantis/server/events/models"
)

type InstrumentedProjectCommandOutputHandler struct {
	ProjectCommandOutputHandler
	numChans stats.Gauge
}

func NewInstrumentedProjectCommandOutputHandler(prjCmdOutputHandler ProjectCommandOutputHandler, statsScope stats.Scope) ProjectCommandOutputHandler {
	return &InstrumentedProjectCommandOutputHandler{
		ProjectCommandOutputHandler: prjCmdOutputHandler,
		numChans:                    statsScope.Scope("log_streaming").NewGauge("num_ws_chans"),
	}
}

func (p *InstrumentedProjectCommandOutputHandler) Clear(ctx models.ProjectCommandContext) {
	p.ProjectCommandOutputHandler.Clear(ctx)
}

func (p *InstrumentedProjectCommandOutputHandler) Send(ctx models.ProjectCommandContext, msg string) {
	p.ProjectCommandOutputHandler.Send(ctx, msg)
}

func (p *InstrumentedProjectCommandOutputHandler) Receive(projectInfo string, receiver chan string, callback func(msg string) error) error {
	p.numChans.Inc()
	defer p.numChans.Dec()
	return p.ProjectCommandOutputHandler.Receive(projectInfo, receiver, callback)
}

func (p *InstrumentedProjectCommandOutputHandler) Handle() {
	p.ProjectCommandOutputHandler.Handle()
}

func (p *InstrumentedProjectCommandOutputHandler) SetJobURLWithStatus(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus) error {
	return p.ProjectCommandOutputHandler.SetJobURLWithStatus(ctx, cmdName, status)
}

func (p *InstrumentedProjectCommandOutputHandler) CleanUp(pull string) {
	p.ProjectCommandOutputHandler.CleanUp(pull)
}
