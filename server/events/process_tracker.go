package events

import (
	"fmt"
	"os"
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
	GetAllRunningProcesses() []RunningProcess
	KillProcess(pid int) error
	CancelOperation(pid int) error
}

// DefaultProcessTracker implements ProcessTracker.
type DefaultProcessTracker struct {
	processes map[int]RunningProcess
	mutex     sync.RWMutex
	nextPID   int
}

// NewProcessTracker creates a new process tracker.
func NewProcessTracker() *DefaultProcessTracker {
	return &DefaultProcessTracker{
		processes: make(map[int]RunningProcess),
		nextPID:   1,
	}
}

// TrackProcess adds a process to the tracker and returns a cancel channel.
func (p *DefaultProcessTracker) TrackProcess(pid int, command string, pull models.PullRequest, project string) (*RunningProcess, chan struct{}) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// If pid is 0, assign a logical PID
	if pid == 0 {
		pid = p.nextPID
		p.nextPID++
	}

	cancelCh := make(chan struct{})
	process := RunningProcess{
		PID:       pid,
		Command:   command,
		Pull:      pull,
		Project:   project,
		StartTime: time.Now(),
		CancelCh:  cancelCh,
	}

	p.processes[pid] = process
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

// GetAllRunningProcesses returns all running processes.
func (p *DefaultProcessTracker) GetAllRunningProcesses() []RunningProcess {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var result []RunningProcess
	for _, process := range p.processes {
		result = append(result, process)
	}

	return result
}

// KillProcess attempts to kill a process by PID.
func (p *DefaultProcessTracker) KillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	return process.Kill()
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
