package jobs

import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
)

type OutputBuffer struct {
	OperationComplete bool
	Buffer            []string
}

type PullInfo struct {
	PullNum     int
	Repo        string
	ProjectName string
	Workspace   string
}

type JobInfo struct {
	PullInfo
	HeadCommit string
}

type ProjectCmdOutputLine struct {
	JobID   string
	JobInfo JobInfo
	Line    string
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_command_output_handler.go ProjectCommandOutputHandler

type ProjectCommandOutputHandler interface {
	// Send will enqueue the msg and wait for Handle() to receive the message.
	Send(ctx command.ProjectContext, msg string)

	// Listens for msg from channel
	Handle()

	// Register registers a channel and blocks until it is caught up. Callers should call this asynchronously when attempting
	// to read the channel in the same goroutine
	Register(jobID string, receiver chan string)

	// Cleans up resources for a pull
	CleanUp(pullInfo PullInfo)

	// Persists job to storage backend and marks operation complete
	CloseJob(jobID string)
}

// AsyncProjectCommandOutputHandler is a handler to transport terraform client
// outputs to the front end.
type AsyncProjectCommandOutputHandler struct {
	// Main channel that receives output from the terraform client
	projectCmdOutput chan *ProjectCmdOutputLine

	// Storage for jobs
	jobStore JobStore

	// Registry to track active connections for a job
	receiverRegistry ReceiverRegistry

	// Map to track jobs in a pull request
	pullToJobMapping sync.Map
	logger           logging.SimpleLogging
}

func NewAsyncProjectCommandOutputHandler(
	projectCmdOutput chan *ProjectCmdOutputLine,
	logger logging.SimpleLogging,
	jobStore JobStore,
) ProjectCommandOutputHandler {
	return &AsyncProjectCommandOutputHandler{
		projectCmdOutput: projectCmdOutput,
		logger:           logger,
		pullToJobMapping: sync.Map{},
		jobStore:         jobStore,
		receiverRegistry: NewReceiverRegistry(),
	}
}

func (p *AsyncProjectCommandOutputHandler) Send(ctx command.ProjectContext, msg string) {
	p.projectCmdOutput <- &ProjectCmdOutputLine{
		JobID: ctx.JobID,
		JobInfo: JobInfo{
			HeadCommit: ctx.Pull.HeadCommit,
			PullInfo: PullInfo{
				PullNum:     ctx.Pull.Num,
				Repo:        ctx.BaseRepo.Name,
				ProjectName: ctx.ProjectName,
				Workspace:   ctx.Workspace,
			},
		},
		Line: msg,
	}
}

func (p *AsyncProjectCommandOutputHandler) Handle() {
	for msg := range p.projectCmdOutput {

		// Add job to pullToJob mapping
		if _, ok := p.pullToJobMapping.Load(msg.JobInfo.PullInfo); !ok {
			p.pullToJobMapping.Store(msg.JobInfo.PullInfo, map[string]bool{})
		}
		value, _ := p.pullToJobMapping.Load(msg.JobInfo.PullInfo)
		jobMapping := value.(map[string]bool)
		jobMapping[msg.JobID] = true

		// Write logs to all active connections
		for ch := range p.receiverRegistry.GetReceivers(msg.JobID) {
			select {
			case ch <- msg.Line:
			default:
				p.receiverRegistry.RemoveReceiver(msg.JobID, ch)
			}
		}

		// Append new log to the output buffer for the job
		p.jobStore.AppendOutput(msg.JobID, msg.Line)
	}
}

func (p *AsyncProjectCommandOutputHandler) Register(jobID string, connection chan string) {
	job, err := p.jobStore.Get(jobID)
	if err != nil {
		p.logger.Err(fmt.Sprintf("getting job: %s", jobID), err)
		return
	}

	// Back fill contents from the output buffer
	for _, line := range job.Output {
		connection <- line
	}

	// Close connection if job is complete
	if job.Status == Complete {
		close(connection)
		return
	}

	// add receiver to registry after backfilling contents from the buffer
	p.receiverRegistry.AddReceiver(jobID, connection)
}

func (p *AsyncProjectCommandOutputHandler) CloseJob(jobID string) {
	// Close active connections and remove receivers from registry
	p.receiverRegistry.CloseAndRemoveReceiversForJob(jobID)

	// Update job status and persist to storage if configured
	if err := p.jobStore.SetJobCompleteStatus(jobID, Complete); err != nil {
		p.logger.Err("updating jobs status to complete", err)
	}
}

func (p *AsyncProjectCommandOutputHandler) CleanUp(pullInfo PullInfo) {
	if value, ok := p.pullToJobMapping.Load(pullInfo); ok {
		jobMapping := value.(map[string]bool)
		for jobID := range jobMapping {
			// Clear output buffer for the job
			p.jobStore.RemoveJob(jobID)

			// Close connections and clear registry for the job
			p.receiverRegistry.CloseAndRemoveReceiversForJob(jobID)
		}

		// Remove pull to job mapping for the job
		p.pullToJobMapping.Delete(pullInfo)
	}
}

// Helper methods for testing
func (p *AsyncProjectCommandOutputHandler) GetReceiverBufferForPull(jobID string) map[chan string]bool {
	return p.receiverRegistry.GetReceivers(jobID)
}

func (p *AsyncProjectCommandOutputHandler) GetJob(jobID string) Job {
	job, _ := p.jobStore.Get(jobID)
	return job
}

func (p *AsyncProjectCommandOutputHandler) GetJobIdMapForPull(pullInfo PullInfo) map[string]bool {
	if value, ok := p.pullToJobMapping.Load(pullInfo); ok {
		return value.(map[string]bool)
	}
	return nil
}

// NoopProjectOutputHandler is a mock that doesn't do anything
type NoopProjectOutputHandler struct{}

func (p *NoopProjectOutputHandler) Send(ctx command.ProjectContext, msg string) {
}

func (p *NoopProjectOutputHandler) Handle() {
}

func (p *NoopProjectOutputHandler) Register(jobID string, receiver chan string) {}

func (p *NoopProjectOutputHandler) CleanUp(pullInfo PullInfo) {
}

func (p *NoopProjectOutputHandler) CloseJob(jobID string) {
}
