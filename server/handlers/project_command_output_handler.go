package handlers

import (
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// AsyncProjectCommandOutputHandler is a handler to transport terraform client
// outputs to the front end.
type AsyncProjectCommandOutputHandler struct {
	projectCmdOutput chan *models.ProjectCmdOutputLine

	projectOutputBuffers     map[string][]string
	projectOutputBuffersLock sync.RWMutex

	receiverBuffers     map[string]map[chan string]bool
	receiverBuffersLock sync.RWMutex

	projectStatusUpdater   ProjectStatusUpdater
	projectJobURLGenerator ProjectJobURLGenerator

	logger logging.SimpleLogging
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_job_url_generator.go ProjectJobURLGenerator

// ProjectJobURLGenerator generates urls to view project's progress.
type ProjectJobURLGenerator interface {
	GenerateProjectJobURL(p models.ProjectCommandContext) string
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_status_updater.go ProjectStatusUpdater

type ProjectStatusUpdater interface {
	// UpdateProject sets the commit status for the project represented by
	// ctx.
	UpdateProject(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus, url string) error
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_output_handler.go ProjectCommandOutputHandler

type ProjectCommandOutputHandler interface {
	// Clear clears the buffer from previous terraform output lines
	Clear(ctx models.ProjectCommandContext)

	// Send will enqueue the msg and wait for Handle() to receive the message.
	Send(ctx models.ProjectCommandContext, msg string)

	// Receive will create a channel for projectPullInfo and run a callback function argument when the new channel receives a message.
	Receive(projectInfo string, receiver chan string, callback func(msg string) error) error

	// Listens for msg from channel
	Handle()

	// SetJobURLWithStatus sets the commit status for the project represented by
	// ctx and updates the status with and url to a job.
	SetJobURLWithStatus(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus) error
}

func NewAsyncProjectCommandOutputHandler(
	projectCmdOutput chan *models.ProjectCmdOutputLine,
	projectStatusUpdater ProjectStatusUpdater,
	projectJobURLGenerator ProjectJobURLGenerator,
	logger logging.SimpleLogging,
) ProjectCommandOutputHandler {
	return &AsyncProjectCommandOutputHandler{
		projectCmdOutput:       projectCmdOutput,
		logger:                 logger,
		receiverBuffers:        map[string]map[chan string]bool{},
		projectStatusUpdater:   projectStatusUpdater,
		projectJobURLGenerator: projectJobURLGenerator,
		projectOutputBuffers:   map[string][]string{},
	}
}

func (p *AsyncProjectCommandOutputHandler) Send(ctx models.ProjectCommandContext, msg string) {
	p.projectCmdOutput <- &models.ProjectCmdOutputLine{
		ProjectInfo: ctx.PullInfo(),
		Line:        msg,
	}
}

func (p *AsyncProjectCommandOutputHandler) Receive(projectInfo string, receiver chan string, callback func(msg string) error) error {
	// Avoid deadlock when projectOutputBuffer size is greater than the channel (currently set to 1000)
	// Running this as a goroutine allows for the channel to be read in callback
	go p.addChan(receiver, projectInfo)
	defer p.cleanUp(projectInfo, receiver)

	for msg := range receiver {
		if err := callback(msg); err != nil {
			return err
		}
	}

	return nil
}

func (p *AsyncProjectCommandOutputHandler) Handle() {
	for msg := range p.projectCmdOutput {
		if msg.ClearBuffBefore {
			p.clearLogLines(msg.ProjectInfo)
			continue
		}
		p.writeLogLine(msg.ProjectInfo, msg.Line)
	}
}

func (p *AsyncProjectCommandOutputHandler) Clear(ctx models.ProjectCommandContext) {
	p.projectCmdOutput <- &models.ProjectCmdOutputLine{
		ProjectInfo:     ctx.PullInfo(),
		ClearBuffBefore: true,
		Line:            "",
	}
}

func (p *AsyncProjectCommandOutputHandler) SetJobURLWithStatus(ctx models.ProjectCommandContext, cmdName models.CommandName, status models.CommitStatus) error {
	url := p.projectJobURLGenerator.GenerateProjectJobURL(ctx)
	return p.projectStatusUpdater.UpdateProject(ctx, cmdName, status, url)
}

func (p *AsyncProjectCommandOutputHandler) clearLogLines(pull string) {
	p.projectOutputBuffersLock.Lock()
	delete(p.projectOutputBuffers, pull)
	p.projectOutputBuffersLock.Unlock()
}

func (p *AsyncProjectCommandOutputHandler) addChan(ch chan string, pull string) {
	p.receiverBuffersLock.Lock()
	if p.receiverBuffers[pull] == nil {
		p.receiverBuffers[pull] = map[chan string]bool{}
	}
	p.receiverBuffers[pull][ch] = true
	p.receiverBuffersLock.Unlock()

	p.projectOutputBuffersLock.RLock()
	for _, line := range p.projectOutputBuffers[pull] {
		ch <- line
	}
	p.projectOutputBuffersLock.RUnlock()
}

//Add log line to buffer and send to all current channels
func (p *AsyncProjectCommandOutputHandler) writeLogLine(pull string, line string) {
	p.receiverBuffersLock.Lock()
	for ch := range p.receiverBuffers[pull] {
		select {
		case ch <- line:
		default:
			// Client ws conn could be closed in two ways:
			// 1. Client closes the conn gracefully -> the closeHandler() is executed which
			//  	closes the channel and cleans up resources.
			// 2. Client does not close the conn and the closeHandler() is not executed -> the
			// 		receiverChan will be blocking for N number of messages (equal to buffer size)
			// 		before we delete the channel and clean up the resources.
			delete(p.receiverBuffers[pull], ch)
		}
	}
	p.receiverBuffersLock.Unlock()

	p.projectOutputBuffersLock.Lock()
	if p.projectOutputBuffers[pull] == nil {
		p.projectOutputBuffers[pull] = []string{}
	}
	p.projectOutputBuffers[pull] = append(p.projectOutputBuffers[pull], line)
	p.projectOutputBuffersLock.Unlock()
}

//Remove channel, so client no longer receives Terraform output
func (p *AsyncProjectCommandOutputHandler) cleanUp(pull string, ch chan string) {
	p.receiverBuffersLock.Lock()
	delete(p.receiverBuffers[pull], ch)
	p.receiverBuffersLock.Unlock()
}

func (p *AsyncProjectCommandOutputHandler) GetReceiverBufferForPull(pull string) map[chan string]bool {
	return p.receiverBuffers[pull]
}

func (p *AsyncProjectCommandOutputHandler) GetProjectOutputBuffer(pull string) []string {
	return p.projectOutputBuffers[pull]
}
