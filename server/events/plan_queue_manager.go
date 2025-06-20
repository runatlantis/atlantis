package events

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	"github.com/runatlantis/atlantis/server/logging"
)

// DefaultPlanQueueManager implements the PlanQueueManager interface
type DefaultPlanQueueManager struct {
	Backend   locking.Backend
	Locker    locking.Locker
	VCSClient vcs.Client
	Logger    logging.SimpleLogging

	// In-memory queues for now (could be moved to backend later)
	queues      map[string]*models.PlanQueue
	queuesMutex sync.RWMutex
}

// NewDefaultPlanQueueManager creates a new plan queue manager
func NewDefaultPlanQueueManager(backend locking.Backend, locker locking.Locker, vcsClient vcs.Client, logger logging.SimpleLogging) *DefaultPlanQueueManager {
	return &DefaultPlanQueueManager{
		Backend:   backend,
		Locker:    locker,
		VCSClient: vcsClient,
		Logger:    logger,
		queues:    make(map[string]*models.PlanQueue),
	}
}

// AddToQueue adds a new entry to the queue
func (p *DefaultPlanQueueManager) AddToQueue(entry models.PlanQueueEntry) error {
	p.queuesMutex.Lock()
	defer p.queuesMutex.Unlock()

	queueKey := p.queueKey(entry.Project, entry.Workspace)

	// Get or create queue
	queue, exists := p.queues[queueKey]
	if !exists {
		queue = &models.PlanQueue{
			Project:   entry.Project,
			Workspace: entry.Workspace,
			Entries:   []models.PlanQueueEntry{},
			CreatedAt: time.Now(),
		}
		p.queues[queueKey] = queue
	}

	// Check if entry already exists
	for _, existingEntry := range queue.Entries {
		if existingEntry.Pull.Num == entry.Pull.Num {
			p.Logger.Debug("Entry already exists in queue for PR %d", entry.Pull.Num)
			return nil
		}
	}

	// Add entry and sort by priority and time
	queue.Entries = append(queue.Entries, entry)
	sort.Slice(queue.Entries, func(i, j int) bool {
		if queue.Entries[i].Priority != queue.Entries[j].Priority {
			return queue.Entries[i].Priority < queue.Entries[j].Priority
		}
		return queue.Entries[i].Time.Before(queue.Entries[j].Time)
	})

	queue.UpdatedAt = time.Now()

	p.Logger.Info("Added PR %d to queue for project %s, workspace %s (position: %d)",
		entry.Pull.Num, entry.Project.String(), entry.Workspace, len(queue.Entries))

	// Notify the user about their position in queue
	go p.notifyQueuePosition(entry, len(queue.Entries))

	return nil
}

// RemoveFromQueue removes an entry from the queue
func (p *DefaultPlanQueueManager) RemoveFromQueue(project models.Project, workspace string, pullNum int) error {
	p.queuesMutex.Lock()
	defer p.queuesMutex.Unlock()

	queueKey := p.queueKey(project, workspace)
	queue, exists := p.queues[queueKey]
	if !exists {
		return nil
	}

	// Find and remove the entry
	for i, entry := range queue.Entries {
		if entry.Pull.Num == pullNum {
			queue.Entries = append(queue.Entries[:i], queue.Entries[i+1:]...)
			queue.UpdatedAt = time.Now()

			p.Logger.Info("Removed PR %d from queue for project %s, workspace %s",
				pullNum, project.String(), workspace)

			// If queue is empty, remove it
			if len(queue.Entries) == 0 {
				delete(p.queues, queueKey)
			}

			return nil
		}
	}

	return nil
}

// GetNextInQueue gets the next entry in the queue
func (p *DefaultPlanQueueManager) GetNextInQueue(project models.Project, workspace string) (*models.PlanQueueEntry, error) {
	p.queuesMutex.RLock()
	defer p.queuesMutex.RUnlock()

	queueKey := p.queueKey(project, workspace)
	queue, exists := p.queues[queueKey]
	if !exists || len(queue.Entries) == 0 {
		return nil, nil
	}

	nextEntry := queue.Entries[0]
	return &nextEntry, nil
}

// IsInQueue checks if a pull request is already in the queue
func (p *DefaultPlanQueueManager) IsInQueue(project models.Project, workspace string, pullNum int) (bool, error) {
	p.queuesMutex.RLock()
	defer p.queuesMutex.RUnlock()

	queueKey := p.queueKey(project, workspace)
	queue, exists := p.queues[queueKey]
	if !exists {
		return false, nil
	}

	for _, entry := range queue.Entries {
		if entry.Pull.Num == pullNum {
			return true, nil
		}
	}

	return false, nil
}

// GetQueueStatus gets the current queue status for a project/workspace
func (p *DefaultPlanQueueManager) GetQueueStatus(project models.Project, workspace string) (*models.PlanQueue, error) {
	p.queuesMutex.RLock()
	defer p.queuesMutex.RUnlock()

	queueKey := p.queueKey(project, workspace)
	queue, exists := p.queues[queueKey]
	if !exists {
		return nil, nil
	}

	// Return a copy to avoid race conditions
	queueCopy := *queue
	queueCopy.Entries = make([]models.PlanQueueEntry, len(queue.Entries))
	copy(queueCopy.Entries, queue.Entries)

	return &queueCopy, nil
}

// TransferLock transfers the lock to the next person in queue
func (p *DefaultPlanQueueManager) TransferLock(project models.Project, workspace string) error {
	p.queuesMutex.Lock()
	defer p.queuesMutex.Unlock()

	queueKey := p.queueKey(project, workspace)
	queue, exists := p.queues[queueKey]
	if !exists || len(queue.Entries) == 0 {
		return nil
	}

	// Get the next entry
	nextEntry := queue.Entries[0]

	// Remove from queue
	queue.Entries = queue.Entries[1:]
	queue.UpdatedAt = time.Now()

	// If queue is empty, remove it
	if len(queue.Entries) == 0 {
		delete(p.queues, queueKey)
	}

	// Try to acquire the lock for the next person
	lockAttempt, err := p.Locker.TryLock(project, workspace, nextEntry.Pull, nextEntry.User)
	if err != nil {
		p.Logger.Warn("Failed to transfer lock to next person in queue: %s", err)
		// Put the entry back at the front of the queue
		queue.Entries = append([]models.PlanQueueEntry{nextEntry}, queue.Entries...)
		return fmt.Errorf("failed to transfer lock: %w", err)
	}

	if !lockAttempt.LockAcquired {
		p.Logger.Warn("Failed to transfer lock to next person in queue")
		// Put the entry back at the front of the queue
		queue.Entries = append([]models.PlanQueueEntry{nextEntry}, queue.Entries...)
		return fmt.Errorf("failed to transfer lock")
	}

	p.Logger.Info("Successfully transferred lock to PR %d for project %s, workspace %s",
		nextEntry.Pull.Num, project.String(), workspace)

	// Notify the user that they now have the lock
	go p.notifyLockAcquired(nextEntry)

	return nil
}

// CleanupQueue removes all queue entries for a pull request
func (p *DefaultPlanQueueManager) CleanupQueue(repoFullName string, pullNum int) error {
	p.queuesMutex.Lock()
	defer p.queuesMutex.Unlock()

	for queueKey, queue := range p.queues {
		if queue.Project.RepoFullName == repoFullName {
			// Remove entries for this pull request
			newEntries := []models.PlanQueueEntry{}
			for _, entry := range queue.Entries {
				if entry.Pull.Num != pullNum {
					newEntries = append(newEntries, entry)
				}
			}

			if len(newEntries) != len(queue.Entries) {
				queue.Entries = newEntries
				queue.UpdatedAt = time.Now()

				p.Logger.Info("Cleaned up queue entries for PR %d in project %s, workspace %s",
					pullNum, queue.Project.String(), queue.Workspace)

				// If queue is empty, remove it
				if len(queue.Entries) == 0 {
					delete(p.queues, queueKey)
				}
			}
		}
	}

	return nil
}

// NotifyQueueUpdate notifies users about queue updates
func (p *DefaultPlanQueueManager) NotifyQueueUpdate(project models.Project, workspace string, message string) error {
	queue, err := p.GetQueueStatus(project, workspace)
	if err != nil {
		return err
	}

	if queue == nil {
		return nil
	}

	// Notify all users in the queue
	for _, entry := range queue.Entries {
		go func(e models.PlanQueueEntry) {
			if err := p.notifyUser(e.Pull, message); err != nil {
				p.Logger.Warn("Failed to notify user about queue update: %s", err)
			}
		}(entry)
	}

	return nil
}

// notifyQueuePosition notifies a user about their position in the queue
func (p *DefaultPlanQueueManager) notifyQueuePosition(entry models.PlanQueueEntry, position int) {
	message := fmt.Sprintf(
		"Your plan request has been added to the queue for project `%s` in workspace `%s`. "+
			"You are currently in position %d. You will be notified when it's your turn to plan.",
		entry.Project.String(), entry.Workspace, position)

	if err := p.notifyUser(entry.Pull, message); err != nil {
		p.Logger.Warn("Failed to notify user about queue position: %s", err)
	}
}

// notifyLockAcquired notifies a user that they have acquired the lock
func (p *DefaultPlanQueueManager) notifyLockAcquired(entry models.PlanQueueEntry) {
	message := fmt.Sprintf(
		"🎉 Your turn! The lock for project `%s` in workspace `%s` has been transferred to you. "+
			"You can now run `atlantis plan` to start planning.",
		entry.Project.String(), entry.Workspace)

	if err := p.notifyUser(entry.Pull, message); err != nil {
		p.Logger.Warn("Failed to notify user about lock acquisition: %s", err)
	}
}

// notifyUser sends a notification to a user
func (p *DefaultPlanQueueManager) notifyUser(pull models.PullRequest, message string) error {
	// For now, we'll just log the notification
	// In a real implementation, this would send a comment to the PR
	p.Logger.Info("Notification for PR %d: %s", pull.Num, message)
	return nil
}

// queueKey generates a key for the queue
func (p *DefaultPlanQueueManager) queueKey(project models.Project, workspace string) string {
	return fmt.Sprintf("queue:%s:%s:%s", project.RepoFullName, project.Path, workspace)
}

// GetAllQueues gets all active queues
func (p *DefaultPlanQueueManager) GetAllQueues() ([]*models.PlanQueue, error) {
	p.queuesMutex.RLock()
	defer p.queuesMutex.RUnlock()

	queues := make([]*models.PlanQueue, 0, len(p.queues))
	for _, queue := range p.queues {
		// Return a copy to avoid race conditions
		queueCopy := *queue
		queueCopy.Entries = make([]models.PlanQueueEntry, len(queue.Entries))
		copy(queueCopy.Entries, queue.Entries)
		queues = append(queues, &queueCopy)
	}

	return queues, nil
}
