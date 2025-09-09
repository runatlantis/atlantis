package events

import (
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
)

// RunningProcess represents a running terraform operation.
type RunningProcess struct {
	PID       int
	Command   string
	Pull      models.PullRequest
	Project   string
	StartTime time.Time
	// For logical operations, we can store a cancel channel
	CancelCh chan struct{}
}

// ProcessTracker tracks running terraform processes.
type ProcessTracker interface {
	TrackProcess(pid int, command string, pull models.PullRequest, project string) (*RunningProcess, chan struct{})
	RemoveProcess(pid int)
	GetRunningProcesses(pull models.PullRequest) []RunningProcess
	CancelOperation(pid int) error
	// CancelPull marks an entire pull request as cancelled so that any future
	// operations for the same pull will not start (their cancel channel will
	// already be closed when TrackProcess returns).
	CancelPull(pull models.PullRequest)
	// IsPullCancelled returns true if the given pull has been marked as
	// cancelled.
	IsPullCancelled(pull models.PullRequest) bool
}

// DefaultProcessTracker implements ProcessTracker.
type DefaultProcessTracker struct {
	processes map[int]RunningProcess
	mutex     sync.RWMutex
	nextPID   int
	// cancelledPulls stores pull identifiers that have been cancelled. Key
	// format: <repoFullName>#<pullNum>
	cancelledPulls map[string]struct{}
}

// NewProcessTracker creates a new process tracker.
func NewProcessTracker() *DefaultProcessTracker {
	return &DefaultProcessTracker{
		processes:      make(map[int]RunningProcess),
		nextPID:        1,
		cancelledPulls: make(map[string]struct{}),
	}
}

// TrackProcess adds a process to the tracker and returns a cancel channel.
func (p *DefaultProcessTracker) TrackProcess(pid int, command string, pull models.PullRequest, project string) (*RunningProcess, chan struct{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// If pid is 0, assign a logical PID
	// This allows callers to pass 0 when they don't care about the PID
	// and want the process tracker to assign one.
	if pid == 0 {
		pid = p.nextPID
		p.nextPID++
	}

	cancelCh := make(chan struct{})

	// If the pull has been cancelled previously, close the channel so callers
	// can detect cancellation immediately. We intentionally do NOT add this
	// process to the processes map so that RemoveProcess becomes a no-op and
	// we don't double-close.
	key := pullKey(pull)
	_, pullCancelled := p.cancelledPulls[key]
	if pullCancelled {
		close(cancelCh)
	}

	// Add the process to the map
	process := RunningProcess{
		PID:       pid,
		Command:   command,
		Pull:      pull,
		Project:   project,
		StartTime: time.Now(),
		CancelCh:  cancelCh,
	}
	if !pullCancelled {
		p.processes[pid] = process
	}

	return &process, cancelCh
}

// RemoveProcess removes a process from the tracker.
func (p *DefaultProcessTracker) RemoveProcess(pid int) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if process, exists := p.processes[pid]; exists {
		close(process.CancelCh)
		delete(p.processes, pid)
	}
}

// GetRunningProcesses returns all processes for a given pull request.
func (p *DefaultProcessTracker) GetRunningProcesses(pull models.PullRequest) []RunningProcess {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var result []RunningProcess
	for _, process := range p.processes {
		if process.Pull.Num == pull.Num && process.Pull.BaseRepo.FullName == pull.BaseRepo.FullName {
			result = append(result, process)
		}
	}

	return result
}

// CancelOperation cancels a logical operation by closing its cancel channel.
func (p *DefaultProcessTracker) CancelOperation(pid int) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	process, exists := p.processes[pid]
	if !exists {
		return fmt.Errorf("process with PID %d not found", pid)
	}

	select {
	case <-process.CancelCh:
		// Already cancelled
		return fmt.Errorf("process %d already cancelled", pid)
	default:
		close(process.CancelCh)
		delete(p.processes, pid)
		return nil
	}
}

// CancelPull marks a pull as cancelled so that future operations won't start.
func (p *DefaultProcessTracker) CancelPull(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cancelledPulls[pullKey(pull)] = struct{}{}
}

// IsPullCancelled returns true if the pull is marked cancelled.
func (p *DefaultProcessTracker) IsPullCancelled(pull models.PullRequest) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, exists := p.cancelledPulls[pullKey(pull)]
	return exists
}

// pullKey builds a unique key for a pull request.
func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
