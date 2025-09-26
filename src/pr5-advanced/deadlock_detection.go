// PR #5: Advanced Features & Monitoring - Deadlock Detection and Observability
// This file implements advanced deadlock detection and comprehensive monitoring

package advanced

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/models"
	"go.uber.org/zap"
)

// DeadlockDetector monitors and prevents deadlock situations
type DeadlockDetector struct {
	mu sync.RWMutex
	
	// Lock dependency graph
	dependencyGraph *LockGraph
	lockRegistry *LockRegistry
	
	// Detection configuration
	config DetectionConfig
	
	// Monitoring and alerting
	metrics *DeadlockMetrics
	alertManager *AlertManager
	logger *zap.Logger
	
	// Control channels
	stopCh chan struct{}
	done chan struct{}
}

// DetectionConfig controls deadlock detection behavior
type DetectionConfig struct {
	// Detection intervals
	ScanInterval time.Duration
	DeepScanInterval time.Duration
	
	// Thresholds
	MaxWaitTime time.Duration
	MaxChainLength int
	SuspiciousPatternThreshold int
	
	// Actions
	AutoResolve bool
	PreventiveAbort bool
	EnableAlerts bool
	
	// Performance tuning
	MaxConcurrentScans int
	GraphPruningInterval time.Duration
}

// LockGraph represents the dependency graph of locks
type LockGraph struct {
	mu sync.RWMutex
	nodes map[string]*LockNode
	edges map[string]map[string]*Dependency
}

// LockNode represents a lock in the dependency graph
type LockNode struct {
	ID string
	Lock models.ProjectLock
	CreatedAt time.Time
	LastAccessed time.Time
	WaitingFor []string // IDs of locks this is waiting for
	WaitedBy []string // IDs of locks waiting for this
}

// Dependency represents a wait-for relationship between locks
type Dependency struct {
	From string
	To string
	CreatedAt time.Time
	Reason string
	Weight int // Strength of dependency
}

// LockRegistry tracks all active locks and their states
type LockRegistry struct {
	mu sync.RWMutex
	activeLocks map[string]*LockInfo
	waitingRequests map[string]*WaitInfo
}

// LockInfo contains detailed information about an active lock
type LockInfo struct {
	ID string
	Lock models.ProjectLock
	AcquiredAt time.Time
	Owner string
	RenewalCount int
	LastRenewal time.Time
	Metadata map[string]interface{}
}

// WaitInfo contains information about waiting lock requests
type WaitInfo struct {
	ID string
	Lock models.ProjectLock
	StartedWaitingAt time.Time
	TimeoutAt time.Time
	Priority int
	WaitingFor []string // Lock IDs this request is waiting for
}

// DeadlockMetrics tracks deadlock detection performance
type DeadlockMetrics struct {
	mu sync.RWMutex
	
	// Detection statistics
	ScansCompleted int64
	DeadlocksDetected int64
	FalsePositives int64
	PreventedDeadlocks int64
	
	// Performance metrics
	AverageScanTime time.Duration
	MaxScanTime time.Duration
	GraphSize int
	ActiveNodes int
	
	// Pattern analysis
	CommonPatterns map[string]int
	SuspiciousChains []string
}

// AlertManager handles deadlock alerts and notifications
type AlertManager struct {
	mu sync.RWMutex
	subscribers []DeadlockSubscriber
	alertHistory []DeadlockAlert
	config AlertConfig
}

type AlertConfig struct {
	MaxHistorySize int
	AlertCooldown time.Duration
	SeverityLevels map[string]int
}

type DeadlockSubscriber interface {
	OnDeadlockDetected(alert DeadlockAlert)
	OnDeadlockResolved(alert DeadlockAlert)
	OnSuspiciousPattern(pattern PatternAlert)
}

// DeadlockAlert represents a deadlock detection alert
type DeadlockAlert struct {
	ID string
	Severity string
	DetectedAt time.Time
	ResolvedAt *time.Time
	InvolvedLocks []string
	Cycle []string // The detected cycle
	ResolutionAction string
	Message string
}

// PatternAlert represents suspicious pattern detection
type PatternAlert struct {
	Pattern string
	Occurrences int
	FirstSeen time.Time
	LastSeen time.Time
	InvolvedProjects []string
	Recommendations []string
}

// NewDeadlockDetector creates a new deadlock detector
func NewDeadlockDetector(config DetectionConfig, logger *zap.Logger) *DeadlockDetector {
	return &DeadlockDetector{
		dependencyGraph: &LockGraph{
			nodes: make(map[string]*LockNode),
			edges: make(map[string]map[string]*Dependency),
		},
		lockRegistry: &LockRegistry{
			activeLocks: make(map[string]*LockInfo),
			waitingRequests: make(map[string]*WaitInfo),
		},
		config: config,
		metrics: &DeadlockMetrics{
			CommonPatterns: make(map[string]int),
		},
		alertManager: &AlertManager{
			subscribers: make([]DeadlockSubscriber, 0),
			alertHistory: make([]DeadlockAlert, 0),
			config: AlertConfig{
				MaxHistorySize: 1000,
				AlertCooldown: 5 * time.Minute,
				SeverityLevels: map[string]int{
					"low": 1,
					"medium": 2,
					"high": 3,
					"critical": 4,
				},
			},
		},
		logger: logger,
		stopCh: make(chan struct{}),
		done: make(chan struct{}),
	}
}

// Start begins deadlock detection monitoring
func (dd *DeadlockDetector) Start(ctx context.Context) error {
	go dd.scanForDeadlocks(ctx)
	go dd.pruneGraph(ctx)
	go dd.analyzePatterns(ctx)
	return nil
}

// Stop gracefully stops deadlock detection
func (dd *DeadlockDetector) Stop() error {
	close(dd.stopCh)
	<-dd.done
	return nil
}

// RegisterLockAcquisition records a successful lock acquisition
func (dd *DeadlockDetector) RegisterLockAcquisition(lock models.ProjectLock, owner string) {
	dd.lockRegistry.mu.Lock()
	defer dd.lockRegistry.mu.Unlock()
	
	lockID := generateLockID(lock)
	lockInfo := &LockInfo{
		ID: lockID,
		Lock: lock,
		AcquiredAt: time.Now(),
		Owner: owner,
		Metadata: make(map[string]interface{}),
	}
	
	dd.lockRegistry.activeLocks[lockID] = lockInfo
	
	// Add to dependency graph
	dd.addLockNode(lockInfo)
	
	dd.logger.Debug("Lock acquisition registered",
		zap.String("lock_id", lockID),
		zap.String("owner", owner))
}

// RegisterLockRequest records a lock request that's waiting
func (dd *DeadlockDetector) RegisterLockRequest(lock models.ProjectLock, timeout time.Duration, priority int) string {
	dd.lockRegistry.mu.Lock()
	defer dd.lockRegistry.mu.Unlock()
	
	requestID := generateRequestID()
	waitInfo := &WaitInfo{
		ID: requestID,
		Lock: lock,
		StartedWaitingAt: time.Now(),
		TimeoutAt: time.Now().Add(timeout),
		Priority: priority,
		WaitingFor: dd.findConflictingLocks(lock),
	}
	
	dd.lockRegistry.waitingRequests[requestID] = waitInfo
	
	// Update dependency graph
	dd.addWaitDependencies(waitInfo)
	
	// Check for immediate deadlock potential
	if dd.config.PreventiveAbort && dd.wouldCreateDeadlock(waitInfo) {
		dd.logger.Warn("Preventive deadlock abort triggered",
			zap.String("request_id", requestID))
		return "" // Signal that request should be aborted
	}
	
	dd.logger.Debug("Lock request registered",
		zap.String("request_id", requestID),
		zap.Int("waiting_for_count", len(waitInfo.WaitingFor)))
	
	return requestID
}

// UnregisterLock removes a lock from tracking
func (dd *DeadlockDetector) UnregisterLock(lock models.ProjectLock) {
	dd.lockRegistry.mu.Lock()
	defer dd.lockRegistry.mu.Unlock()
	
	lockID := generateLockID(lock)
	delete(dd.lockRegistry.activeLocks, lockID)
	
	// Remove from dependency graph
	dd.removeLockNode(lockID)
	
	dd.logger.Debug("Lock unregistered", zap.String("lock_id", lockID))
}

// scanForDeadlocks is the main detection loop
func (dd *DeadlockDetector) scanForDeadlocks(ctx context.Context) {
	defer close(dd.done)
	
	shallowTicker := time.NewTicker(dd.config.ScanInterval)
	deepTicker := time.NewTicker(dd.config.DeepScanInterval)
	defer shallowTicker.Stop()
	defer deepTicker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-dd.stopCh:
			return
		case <-shallowTicker.C:
			dd.performShallowScan()
		case <-deepTicker.C:
			dd.performDeepScan()
		}
	}
}

// performShallowScan does a quick deadlock check
func (dd *DeadlockDetector) performShallowScan() {
	start := time.Now()
	defer func() {
		scanTime := time.Since(start)
		dd.updateScanMetrics(scanTime)
	}()
	
	// Check for obvious deadlocks (cycles of length 2)
	dd.dependencyGraph.mu.RLock()
	defer dd.dependencyGraph.mu.RUnlock()
	
	for nodeID, node := range dd.dependencyGraph.nodes {
		for _, waitingFor := range node.WaitingFor {
			if targetNode, exists := dd.dependencyGraph.nodes[waitingFor]; exists {
				// Check if target is waiting for this node (2-cycle)
				for _, targetWaiting := range targetNode.WaitingFor {
					if targetWaiting == nodeID {
						// Deadlock detected!
						dd.handleDeadlock([]string{nodeID, waitingFor})
						return
					}
				}
			}
		}
	}
}

// performDeepScan does comprehensive cycle detection
func (dd *DeadlockDetector) performDeepScan() {
	start := time.Now()
	defer func() {
		scanTime := time.Since(start)
		dd.updateScanMetrics(scanTime)
	}()
	
	dd.dependencyGraph.mu.RLock()
	defer dd.dependencyGraph.mu.RUnlock()
	
	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)
	
	for nodeID := range dd.dependencyGraph.nodes {
		if !visited[nodeID] {
			if cycle := dd.dfsDetectCycle(nodeID, visited, recStack, path); cycle != nil {
				dd.handleDeadlock(cycle)
				return
			}
		}
	}
}

// dfsDetectCycle performs depth-first search for cycle detection
func (dd *DeadlockDetector) dfsDetectCycle(nodeID string, visited, recStack map[string]bool, path []string) []string {
	visited[nodeID] = true
	recStack[nodeID] = true
	path = append(path, nodeID)
	
	node := dd.dependencyGraph.nodes[nodeID]
	if node == nil {
		return nil
	}
	
	for _, waitingFor := range node.WaitingFor {
		if !visited[waitingFor] {
			if cycle := dd.dfsDetectCycle(waitingFor, visited, recStack, path); cycle != nil {
				return cycle
			}
		} else if recStack[waitingFor] {
			// Cycle detected - extract the cycle
			cycleStart := -1
			for i, p := range path {
				if p == waitingFor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				return cycle
			}
		}
	}
	
	recStack[nodeID] = false
	return nil
}

// handleDeadlock responds to detected deadlocks
func (dd *DeadlockDetector) handleDeadlock(cycle []string) {
	dd.metrics.mu.Lock()
	dd.metrics.DeadlocksDetected++
	dd.metrics.mu.Unlock()
	
	// Create deadlock alert
	alert := DeadlockAlert{
		ID: generateAlertID(),
		Severity: "high",
		DetectedAt: time.Now(),
		InvolvedLocks: cycle,
		Cycle: cycle,
		Message: fmt.Sprintf("Deadlock detected involving %d locks", len(cycle)),
	}
	
	// Determine resolution action
	if dd.config.AutoResolve {
		alert.ResolutionAction = dd.resolveDeadlock(cycle)
	} else {
		alert.ResolutionAction = "manual_intervention_required"
	}
	
	// Notify subscribers
	dd.alertManager.notifyDeadlockDetected(alert)
	
	dd.logger.Error("Deadlock detected",
		zap.String("alert_id", alert.ID),
		zap.Strings("cycle", cycle),
		zap.String("resolution", alert.ResolutionAction))
}

// resolveDeadlock automatically resolves detected deadlocks
func (dd *DeadlockDetector) resolveDeadlock(cycle []string) string {
	// Strategy: Abort the youngest lock request in the cycle
	var youngestRequest *WaitInfo
	var youngestID string
	
	dd.lockRegistry.mu.RLock()
	for _, nodeID := range cycle {
		if waitInfo, exists := dd.lockRegistry.waitingRequests[nodeID]; exists {
			if youngestRequest == nil || waitInfo.StartedWaitingAt.After(youngestRequest.StartedWaitingAt) {
				youngestRequest = waitInfo
				youngestID = nodeID
			}
		}
	}
	dd.lockRegistry.mu.RUnlock()
	
	if youngestRequest != nil {
		// Abort the youngest request
		dd.UnregisterLockRequest(youngestID)
		return fmt.Sprintf("aborted_youngest_request:%s", youngestID)
	}
	
	return "no_resolution_found"
}

// Helper methods and utility functions...

func (dd *DeadlockDetector) addLockNode(lockInfo *LockInfo) {
	dd.dependencyGraph.mu.Lock()
	defer dd.dependencyGraph.mu.Unlock()
	
	node := &LockNode{
		ID: lockInfo.ID,
		Lock: lockInfo.Lock,
		CreatedAt: lockInfo.AcquiredAt,
		LastAccessed: time.Now(),
		WaitingFor: make([]string, 0),
		WaitedBy: make([]string, 0),
	}
	
	dd.dependencyGraph.nodes[lockInfo.ID] = node
}

func (dd *DeadlockDetector) removeLockNode(lockID string) {
	dd.dependencyGraph.mu.Lock()
	defer dd.dependencyGraph.mu.Unlock()
	
	delete(dd.dependencyGraph.nodes, lockID)
	delete(dd.dependencyGraph.edges, lockID)
	
	// Remove edges pointing to this node
	for _, edges := range dd.dependencyGraph.edges {
		delete(edges, lockID)
	}
}

// Additional helper methods, metrics collection, pattern analysis, etc.
// would continue here...

// Error definitions
var (
	ErrDeadlockDetected = fmt.Errorf("deadlock detected")
	ErrWouldCreateDeadlock = fmt.Errorf("operation would create deadlock")
)
