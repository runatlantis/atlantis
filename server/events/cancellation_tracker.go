package events

import (
	"fmt"
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

type CancellationTracker interface {
	CancelPull(pull models.PullRequest)
	IsPullCancelled(pull models.PullRequest) bool
	ClearPull(pull models.PullRequest)
}

type DefaultCancellationTracker struct {
	mutex          sync.RWMutex
	cancelledPulls map[string]struct{}
}

func NewCancellationTracker() *DefaultCancellationTracker {
	return &DefaultCancellationTracker{cancelledPulls: make(map[string]struct{})}
}

func (p *DefaultCancellationTracker) CancelPull(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.cancelledPulls[pullKey(pull)] = struct{}{}
}

func (p *DefaultCancellationTracker) IsPullCancelled(pull models.PullRequest) bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	_, exists := p.cancelledPulls[pullKey(pull)]
	return exists
}

// ClearPull removes a pull from the cancelled set (called when a PR is closed).
func (p *DefaultCancellationTracker) ClearPull(pull models.PullRequest) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	delete(p.cancelledPulls, pullKey(pull))
}

func pullKey(pull models.PullRequest) string {
	return fmt.Sprintf("%s#%d", pull.BaseRepo.FullName, pull.Num)
}
