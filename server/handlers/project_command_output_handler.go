package handlers

import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type DefaultProjectCommandOutputHandler struct {
	// this is TerraformOutputChan
	ProjectCmdOutput chan *models.ProjectCmdOutputLine
	// this logBuffers
	projectOutputBuffers map[string][]string
	// this is wsChans
	receiverBuffers map[string]map[chan string]bool
	// same as chanLock
	controllerBufferLock sync.RWMutex
	logger               logging.SimpleLogging
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_output_handler.go ProjectCommandOutputHandler

type ProjectCommandOutputHandler interface {
	// Send will enqueue the msg and wait for Handle() to receive the message.
	Send(ctx models.ProjectCommandContext, msg string)

	// Clears buffer for new project to run
	Clear(ctx models.ProjectCommandContext)

	// Receive will create a channel for projectPullInfo and run a callback function argument when the new channel receives a message.
	Receive(projectInfo string, callback func(msg string) error) error

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

func (p *DefaultProjectCommandOutputHandler) Receive(projectInfo string, callback func(msg string) error) error {
	ch := p.addChan(projectInfo)
	defer p.removeChan(projectInfo, ch)

	for msg := range ch {
		if err := callback(msg); err != nil {
			return err
		}
	}

	return nil
}

func (p *DefaultProjectCommandOutputHandler) Handle() {
	fmt.Printf("Testing Handle func")
	for msg := range p.ProjectCmdOutput {
		p.logger.Info("Receiving message %s", msg.Line)
		fmt.Printf("Receiving message %s", msg.Line)
		if msg.ClearBuffBefore {
			p.clearLogLines(msg.ProjectInfo)
		}
		p.writeLogLine(msg.ProjectInfo, msg.Line)
		if msg.ClearBuffAfter {
			p.clearLogLines(msg.ProjectInfo)
		}
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
	p.controllerBufferLock.Lock()
	delete(p.projectOutputBuffers, pull)
	p.controllerBufferLock.Unlock()
}

func (p *DefaultProjectCommandOutputHandler) addChan(pull string) chan string {
	ch := make(chan string, 1000)
	p.controllerBufferLock.Lock()
	for _, line := range p.projectOutputBuffers[pull] {
		ch <- line
	}
	if p.receiverBuffers[pull] == nil {
		p.receiverBuffers[pull] = map[chan string]bool{}
	}
	p.receiverBuffers[pull][ch] = true
	p.controllerBufferLock.Unlock()
	return ch
}

//Add log line to buffer and send to all current channels
func (p *DefaultProjectCommandOutputHandler) writeLogLine(pull string, line string) {
	p.controllerBufferLock.Lock()
	p.logger.Info("Project info: %s, content: %s", pull, line)

	for ch := range p.receiverBuffers[pull] {
		select {
		case ch <- line:
		default:
			delete(p.receiverBuffers[pull], ch)
		}
	}
	if p.projectOutputBuffers[pull] == nil {
		p.projectOutputBuffers[pull] = []string{}
	}
	p.projectOutputBuffers[pull] = append(p.projectOutputBuffers[pull], line)
	p.controllerBufferLock.Unlock()
}

//Remove channel, so client no longer receives Terraform output
func (p *DefaultProjectCommandOutputHandler) removeChan(pull string, ch chan string) {
	p.controllerBufferLock.Lock()
	delete(p.receiverBuffers[pull], ch)
	close(ch)
	p.controllerBufferLock.Unlock()
}
