package handlers

import (
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type DefaultProjectCommandOutputHandler struct {
	// this is TerraformOutputChan
	projectCmdOutput chan *models.ProjectCmdOutputLine
	// this logBuffers
	projectOutputBuffers map[string][]string
	// this is wsChans
	controllerBuffers map[string]map[chan string]bool
	// same as chanLock
	controllerBufferLock sync.RWMutex
	logger               logging.SimpleLogging
}

type ProjectCommandOutputHandler interface {
	// Send will enqueue the msg and wait for Handle() to receive the message.
	// this will be called in the RunCommandAsync where we currently sending messages to the terraform channel
	Send(ctx models.ProjectCommandContext, msg string)

	// Receive will create a channel for projectPullInfo and run a callback function argument when the new channel receives a message. This will be called from the controller.
	// j.ProjectCommandOutputHandler.Receive(pull, func(msg string) {
	//   j.Logger.Info(msg)
	//   if err := c.WriteMessage(websocket.BinaryMessage, []byte(msg+"\r\n\t")); err != nil {
	//     j.Logger.Warn("Failed to write ws message: %s", err)
	//     return err
	//   }
	// })
	Receive(projectInfo string, callback func(msg string) error) error

	// Is basically the same as Listen function from logstreaming controller.
	Handle()
}

func (p *DefaultProjectCommandOutputHandler) Send(ctx models.ProjectCommandContext, msg string) {
	p.projectCmdOutput <- &models.ProjectCmdOutputLine{
		ProjectInfo: ctx.PullInfo(),
		Line:        msg,
	}
}

func (p *DefaultProjectCommandOutputHandler) Receive(projectInfo string, callback func(msg string) error) error {
	ch := make(chan string, 1000)
	defer p.removeChan(projectInfo, ch)
	p.controllerBufferLock.Lock()
	for _, line := range p.projectOutputBuffers[projectInfo] {
		ch <- line
	}
	if p.controllerBuffers == nil {
		p.controllerBuffers = map[string]map[chan string]bool{}
	}
	if p.controllerBuffers[projectInfo] == nil {
		p.controllerBuffers[projectInfo] = map[chan string]bool{}
	}
	p.controllerBuffers[projectInfo][ch] = true
	p.controllerBufferLock.Unlock()

	for msg := range p.projectCmdOutput {
		err := callback(msg.Line)
		if err != nil {
			p.logger.Err(err.Error())
		}
		return nil
	}

	return nil
}

func (p *DefaultProjectCommandOutputHandler) Handle() {
	for msg := range p.projectCmdOutput {
		p.logger.Info("Recieving message %s", msg.Line)
		if msg.ClearBuffBefore {
			p.clearLogLines(msg.ProjectInfo)
		}
		p.writeLogLine(msg.ProjectInfo, msg.Line)
		if msg.ClearBuffAfter {
			p.clearLogLines(msg.ProjectInfo)
		}
	}
}

func (p *DefaultProjectCommandOutputHandler) clearLogLines(pull string) {
	p.controllerBufferLock.Lock()
	delete(p.projectOutputBuffers, pull)
	p.controllerBufferLock.Unlock()
}

//Add log line to buffer and send to all current channels
func (p *DefaultProjectCommandOutputHandler) writeLogLine(pull string, line string) {
	p.controllerBufferLock.Lock()
	if p.projectOutputBuffers == nil {
		p.projectOutputBuffers = map[string][]string{}
	}
	p.logger.Info("Project info: %s, content: %s", pull, line)

	for ch := range p.controllerBuffers[pull] {
		select {
		case ch <- line:
		default:
			delete(p.controllerBuffers[pull], ch)
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
	delete(p.controllerBuffers[pull], ch)
	p.controllerBufferLock.Unlock()
}
