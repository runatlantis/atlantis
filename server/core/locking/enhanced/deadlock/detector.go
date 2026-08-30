package deadlock

import (
	"context"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/logging"
)

// DeadlockDetector detects and resolves deadlocks in the locking system
type DeadlockDetector struct {
	waitGraph       *WaitForGraph
	config          *DetectorConfig
	log             logging.SimpleLogging
	metrics         *DeadlockMetrics
	mutex           sync.RWMutex
	running         bool
	stopChan        chan struct{}
	resolutionHooks []ResolutionHook
}

// DetectorConfig configures deadlock detection behavior
type DetectorConfig struct {
	Enabled           bool                    `json:"enabled"`
	CheckInterval     time.Duration           `json:"check_interval"`
	MaxWaitTime       time.Duration           `json:"max_wait_time"`
	ResolutionPolicy  ResolutionPolicy        `json:"resolution_policy"`
	HistorySize       int                     `json:"history_size"`
	EnablePrevention  bool                    `json:"enable_prevention"`
}

// ResolutionPolicy defines how deadlocks should be resolved
type ResolutionPolicy string

const (
	ResolveLIFO         ResolutionPolicy = "lifo"         // Last In, First Out
	ResolveFIFO         ResolutionPolicy = "fifo"         // First In, First Out
	ResolveLowestPriority ResolutionPolicy = "lowest_priority" // Abort lowest priority
	ResolveRandomVictim   ResolutionPolicy = "random"       // Random victim selection
	ResolveYoungestFirst  ResolutionPolicy = "youngest"     // Abort youngest lock
)

// DefaultDetectorConfig returns default deadlock detection configuration
func DefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		Enabled:           true,
		CheckInterval:     30 * time.Second,
		MaxWaitTime:       5 * time.Minute,
		ResolutionPolicy:  ResolveLowestPriority,
		HistorySize:       1000,
		EnablePrevention:  true,
	}
}

// NewDeadlockDetector creates a new deadlock detector
func NewDeadlockDetector(config *DetectorConfig, log logging.SimpleLogging) *DeadlockDetector {
	if config == nil {
		config = DefaultDetectorConfig()
	}

	return &DeadlockDetector{
		waitGraph: NewWaitForGraph(),
		config:    config,
		log:       log,
		metrics:   NewDeadlockMetrics(config.HistorySize),
		stopChan:  make(chan struct{}),
	}
}

// Start begins deadlock detection monitoring
func (dd *DeadlockDetector) Start(ctx context.Context) {
	if !dd.config.Enabled {
		dd.log.Info("Deadlock detection is disabled")
		return
	}

	dd.mutex.Lock()
	if dd.running {
		dd.mutex.Unlock()
		return
	}
	dd.running = true
	dd.mutex.Unlock()

	dd.log.Info("Starting deadlock detector with %v check interval", dd.config.CheckInterval)

	go dd.detectLoop(ctx)
}

// Stop stops deadlock detection monitoring
func (dd *DeadlockDetector) Stop() {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	if !dd.running {
		return
	}

	dd.log.Info("Stopping deadlock detector")
	dd.running = false
	close(dd.stopChan)
}

// AddLockRequest adds a lock request to the wait-for graph
func (dd *DeadlockDetector) AddLockRequest(request *enhanced.EnhancedLockRequest, blockedBy []*enhanced.EnhancedLock) error {
	if !dd.config.Enabled {
		return nil
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// Add the waiting relationship to the graph
	for _, lock := range blockedBy {
		dd.waitGraph.AddEdge(request.ID, lock.Owner)
	}

	dd.metrics.IncrementWaitingRequests()
	return nil
}

// RemoveLockRequest removes a lock request from the wait-for graph
func (dd *DeadlockDetector) RemoveLockRequest(requestID string) {
	if !dd.config.Enabled {
		return
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	dd.waitGraph.RemoveNode(requestID)
	dd.metrics.DecrementWaitingRequests()
}

// AddLockAcquisition records a lock acquisition
func (dd *DeadlockDetector) AddLockAcquisition(lock *enhanced.EnhancedLock) {
	if !dd.config.Enabled {
		return
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// Remove from wait graph since lock is acquired
	dd.waitGraph.RemoveNode(lock.ID)
}

// CheckForDeadlocks performs immediate deadlock detection
func (dd *DeadlockDetector) CheckForDeadlocks(ctx context.Context) ([]*Deadlock, error) {
	if !dd.config.Enabled {
		return nil, nil
	}

	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	cycles := dd.waitGraph.FindCycles()
	var deadlocks []*Deadlock

	for _, cycle := range cycles {
		deadlock := &Deadlock{
			ID:          generateDeadlockID(),
			Cycle:       cycle,
			DetectedAt:  time.Now(),
			Resolved:    false,
		}

		// Analyze the deadlock for additional context
		dd.analyzeDeadlock(deadlock)
		deadlocks = append(deadlocks, deadlock)
	}

	if len(deadlocks) > 0 {
		dd.metrics.IncrementDeadlocksDetected(len(deadlocks))
		dd.log.Warn("Detected %d deadlocks", len(deadlocks))
	}

	return deadlocks, nil
}

// ResolveDeadlock resolves a detected deadlock
func (dd *DeadlockDetector) ResolveDeadlock(ctx context.Context, deadlock *Deadlock) error {
	if deadlock.Resolved {
		return nil
	}

	dd.log.Info("Resolving deadlock: %s with policy: %s", deadlock.ID, dd.config.ResolutionPolicy)

	// Select victim based on resolution policy
	victim := dd.selectVictim(deadlock)
	if victim == "" {
		return &enhanced.LockError{
			Type:    "ResolutionFailed",
			Message: "failed to select victim for deadlock resolution",
			Code:    "RESOLUTION_FAILED",
		}
	}

	// Execute resolution hooks
	for _, hook := range dd.resolutionHooks {
		if err := hook.BeforeResolution(ctx, deadlock, victim); err != nil {
			dd.log.Warn("Resolution hook failed: %v", err)
		}
	}

	// Mark as resolved
	deadlock.Resolved = true
	deadlock.ResolvedAt = time.Now()
	deadlock.VictimID = victim

	// Remove victim from wait graph
	dd.mutex.Lock()
	dd.waitGraph.RemoveNode(victim)
	dd.mutex.Unlock()

	dd.metrics.IncrementDeadlocksResolved()
	dd.metrics.RecordResolution(deadlock)

	// Execute post-resolution hooks
	for _, hook := range dd.resolutionHooks {
		hook.AfterResolution(ctx, deadlock)
	}

	dd.log.Info("Deadlock resolved: %s, victim: %s", deadlock.ID, victim)
	return nil
}

// PreventDeadlock checks if a new lock request would create a deadlock
func (dd *DeadlockDetector) PreventDeadlock(request *enhanced.EnhancedLockRequest, blockedBy []*enhanced.EnhancedLock) (bool, error) {
	if !dd.config.EnablePrevention {
		return true, nil // Allow if prevention is disabled
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	// Temporarily add the edges to simulate the request
	tempGraph := dd.waitGraph.Clone()
	for _, lock := range blockedBy {
		tempGraph.AddEdge(request.ID, lock.Owner)
	}

	// Check if this would create a cycle
	cycles := tempGraph.FindCycles()
	wouldDeadlock := len(cycles) > 0

	if wouldDeadlock {
		dd.metrics.IncrementPreventedDeadlocks()
		dd.log.Info("Prevented potential deadlock for request: %s", request.ID)
	}

	return !wouldDeadlock, nil
}

// GetStats returns deadlock detection statistics
func (dd *DeadlockDetector) GetStats() *DeadlockMetrics {
	return dd.metrics
}

// AddResolutionHook adds a hook that will be called during deadlock resolution
func (dd *DeadlockDetector) AddResolutionHook(hook ResolutionHook) {
	dd.resolutionHooks = append(dd.resolutionHooks, hook)
}

// detectLoop runs the periodic deadlock detection
func (dd *DeadlockDetector) detectLoop(ctx context.Context) {
	ticker := time.NewTicker(dd.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dd.stopChan:
			return
		case <-ticker.C:
			deadlocks, err := dd.CheckForDeadlocks(ctx)
			if err != nil {
				dd.log.Error("Error during deadlock detection: %v", err)
				continue
			}

			// Resolve detected deadlocks
			for _, deadlock := range deadlocks {
				if err := dd.ResolveDeadlock(ctx, deadlock); err != nil {
					dd.log.Error("Error resolving deadlock %s: %v", deadlock.ID, err)
				}
			}
		}
	}
}

// selectVictim selects which request should be aborted to resolve the deadlock
func (dd *DeadlockDetector) selectVictim(deadlock *Deadlock) string {
	if len(deadlock.Cycle) == 0 {
		return ""
	}

	switch dd.config.ResolutionPolicy {
	case ResolveLIFO:
		return dd.selectNewestInCycle(deadlock.Cycle)
	case ResolveFIFO:
		return dd.selectOldestInCycle(deadlock.Cycle)
	case ResolveLowestPriority:
		return dd.selectLowestPriorityInCycle(deadlock.Cycle)
	case ResolveYoungestFirst:
		return dd.selectYoungestInCycle(deadlock.Cycle)
	case ResolveRandomVictim:
		return dd.selectRandomInCycle(deadlock.Cycle)
	default:
		// Default to lowest priority
		return dd.selectLowestPriorityInCycle(deadlock.Cycle)
	}
}

func (dd *DeadlockDetector) selectNewestInCycle(cycle []string) string {
	// For now, return the first node - would need request timestamp data
	if len(cycle) > 0 {
		return cycle[0]
	}
	return ""
}

func (dd *DeadlockDetector) selectOldestInCycle(cycle []string) string {
	// For now, return the last node - would need request timestamp data
	if len(cycle) > 0 {
		return cycle[len(cycle)-1]
	}
	return ""
}

func (dd *DeadlockDetector) selectLowestPriorityInCycle(cycle []string) string {
	// For now, return the first node - would need priority data
	if len(cycle) > 0 {
		return cycle[0]
	}
	return ""
}

func (dd *DeadlockDetector) selectYoungestInCycle(cycle []string) string {
	// For now, return the first node - would need age data
	if len(cycle) > 0 {
		return cycle[0]
	}
	return ""
}

func (dd *DeadlockDetector) selectRandomInCycle(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	// Simple pseudo-random selection based on current time
	index := int(time.Now().UnixNano()) % len(cycle)
	return cycle[index]
}

// analyzeDeadlock adds additional context to a deadlock
func (dd *DeadlockDetector) analyzeDeadlock(deadlock *Deadlock) {
	// Add metadata about the deadlock
	deadlock.Metadata = map[string]interface{}{
		"cycle_length":    len(deadlock.Cycle),
		"detection_time":  deadlock.DetectedAt,
		"resolution_policy": dd.config.ResolutionPolicy,
	}
}

// Deadlock represents a detected deadlock situation
type Deadlock struct {
	ID         string                 `json:"id"`
	Cycle      []string               `json:"cycle"`       // Node IDs in the cycle
	DetectedAt time.Time              `json:"detected_at"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt time.Time              `json:"resolved_at,omitempty"`
	VictimID   string                 `json:"victim_id,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ResolutionHook provides callbacks for deadlock resolution events
type ResolutionHook interface {
	BeforeResolution(ctx context.Context, deadlock *Deadlock, victim string) error
	AfterResolution(ctx context.Context, deadlock *Deadlock)
}

// WaitForGraph represents the wait-for graph for deadlock detection
type WaitForGraph struct {
	nodes map[string]bool          // Set of nodes
	edges map[string][]string      // Adjacency list
	mutex sync.RWMutex
}

// NewWaitForGraph creates a new wait-for graph
func NewWaitForGraph() *WaitForGraph {
	return &WaitForGraph{
		nodes: make(map[string]bool),
		edges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (wfg *WaitForGraph) AddNode(nodeID string) {
	wfg.mutex.Lock()
	defer wfg.mutex.Unlock()

	wfg.nodes[nodeID] = true
	if wfg.edges[nodeID] == nil {
		wfg.edges[nodeID] = make([]string, 0)
	}
}

// RemoveNode removes a node and all its edges from the graph
func (wfg *WaitForGraph) RemoveNode(nodeID string) {
	wfg.mutex.Lock()
	defer wfg.mutex.Unlock()

	delete(wfg.nodes, nodeID)
	delete(wfg.edges, nodeID)

	// Remove edges pointing to this node
	for source, targets := range wfg.edges {
		newTargets := make([]string, 0)
		for _, target := range targets {
			if target != nodeID {
				newTargets = append(newTargets, target)
			}
		}
		wfg.edges[source] = newTargets
	}
}

// AddEdge adds a directed edge from source to target (source waits for target)
func (wfg *WaitForGraph) AddEdge(source, target string) {
	wfg.mutex.Lock()
	defer wfg.mutex.Unlock()

	wfg.nodes[source] = true
	wfg.nodes[target] = true

	if wfg.edges[source] == nil {
		wfg.edges[source] = make([]string, 0)
	}
	if wfg.edges[target] == nil {
		wfg.edges[target] = make([]string, 0)
	}

	// Add edge if it doesn't already exist
	for _, existing := range wfg.edges[source] {
		if existing == target {
			return // Edge already exists
		}
	}

	wfg.edges[source] = append(wfg.edges[source], target)
}

// FindCycles finds all cycles in the graph using DFS
func (wfg *WaitForGraph) FindCycles() [][]string {
	wfg.mutex.RLock()
	defer wfg.mutex.RUnlock()

	var cycles [][]string
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	for node := range wfg.nodes {
		if !visited[node] {
			path := make([]string, 0)
			wfg.dfsCycles(node, visited, recursionStack, path, &cycles)
		}
	}

	return cycles
}

// dfsCycles performs DFS to find cycles
func (wfg *WaitForGraph) dfsCycles(node string, visited, recursionStack map[string]bool, path []string, cycles *[][]string) {
	visited[node] = true
	recursionStack[node] = true
	path = append(path, node)

	for _, neighbor := range wfg.edges[node] {
		if !visited[neighbor] {
			wfg.dfsCycles(neighbor, visited, recursionStack, path, cycles)
		} else if recursionStack[neighbor] {
			// Found a cycle - extract the cycle from the path
			cycleStart := -1
			for i, pathNode := range path {
				if pathNode == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cycle := make([]string, len(path)-cycleStart)
				copy(cycle, path[cycleStart:])
				*cycles = append(*cycles, cycle)
			}
		}
	}

	recursionStack[node] = false
}

// Clone creates a copy of the wait-for graph
func (wfg *WaitForGraph) Clone() *WaitForGraph {
	wfg.mutex.RLock()
	defer wfg.mutex.RUnlock()

	clone := NewWaitForGraph()

	// Copy nodes
	for node := range wfg.nodes {
		clone.nodes[node] = true
	}

	// Copy edges
	for source, targets := range wfg.edges {
		clone.edges[source] = make([]string, len(targets))
		copy(clone.edges[source], targets)
	}

	return clone
}

// GetStats returns graph statistics
func (wfg *WaitForGraph) GetStats() map[string]interface{} {
	wfg.mutex.RLock()
	defer wfg.mutex.RUnlock()

	totalEdges := 0
	for _, edges := range wfg.edges {
		totalEdges += len(edges)
	}

	return map[string]interface{}{
		"nodes": len(wfg.nodes),
		"edges": totalEdges,
	}
}

// DeadlockMetrics tracks deadlock-related metrics
type DeadlockMetrics struct {
	DeadlocksDetected     int64      `json:"deadlocks_detected"`
	DeadlocksResolved     int64      `json:"deadlocks_resolved"`
	PreventedDeadlocks    int64      `json:"prevented_deadlocks"`
	WaitingRequests       int64      `json:"waiting_requests"`
	ResolutionHistory     []*Deadlock `json:"resolution_history"`
	MaxHistorySize        int        `json:"max_history_size"`
	mutex                 sync.RWMutex
}

// NewDeadlockMetrics creates new deadlock metrics
func NewDeadlockMetrics(historySize int) *DeadlockMetrics {
	return &DeadlockMetrics{
		ResolutionHistory: make([]*Deadlock, 0),
		MaxHistorySize:    historySize,
	}
}

func (dm *DeadlockMetrics) IncrementDeadlocksDetected(count int) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	dm.DeadlocksDetected += int64(count)
}

func (dm *DeadlockMetrics) IncrementDeadlocksResolved() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	dm.DeadlocksResolved++
}

func (dm *DeadlockMetrics) IncrementPreventedDeadlocks() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	dm.PreventedDeadlocks++
}

func (dm *DeadlockMetrics) IncrementWaitingRequests() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	dm.WaitingRequests++
}

func (dm *DeadlockMetrics) DecrementWaitingRequests() {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()
	if dm.WaitingRequests > 0 {
		dm.WaitingRequests--
	}
}

func (dm *DeadlockMetrics) RecordResolution(deadlock *Deadlock) {
	dm.mutex.Lock()
	defer dm.mutex.Unlock()

	dm.ResolutionHistory = append(dm.ResolutionHistory, deadlock)

	// Keep only recent history
	if len(dm.ResolutionHistory) > dm.MaxHistorySize {
		dm.ResolutionHistory = dm.ResolutionHistory[1:]
	}
}

func (dm *DeadlockMetrics) GetSnapshot() *DeadlockMetrics {
	dm.mutex.RLock()
	defer dm.mutex.RUnlock()

	return &DeadlockMetrics{
		DeadlocksDetected:  dm.DeadlocksDetected,
		DeadlocksResolved:  dm.DeadlocksResolved,
		PreventedDeadlocks: dm.PreventedDeadlocks,
		WaitingRequests:    dm.WaitingRequests,
		MaxHistorySize:     dm.MaxHistorySize,
		ResolutionHistory:  append([]*Deadlock(nil), dm.ResolutionHistory...),
	}
}

// generateDeadlockID generates a unique ID for a deadlock
func generateDeadlockID() string {
	return "deadlock_" + time.Now().Format("20060102_150405") + "_" +
		   time.Now().Format("000")
}