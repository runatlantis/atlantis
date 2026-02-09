// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package jobs

import (
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

type OutputBuffer struct {
	OperationComplete bool
	Buffer            []string
}

type PullInfo struct {
	PullNum      int
	Repo         string
	RepoFullName string
	ProjectName  string
	Path         string
	Workspace    string
}

type JobIDInfo struct {
	JobID          string
	JobIDUrl       string
	JobDescription string
	Time           time.Time
	TimeFormatted  string
	JobStep        string
	CompletedAt    time.Time // Zero value if still running
	TriggeredBy    string    // Username who triggered the job
}

type PullInfoWithJobIDs struct {
	Pull       PullInfo
	JobIDInfos []JobIDInfo
}

type JobInfo struct {
	PullInfo
	HeadCommit     string
	JobDescription string
	JobStep        string
	TriggeredBy    string // Username who triggered the job
}

type ProjectCmdOutputLine struct {
	JobID             string
	JobInfo           JobInfo
	Line              string
	OperationComplete bool
}

// AsyncProjectCommandOutputHandler is a handler to transport terraform client
// outputs to the front end.
type AsyncProjectCommandOutputHandler struct {
	projectCmdOutput chan *ProjectCmdOutputLine

	projectOutputBuffers     map[string]OutputBuffer
	projectOutputBuffersLock sync.RWMutex

	receiverBuffers     map[string]map[chan string]bool
	receiverBuffersLock sync.RWMutex

	logger logging.SimpleLogging

	// Tracks all the jobs for a pull request which is used for clean up after a pull request is closed.
	pullToJobMapping sync.Map
}

//go:generate pegomock generate --package mocks -o mocks/mock_project_command_output_handler.go ProjectCommandOutputHandler

type ProjectCommandOutputHandler interface {
	// Send will enqueue the msg and wait for Handle() to receive the message.
	Send(ctx command.ProjectContext, msg string, operationComplete bool)

	SendWorkflowHook(ctx models.WorkflowHookCommandContext, msg string, operationComplete bool)

	// Register registers a channel and blocks until it is caught up. Callers should call this asynchronously when attempting
	// to read the channel in the same goroutine
	Register(jobID string, receiver chan string)

	// Deregister removes a channel from successive updates and closes it.
	Deregister(jobID string, receiver chan string)

	IsKeyExists(key string) bool

	// Listens for msg from channel
	Handle()

	// Cleans up resources for a pull
	CleanUp(pullInfo PullInfo)

	// Returns a map from Pull Requests to Jobs
	GetPullToJobMapping() []PullInfoWithJobIDs

	// GetProjectOutputBuffer returns the output buffer for a job (for persistence)
	GetProjectOutputBuffer(jobID string) OutputBuffer
}

func NewAsyncProjectCommandOutputHandler(
	projectCmdOutput chan *ProjectCmdOutputLine,
	logger logging.SimpleLogging,
) ProjectCommandOutputHandler {
	return &AsyncProjectCommandOutputHandler{
		projectCmdOutput:     projectCmdOutput,
		logger:               logger,
		receiverBuffers:      map[string]map[chan string]bool{},
		projectOutputBuffers: map[string]OutputBuffer{},
		pullToJobMapping:     sync.Map{},
	}
}

func (p *AsyncProjectCommandOutputHandler) GetPullToJobMapping() []PullInfoWithJobIDs {
	var pullToJobMappings []PullInfoWithJobIDs

	p.pullToJobMapping.Range(func(key, value any) bool {
		pullInfo := key.(PullInfo)
		jobIDSyncMap := value.(*sync.Map)

		var jobIDInfos []JobIDInfo
		jobIDSyncMap.Range(func(_, v any) bool {
			jobIDInfos = append(jobIDInfos, v.(JobIDInfo))
			return true
		})

		pullToJobMappings = append(pullToJobMappings, PullInfoWithJobIDs{
			Pull:       pullInfo,
			JobIDInfos: jobIDInfos,
		})
		return true
	})

	return pullToJobMappings
}

func (p *AsyncProjectCommandOutputHandler) IsKeyExists(key string) bool {
	p.projectOutputBuffersLock.RLock()
	defer p.projectOutputBuffersLock.RUnlock()
	_, ok := p.projectOutputBuffers[key]
	return ok
}

func (p *AsyncProjectCommandOutputHandler) Send(ctx command.ProjectContext, msg string, operationComplete bool) {
	p.projectCmdOutput <- &ProjectCmdOutputLine{
		JobID: ctx.JobID,
		JobInfo: JobInfo{
			HeadCommit: ctx.Pull.HeadCommit,
			PullInfo: PullInfo{
				PullNum:      ctx.Pull.Num,
				Repo:         ctx.BaseRepo.Name,
				RepoFullName: ctx.BaseRepo.FullName,
				ProjectName:  ctx.ProjectName,
				Path:         ctx.RepoRelDir,
				Workspace:    ctx.Workspace,
			},
			JobStep:     ctx.CommandName.String(),
			TriggeredBy: ctx.User.Username,
		},
		Line:              msg,
		OperationComplete: operationComplete,
	}
}

func (p *AsyncProjectCommandOutputHandler) SendWorkflowHook(ctx models.WorkflowHookCommandContext, msg string, operationComplete bool) {
	p.projectCmdOutput <- &ProjectCmdOutputLine{
		JobID: ctx.HookID,
		JobInfo: JobInfo{
			HeadCommit: ctx.Pull.HeadCommit,
			PullInfo: PullInfo{
				PullNum:      ctx.Pull.Num,
				Repo:         ctx.BaseRepo.Name,
				RepoFullName: ctx.BaseRepo.FullName,
			},
			JobDescription: ctx.HookDescription,
			JobStep:        ctx.HookStepName,
		},
		Line:              msg,
		OperationComplete: operationComplete,
	}
}

func (p *AsyncProjectCommandOutputHandler) Register(jobID string, receiver chan string) {
	p.addChan(receiver, jobID)
}

func (p *AsyncProjectCommandOutputHandler) Handle() {
	for msg := range p.projectCmdOutput {
		if msg.OperationComplete {
			p.completeJob(msg.JobID)
			continue
		}

		// Add job to pullToJob mapping
		if _, ok := p.pullToJobMapping.Load(msg.JobInfo.PullInfo); !ok {
			p.pullToJobMapping.Store(msg.JobInfo.PullInfo, &sync.Map{})
		}
		value, _ := p.pullToJobMapping.Load(msg.JobInfo.PullInfo)
		jobMapping := value.(*sync.Map)
		jobMapping.Store(msg.JobID, JobIDInfo{
			JobID:          msg.JobID,
			JobIDUrl:       "/jobs/" + msg.JobID,
			JobDescription: msg.JobInfo.JobDescription,
			Time:           time.Now(),
			JobStep:        msg.JobInfo.JobStep,
			TriggeredBy:    msg.JobInfo.TriggeredBy,
		})

		// Forward new message to all receiver channels and output buffer
		p.writeLogLine(msg.JobID, msg.Line)
	}
}

func (p *AsyncProjectCommandOutputHandler) completeJob(jobID string) {
	p.projectOutputBuffersLock.Lock()
	p.receiverBuffersLock.Lock()
	defer func() {
		p.projectOutputBuffersLock.Unlock()
		p.receiverBuffersLock.Unlock()
	}()

	// Update operation status to complete
	if outputBuffer, ok := p.projectOutputBuffers[jobID]; ok {
		outputBuffer.OperationComplete = true
		p.projectOutputBuffers[jobID] = outputBuffer
	}

	// Update completion time in job mapping
	p.setJobCompletionTime(jobID)

	// Close active receiver channels
	if openChannels, ok := p.receiverBuffers[jobID]; ok {
		for ch := range openChannels {
			close(ch)
		}
	}

}

func (p *AsyncProjectCommandOutputHandler) addChan(ch chan string, jobID string) {
	p.projectOutputBuffersLock.RLock()
	outputBuffer := p.projectOutputBuffers[jobID]
	p.projectOutputBuffersLock.RUnlock()

	for _, line := range outputBuffer.Buffer {
		ch <- line
	}

	// No need register receiver since all the logs have been streamed
	if outputBuffer.OperationComplete {
		close(ch)
		return
	}

	// add the channel to our registry after we backfill the contents of the buffer,
	// to prevent new messages coming in interleaving with this backfill.
	p.receiverBuffersLock.Lock()
	if p.receiverBuffers[jobID] == nil {
		p.receiverBuffers[jobID] = map[chan string]bool{}
	}
	p.receiverBuffers[jobID][ch] = true
	p.receiverBuffersLock.Unlock()
}

// Add log line to buffer and send to all current channels
func (p *AsyncProjectCommandOutputHandler) writeLogLine(jobID string, line string) {
	p.receiverBuffersLock.Lock()
	for ch := range p.receiverBuffers[jobID] {
		select {
		case ch <- line:
		default:
			// Close channel gracefully instead of silent delete
			// This allows SSE to auto-reconnect
			close(ch)
			delete(p.receiverBuffers[jobID], ch)
			p.logger.Warn("Buffer full, closing connection", map[string]any{
				"jobID": jobID,
			})
		}
	}
	p.receiverBuffersLock.Unlock()

	p.projectOutputBuffersLock.Lock()
	if _, ok := p.projectOutputBuffers[jobID]; !ok {
		p.projectOutputBuffers[jobID] = OutputBuffer{
			Buffer: []string{},
		}
	}
	outputBuffer := p.projectOutputBuffers[jobID]
	outputBuffer.Buffer = append(outputBuffer.Buffer, line)
	p.projectOutputBuffers[jobID] = outputBuffer

	p.projectOutputBuffersLock.Unlock()
}

// Remove channel, so client no longer receives Terraform output
func (p *AsyncProjectCommandOutputHandler) Deregister(jobID string, ch chan string) {
	p.logger.Debug("Removing channel for %s", jobID)
	p.receiverBuffersLock.Lock()
	delete(p.receiverBuffers[jobID], ch)
	p.receiverBuffersLock.Unlock()
}

func (p *AsyncProjectCommandOutputHandler) GetReceiverBufferForPull(jobID string) map[chan string]bool {
	p.receiverBuffersLock.RLock()
	defer p.receiverBuffersLock.RUnlock()
	return p.receiverBuffers[jobID]
}

func (p *AsyncProjectCommandOutputHandler) GetProjectOutputBuffer(jobID string) OutputBuffer {
	p.projectOutputBuffersLock.RLock()
	defer p.projectOutputBuffersLock.RUnlock()
	return p.projectOutputBuffers[jobID]
}

func (p *AsyncProjectCommandOutputHandler) GetJobIDMapForPull(pullInfo PullInfo) map[string]JobIDInfo {
	result := make(map[string]JobIDInfo)
	if value, ok := p.pullToJobMapping.Load(pullInfo); ok {
		jobIDSyncMap := value.(*sync.Map)
		jobIDSyncMap.Range(func(k, v any) bool {
			result[k.(string)] = v.(JobIDInfo)
			return true
		})
		return result
	}
	return nil
}

func (p *AsyncProjectCommandOutputHandler) CleanUp(pullInfo PullInfo) {
	if value, ok := p.pullToJobMapping.Load(pullInfo); ok {
		jobIDSyncMap := value.(*sync.Map)
		jobIDSyncMap.Range(func(k, _ any) bool {
			jobID := k.(string)
			p.projectOutputBuffersLock.Lock()
			delete(p.projectOutputBuffers, jobID)
			p.projectOutputBuffersLock.Unlock()

			p.receiverBuffersLock.Lock()
			delete(p.receiverBuffers, jobID)
			p.receiverBuffersLock.Unlock()
			return true
		})
		// Remove job mapping
		p.pullToJobMapping.Delete(pullInfo)
	}
}

// InitTestJob initializes an empty buffer for a test job (for testing)
func (p *AsyncProjectCommandOutputHandler) InitTestJob(jobID string) {
	p.projectOutputBuffersLock.Lock()
	if _, ok := p.projectOutputBuffers[jobID]; !ok {
		p.projectOutputBuffers[jobID] = OutputBuffer{
			Buffer: []string{},
		}
	}
	p.projectOutputBuffersLock.Unlock()
}

// SendTestLine sends a line directly to a job's buffer (for testing)
func (p *AsyncProjectCommandOutputHandler) SendTestLine(jobID string, line string) {
	p.InitTestJob(jobID)
	p.writeLogLine(jobID, line)
}

// MarkComplete marks a job as complete (for testing)
func (p *AsyncProjectCommandOutputHandler) MarkComplete(jobID string) {
	p.projectOutputBuffersLock.Lock()
	if outputBuffer, ok := p.projectOutputBuffers[jobID]; ok {
		outputBuffer.OperationComplete = true
		p.projectOutputBuffers[jobID] = outputBuffer
	}
	p.projectOutputBuffersLock.Unlock()

	// Update completion time in job mapping
	p.setJobCompletionTime(jobID)

	// Close active receiver channels
	p.receiverBuffersLock.Lock()
	if openChannels, ok := p.receiverBuffers[jobID]; ok {
		for ch := range openChannels {
			close(ch)
		}
		delete(p.receiverBuffers, jobID)
	}
	p.receiverBuffersLock.Unlock()
}

// setJobCompletionTime updates the CompletedAt field for a job in the pull mapping
func (p *AsyncProjectCommandOutputHandler) setJobCompletionTime(jobID string) {
	completionTime := time.Now()
	p.pullToJobMapping.Range(func(key, value any) bool {
		jobIDSyncMap := value.(*sync.Map)
		if jobInfoValue, ok := jobIDSyncMap.Load(jobID); ok {
			jobInfo := jobInfoValue.(JobIDInfo)
			jobInfo.CompletedAt = completionTime
			jobIDSyncMap.Store(jobID, jobInfo)
			return false // Found, stop iterating
		}
		return true // Continue looking
	})
}

// RegisterTestJob initializes a test job with PR association for dev mode testing.
// This makes the job appear in the jobs list page.
func (p *AsyncProjectCommandOutputHandler) RegisterTestJob(jobID string, pullInfo PullInfo, jobStep string) {
	// Initialize output buffer
	p.projectOutputBuffersLock.Lock()
	if _, ok := p.projectOutputBuffers[jobID]; !ok {
		p.projectOutputBuffers[jobID] = OutputBuffer{
			Buffer: []string{},
		}
	}
	p.projectOutputBuffersLock.Unlock()

	// Register in pullToJobMapping so it appears on jobs list
	if _, ok := p.pullToJobMapping.Load(pullInfo); !ok {
		p.pullToJobMapping.Store(pullInfo, &sync.Map{})
	}
	value, _ := p.pullToJobMapping.Load(pullInfo)
	jobMapping := value.(*sync.Map)
	jobMapping.Store(jobID, JobIDInfo{
		JobID:          jobID,
		JobIDUrl:       "/jobs/" + jobID,
		JobDescription: "Test job",
		Time:           time.Now(),
		TimeFormatted:  time.Now().Format("Jan 2, 2006 3:04 PM"),
		JobStep:        jobStep,
	})
}

// NoopProjectOutputHandler is a mock that doesn't do anything
type NoopProjectOutputHandler struct{}

func (p *NoopProjectOutputHandler) Send(_ command.ProjectContext, _ string, _ bool) {
}

func (p *NoopProjectOutputHandler) SendWorkflowHook(_ models.WorkflowHookCommandContext, _ string, _ bool) {
}

func (p *NoopProjectOutputHandler) Register(_ string, _ chan string) {}

func (p *NoopProjectOutputHandler) Deregister(_ string, _ chan string) {}

func (p *NoopProjectOutputHandler) Handle() {
}

func (p *NoopProjectOutputHandler) CleanUp(_ PullInfo) {
}

func (p *NoopProjectOutputHandler) IsKeyExists(_ string) bool {
	return false
}

func (p *NoopProjectOutputHandler) GetPullToJobMapping() []PullInfoWithJobIDs {
	return []PullInfoWithJobIDs{}
}

func (p *NoopProjectOutputHandler) GetProjectOutputBuffer(_ string) OutputBuffer {
	return OutputBuffer{}
}
