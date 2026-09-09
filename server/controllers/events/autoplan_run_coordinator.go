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
	request       autoplanRequest
	pendingEvents uint64
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
		if run.request.pull.HeadCommit != request.pull.HeadCommit {
			run.pendingEvents++
			c.runs[key] = run
		}
		return autoplanRequest{}, false
	}
	c.runs[key] = autoplanRun{request: request}
	return request, true
}

func (c *AutoplanRunCoordinator) pendingEvents(request autoplanRequest) uint64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := autoplanKey(request)
	run, exists := c.runs[key]
	if !exists || run.request.pull.HeadCommit != request.pull.HeadCommit {
		return 0
	}
	return run.pendingEvents
}

func (c *AutoplanRunCoordinator) scheduleNext(request autoplanRequest, observedEvents uint64, livePull models.PullRequest) (autoplanRequest, bool, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := autoplanKey(request)
	run, exists := c.runs[key]
	if !exists || run.request.pull.HeadCommit != request.pull.HeadCommit {
		return autoplanRequest{}, false, false
	}
	if livePull.HeadCommit == "" {
		return autoplanRequest{}, false, false
	}
	if livePull.HeadCommit == request.pull.HeadCommit {
		if run.pendingEvents > observedEvents {
			run.pendingEvents -= observedEvents
			c.runs[key] = run
			return autoplanRequest{}, false, true
		}
		return autoplanRequest{}, false, false
	}
	next := request
	next.pull.HeadCommit = livePull.HeadCommit
	next.pull.BaseBranch = livePull.BaseBranch
	run.request = next
	run.pendingEvents -= observedEvents
	c.runs[key] = run
	return next, true, false
}

func (c *AutoplanRunCoordinator) release(request autoplanRequest) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := autoplanKey(request)
	run, exists := c.runs[key]
	if exists && run.request.pull.HeadCommit == request.pull.HeadCommit {
		delete(c.runs, key)
	}
}

func autoplanKey(request autoplanRequest) autoplanRunKey {
	return autoplanRunKey{
		host:       request.baseRepo.VCSHost.Type,
		hostname:   request.baseRepo.VCSHost.Hostname,
		repository: request.baseRepo.FullName,
		pullNum:    request.pull.Num,
	}
}
