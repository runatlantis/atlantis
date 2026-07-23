package events

import (
	"sync"

	"github.com/runatlantis/atlantis/server/events/models"
)

type autoplanRunKey struct {
	host       models.VCSHostType
	hostname   string
	repository string
	pullNum    int
}

type autoplanRun struct {
	request autoplanRequest
	pending *autoplanRequest
}

type autoplanRequest struct {
	baseRepo models.Repo
	headRepo models.Repo
	pull     models.PullRequest
	user     models.User
}

type AutoplanRunCoordinator struct {
	mutex sync.Mutex
	runs  map[autoplanRunKey]autoplanRun
}

func NewAutoplanRunCoordinator() *AutoplanRunCoordinator {
	return &AutoplanRunCoordinator{runs: make(map[autoplanRunKey]autoplanRun)}
}

func (c *AutoplanRunCoordinator) start(request autoplanRequest) (autoplanRequest, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := autoplanKey(request)
	if c.runs == nil {
		c.runs = make(map[autoplanRunKey]autoplanRun)
	}
	if run, exists := c.runs[key]; exists {
		if c.matchesExistingRun(run, request) {
			return autoplanRequest{}, false
		}
		run.pending = &request
		c.runs[key] = run
		return autoplanRequest{}, false
	}
	c.runs[key] = autoplanRun{request: request}
	return request, true
}

func (c *AutoplanRunCoordinator) complete(request autoplanRequest) (autoplanRequest, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := autoplanKey(request)
	run, exists := c.runs[key]
	if !exists || run.request.pull.HeadCommit != request.pull.HeadCommit {
		return autoplanRequest{}, false
	}
	if run.pending == nil {
		delete(c.runs, key)
		return autoplanRequest{}, false
	}
	next := *run.pending
	run.request = next
	run.pending = nil
	c.runs[key] = run
	return next, true
}

func (c *AutoplanRunCoordinator) matchesExistingRun(run autoplanRun, request autoplanRequest) bool {
	if run.request.pull.HeadCommit == request.pull.HeadCommit {
		return true
	}
	return run.pending != nil && run.pending.pull.HeadCommit == request.pull.HeadCommit
}

func autoplanKey(request autoplanRequest) autoplanRunKey {
	return autoplanRunKey{
		host:       request.baseRepo.VCSHost.Type,
		hostname:   request.baseRepo.VCSHost.Hostname,
		repository: request.baseRepo.FullName,
		pullNum:    request.pull.Num,
	}
}
