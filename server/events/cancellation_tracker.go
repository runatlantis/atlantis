package events

import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

type CancellationTracker interface {
	CancelPullRequest(pull models.PullRequest)
	IsPullRequestCancelled(pull models.PullRequest) bool
	ClearPullRequest(pull models.PullRequest)
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

// CancelPullRequest marks an entire pull request as cancelled, preventing any future operations
func (p *DefaultCancellationTracker) CancelPullRequest(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	pullKeyStr := pullKey(pull)
	p.cancelledPulls[pullKeyStr] = struct{}{}
}

// IsPullRequestCancelled checks if the entire pull request has been cancelled
func (p *DefaultCancellationTracker) IsPullRequestCancelled(pull models.PullRequest) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, exists := p.cancelledPulls[pullKey(pull)]
	return exists
}

// ClearPullRequest removes cancellation for a pull request (called when a PR is closed)
func (p *DefaultCancellationTracker) ClearPullRequest(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.cancelledPulls, pullKey(pull))
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
