package handlers

import (
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type DefaultProjectCommandOutputHandler struct {
	// this is TerraformOutputChan
	ProjectCmdOutput chan *models.ProjectCmdOutputLine
	// this logBuffers
	projectOutputBuffers     map[string][]string
	projectOutputBuffersLock sync.RWMutex

	// this is wsChans
	receiverBuffers     map[string]map[chan string]bool
	receiverBuffersLock sync.RWMutex

	logger logging.SimpleLogging
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_output_handler.go ProjectCommandOutputHandler

type ProjectCommandOutputHandler interface {
	// Send will enqueue the msg and wait for Handle() to receive the message.
	Send(ctx models.ProjectCommandContext, msg string)

	// Clears buffer for new project to run
	Clear(ctx models.ProjectCommandContext)

	// Receive will create a channel for projectPullInfo and run a callback function argument when the new channel receives a message.
	Receive(projectInfo string, receiver chan string, callback func(msg string) error) error

	// Listens for msg from channel
	Handle()
}

func NewProjectCommandOutputHandler(projectCmdOutput chan *models.ProjectCmdOutputLine, logger logging.SimpleLogging) ProjectCommandOutputHandler {
	return &DefaultProjectCommandOutputHandler{
		ProjectCmdOutput:     projectCmdOutput,
		logger:               logger,
		receiverBuffers:      map[string]map[chan string]bool{},
		projectOutputBuffers: map[string][]string{},
	}
}

func (p *DefaultProjectCommandOutputHandler) Send(ctx models.ProjectCommandContext, msg string) {
	p.ProjectCmdOutput <- &models.ProjectCmdOutputLine{
		ProjectInfo: ctx.PullInfo(),
		Line:        msg,
	}
}

func (p *DefaultProjectCommandOutputHandler) Receive(projectInfo string, receiver chan string, callback func(msg string) error) error {

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

func (p *DefaultProjectCommandOutputHandler) Handle() {
	for msg := range p.ProjectCmdOutput {
		if msg.ClearBuffBefore {
			p.clearLogLines(msg.ProjectInfo)
			continue
		}
		p.writeLogLine(msg.ProjectInfo, msg.Line)
	}
}

func (p *DefaultProjectCommandOutputHandler) Clear(ctx models.ProjectCommandContext) {
	p.ProjectCmdOutput <- &models.ProjectCmdOutputLine{
		ProjectInfo:     ctx.PullInfo(),
		Line:            "",
		ClearBuffBefore: true,
	}
}

func (p *DefaultProjectCommandOutputHandler) clearLogLines(pull string) {
	p.projectOutputBuffersLock.Lock()
	delete(p.projectOutputBuffers, pull)
	p.projectOutputBuffersLock.Unlock()
}

func (p *DefaultProjectCommandOutputHandler) addChan(ch chan string, pull string) {
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
func (p *DefaultProjectCommandOutputHandler) writeLogLine(pull string, line string) {
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
func (p *DefaultProjectCommandOutputHandler) cleanUp(pull string, ch chan string) {
	p.receiverBuffersLock.Lock()
	delete(p.receiverBuffers[pull], ch)
	p.receiverBuffersLock.Unlock()
}

func (p *DefaultProjectCommandOutputHandler) GetReceiverBufferForPull(pull string) map[chan string]bool {
	return p.receiverBuffers[pull]
}

func (p *DefaultProjectCommandOutputHandler) GetProjectOutputBuffer(pull string) []string {
	return p.projectOutputBuffers[pull]
}
