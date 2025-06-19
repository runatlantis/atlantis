package models

import (
	"time"
)

// PlanQueueEntry represents a single entry in the plan queue
type PlanQueueEntry struct {
	// ID is a unique identifier for this queue entry
	ID string
	// Project is the project that this queue entry is for
	Project Project
	// Workspace is the workspace that this queue entry is for
	Workspace string
	// Pull is the pull request that requested the plan
	Pull PullRequest
	// User is the user that requested the plan
	User User
	// Time is when this entry was added to the queue
	Time time.Time
	// Priority is the priority of this entry (lower numbers = higher priority)
	Priority int
	// Command is the command that was requested (usually "plan")
	Command string
}

// PlanQueue represents the entire plan queue for a project/workspace
type PlanQueue struct {
	// Project is the project this queue is for
	Project Project
	// Workspace is the workspace this queue is for
	Workspace string
	// Entries are the queued entries, ordered by priority and time
	Entries []PlanQueueEntry
	// CreatedAt is when this queue was created
	CreatedAt time.Time
	// UpdatedAt is when this queue was last updated
	UpdatedAt time.Time
}

// PlanQueueManager defines the interface for managing plan queues
type PlanQueueManager interface {
	// AddToQueue adds a new entry to the queue
	AddToQueue(entry PlanQueueEntry) error
	
	// RemoveFromQueue removes an entry from the queue
	RemoveFromQueue(project Project, workspace string, pullNum int) error
	
	// GetNextInQueue gets the next entry in the queue
	GetNextInQueue(project Project, workspace string) (*PlanQueueEntry, error)
	
	// IsInQueue checks if a pull request is already in the queue
	IsInQueue(project Project, workspace string, pullNum int) (bool, error)
	
	// GetQueueStatus gets the current queue status for a project/workspace
	GetQueueStatus(project Project, workspace string) (*PlanQueue, error)
	
	// TransferLock transfers the lock to the next person in queue
	TransferLock(project Project, workspace string) error
	
	// CleanupQueue removes all queue entries for a pull request
	CleanupQueue(repoFullName string, pullNum int) error
	
	// NotifyQueueUpdate notifies users about queue updates
	NotifyQueueUpdate(project Project, workspace string, message string) error
} 