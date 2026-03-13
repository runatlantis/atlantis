package events

//go:generate mockgen -package mocks -destination mocks/mock_cancellation_tracker.go . CancellationTracker
import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

type CancellationTracker interface {
	Cancel(pull models.PullRequest)
	IsCancelled(pull models.PullRequest) bool
	Clear(pull models.PullRequest)
}

type DefaultCancellationTracker struct {
	mutex          sync.RWMutex
	cancelledPulls map[string]struct{}
}

func NewCancellationTracker() *DefaultCancellationTracker {
	return &DefaultCancellationTracker{
		cancelledPulls: make(map[string]struct{}),
	}
}

// Cancel marks an entire pull request as cancelled, preventing any future operations
func (p *DefaultCancellationTracker) Cancel(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pullKeyStr := pullKey(pull)
	p.cancelledPulls[pullKeyStr] = struct{}{}
}

// IsCancelled checks if the entire pull request has been cancelled
func (p *DefaultCancellationTracker) IsCancelled(pull models.PullRequest) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, exists := p.cancelledPulls[pullKey(pull)]
	return exists
}

// Clear removes cancellation for a pull request (called when a PR is closed)
func (p *DefaultCancellationTracker) Clear(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.cancelledPulls, pullKey(pull))
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
