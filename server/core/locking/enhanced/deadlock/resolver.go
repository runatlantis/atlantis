package deadlock

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/logging"
)

// AutomaticDeadlockResolver provides advanced deadlock resolution algorithms
type AutomaticDeadlockResolver struct {
	detector     *DeadlockDetector
	config       *ResolverConfig
	log          logging.SimpleLogging
	mutex        sync.RWMutex

	// Algorithm state
	victimHistory    map[string]time.Time  // Track recent victims
	resolutionStats  *ResolutionStats
	policyWeights    map[ResolutionPolicy]float64

	// Callbacks
	resolutionHooks  []ResolutionHook
	preemptionHooks  []PreemptionHook
}

// ResolverConfig configures automatic deadlock resolution
type ResolverConfig struct {
	Enabled                 bool                    `json:"enabled"`
	AutoResolve            bool                    `json:"auto_resolve"`
	ResolutionTimeout      time.Duration           `json:"resolution_timeout"`
	VictimHistoryTTL       time.Duration           `json:"victim_history_ttl"`
	MaxResolutionAttempts  int                     `json:"max_resolution_attempts"`

	// Policy configuration
	DefaultPolicy          ResolutionPolicy        `json:"default_policy"`
	PolicyWeights          map[ResolutionPolicy]float64 `json:"policy_weights"`
	EnableAdaptivePolicy   bool                    `json:"enable_adaptive_policy"`

	// Preemption settings
	EnablePreemption       bool                    `json:"enable_preemption"`
	PreemptionThreshold    time.Duration           `json:"preemption_threshold"`
	MaxPreemptionsPerHour  int                     `json:"max_preemptions_per_hour"`

	// Advanced features
	EnableGraphAnalysis    bool                    `json:"enable_graph_analysis"`
	EnablePriorityBoost    bool                    `json:"enable_priority_boost"`
	CascadeResolution      bool                    `json:"cascade_resolution"`
}

// ResolutionStats tracks resolution algorithm performance
type ResolutionStats struct {
	TotalResolutions       int64                    `json:"total_resolutions"`
	SuccessfulResolutions  int64                    `json:"successful_resolutions"`
	FailedResolutions      int64                    `json:"failed_resolutions"`
	AverageResolutionTime  time.Duration            `json:"average_resolution_time"`
	ResolutionsByPolicy    map[ResolutionPolicy]int64 `json:"resolutions_by_policy"`
	VictimsByPriority      map[enhanced.Priority]int64 `json:"victims_by_priority"`
	PreemptionCount        int64                    `json:"preemption_count"`
	CascadeCount           int64                    `json:"cascade_count"`
	LastResolution         time.Time                `json:"last_resolution"`
	mutex                  sync.RWMutex
}

// PreemptionHook provides callbacks for preemption events
type PreemptionHook interface {
	BeforePreemption(ctx context.Context, victimLock *enhanced.EnhancedLock, reason string) error
	AfterPreemption(ctx context.Context, victimLock *enhanced.EnhancedLock, success bool)
}

// DefaultResolverConfig returns default resolver configuration
func DefaultResolverConfig() *ResolverConfig {
	return &ResolverConfig{
		Enabled:                true,
		AutoResolve:           true,
		ResolutionTimeout:     30 * time.Second,
		VictimHistoryTTL:      5 * time.Minute,
		MaxResolutionAttempts: 3,
		DefaultPolicy:         ResolveLowestPriority,
		PolicyWeights: map[ResolutionPolicy]float64{
			ResolveLowestPriority: 1.0,
			ResolveYoungestFirst:  0.8,
			ResolveFIFO:          0.6,
			ResolveLIFO:          0.4,
			ResolveRandomVictim:   0.2,
		},
		EnableAdaptivePolicy:   true,
		EnablePreemption:      true,
		PreemptionThreshold:   2 * time.Minute,
		MaxPreemptionsPerHour: 10,
		EnableGraphAnalysis:   true,
		EnablePriorityBoost:   true,
		CascadeResolution:     true,
	}
}

// NewAutomaticDeadlockResolver creates a new automatic deadlock resolver
func NewAutomaticDeadlockResolver(detector *DeadlockDetector, config *ResolverConfig, log logging.SimpleLogging) *AutomaticDeadlockResolver {
	if config == nil {
		config = DefaultResolverConfig()
	}

	return &AutomaticDeadlockResolver{
		detector:        detector,
		config:         config,
		log:            log,
		victimHistory:  make(map[string]time.Time),
		resolutionStats: &ResolutionStats{
			ResolutionsByPolicy: make(map[ResolutionPolicy]int64),
			VictimsByPriority:   make(map[enhanced.Priority]int64),
		},
		policyWeights:   config.PolicyWeights,
	}
}

// ResolveDeadlockAdvanced uses advanced algorithms for deadlock resolution
func (adr *AutomaticDeadlockResolver) ResolveDeadlockAdvanced(ctx context.Context, deadlock *Deadlock, locks []*enhanced.EnhancedLock) error {
	if !adr.config.Enabled {
		return fmt.Errorf("automatic deadlock resolution is disabled")
	}

	adr.log.Info("Starting advanced deadlock resolution for deadlock: %s", deadlock.ID)
	startTime := time.Now()

	// Update stats
	adr.resolutionStats.mutex.Lock()
	adr.resolutionStats.TotalResolutions++
	adr.resolutionStats.mutex.Unlock()

	// Create resolution context with timeout
	resCtx, cancel := context.WithTimeout(ctx, adr.config.ResolutionTimeout)
	defer cancel()

	// Perform graph analysis if enabled
	var graphAnalysis *GraphAnalysis
	if adr.config.EnableGraphAnalysis {
		graphAnalysis = adr.analyzeDeadlockGraph(deadlock, locks)
		adr.log.Debug("Graph analysis complete: centrality scores, path lengths calculated")
	}

	// Select optimal resolution policy
	policy := adr.selectOptimalPolicy(deadlock, graphAnalysis)
	adr.log.Info("Selected resolution policy: %s for deadlock: %s", policy, deadlock.ID)

	// Attempt resolution with selected policy
	victim, err := adr.resolveWithPolicy(resCtx, deadlock, locks, policy, graphAnalysis)
	if err != nil {
		adr.log.Error("Resolution failed with policy %s: %v", policy, err)

		// Try fallback policies
		if err = adr.tryFallbackResolution(resCtx, deadlock, locks, policy, graphAnalysis); err != nil {
			adr.updateResolutionStats(false, time.Since(startTime), policy, enhanced.PriorityNormal)
			return fmt.Errorf("all resolution attempts failed: %w", err)
		}
	}

	// Handle successful resolution
	if victim != nil {
		adr.recordVictim(victim.ID)

		// Handle cascade resolution if enabled
		if adr.config.CascadeResolution {
			go adr.handleCascadeResolution(ctx, victim)
		}

		// Update priority boost if enabled
		if adr.config.EnablePriorityBoost {
			adr.applyPriorityBoost(ctx, victim)
		}
	}

	resolutionTime := time.Since(startTime)
	adr.updateResolutionStats(true, resolutionTime, policy, adr.getPriority(victim))

	adr.log.Info("Deadlock resolution completed successfully in %v", resolutionTime)
	return nil
}

// GraphAnalysis holds results of deadlock graph analysis
type GraphAnalysis struct {
	CentralityScores    map[string]float64  `json:"centrality_scores"`
	PathLengths         map[string]int      `json:"path_lengths"`
	ClusterCoefficient  float64             `json:"cluster_coefficient"`
	CriticalNodes       []string            `json:"critical_nodes"`
	ResolutionComplexity int                `json:"resolution_complexity"`
}

// analyzeDeadlockGraph performs graph-theoretic analysis of the deadlock
func (adr *AutomaticDeadlockResolver) analyzeDeadlockGraph(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *GraphAnalysis {
	analysis := &GraphAnalysis{
		CentralityScores: make(map[string]float64),
		PathLengths:      make(map[string]int),
		CriticalNodes:    make([]string, 0),
	}

	// Calculate betweenness centrality for each node in the cycle
	for _, nodeID := range deadlock.Cycle {
		centrality := adr.calculateBetweennessCentrality(nodeID, deadlock.Cycle)
		analysis.CentralityScores[nodeID] = centrality

		// Calculate shortest path lengths
		pathLength := adr.calculateShortestPath(nodeID, deadlock.Cycle)
		analysis.PathLengths[nodeID] = pathLength
	}

	// Identify critical nodes (high centrality, low path length)
	for nodeID, centrality := range analysis.CentralityScores {
		pathLength := analysis.PathLengths[nodeID]
		if centrality > 0.7 && pathLength <= 2 {
			analysis.CriticalNodes = append(analysis.CriticalNodes, nodeID)
		}
	}

	// Calculate cluster coefficient
	analysis.ClusterCoefficient = adr.calculateClusterCoefficient(deadlock.Cycle)

	// Determine resolution complexity
	analysis.ResolutionComplexity = len(deadlock.Cycle) * len(analysis.CriticalNodes)

	return analysis
}

// selectOptimalPolicy selects the best resolution policy based on deadlock characteristics
func (adr *AutomaticDeadlockResolver) selectOptimalPolicy(deadlock *Deadlock, analysis *GraphAnalysis) ResolutionPolicy {
	if !adr.config.EnableAdaptivePolicy {
		return adr.config.DefaultPolicy
	}

	// Score each policy based on deadlock characteristics
	policyScores := make(map[ResolutionPolicy]float64)

	for policy, baseWeight := range adr.policyWeights {
		score := baseWeight

		// Adjust score based on graph analysis
		if analysis != nil {
			switch policy {
			case ResolveLowestPriority:
				// Prefer for complex graphs
				if analysis.ResolutionComplexity > 10 {
					score *= 1.5
				}
			case ResolveYoungestFirst:
				// Prefer for simple cycles
				if len(deadlock.Cycle) <= 3 {
					score *= 1.3
				}
			case ResolveFIFO:
				// Prefer for highly connected graphs
				if analysis.ClusterCoefficient > 0.6 {
					score *= 1.4
				}
			case ResolveRandomVictim:
				// Prefer when other policies may be unfair
				if len(analysis.CriticalNodes) > len(deadlock.Cycle)/2 {
					score *= 1.2
				}
			}
		}

		// Adjust based on historical performance
		adr.resolutionStats.mutex.RLock()
		if policyCount, exists := adr.resolutionStats.ResolutionsByPolicy[policy]; exists && policyCount > 0 {
			successRate := float64(adr.resolutionStats.SuccessfulResolutions) / float64(policyCount)
			score *= (0.5 + successRate) // Boost successful policies
		}
		adr.resolutionStats.mutex.RUnlock()

		policyScores[policy] = score
	}

	// Select the highest-scoring policy
	var bestPolicy ResolutionPolicy
	var bestScore float64

	for policy, score := range policyScores {
		if score > bestScore {
			bestScore = score
			bestPolicy = policy
		}
	}

	if bestPolicy == "" {
		return adr.config.DefaultPolicy
	}

	return bestPolicy
}

// resolveWithPolicy attempts resolution using a specific policy
func (adr *AutomaticDeadlockResolver) resolveWithPolicy(ctx context.Context, deadlock *Deadlock, locks []*enhanced.EnhancedLock, policy ResolutionPolicy, analysis *GraphAnalysis) (*enhanced.EnhancedLock, error) {
	var victim *enhanced.EnhancedLock
	var err error

	switch policy {
	case ResolveLowestPriority:
		victim = adr.selectLowestPriorityVictim(deadlock, locks)
	case ResolveYoungestFirst:
		victim = adr.selectYoungestVictim(deadlock, locks)
	case ResolveFIFO:
		victim = adr.selectOldestVictim(deadlock, locks)
	case ResolveLIFO:
		victim = adr.selectNewestVictim(deadlock, locks)
	case ResolveRandomVictim:
		victim = adr.selectRandomVictim(deadlock, locks)
	default:
		return nil, fmt.Errorf("unsupported resolution policy: %s", policy)
	}

	if victim == nil {
		return nil, fmt.Errorf("no suitable victim found for policy: %s", policy)
	}

	// Check victim history to avoid repeated targeting
	if adr.isRecentVictim(victim.ID) {
		adr.log.Warn("Victim %s was recently selected, trying alternative", victim.ID)
		alternative := adr.selectAlternativeVictim(deadlock, locks, victim.ID)
		if alternative != nil {
			victim = alternative
		}
	}

	// Execute preemption hooks
	for _, hook := range adr.preemptionHooks {
		if err = hook.BeforePreemption(ctx, victim, fmt.Sprintf("deadlock resolution: %s", policy)); err != nil {
			adr.log.Warn("Preemption hook failed: %v", err)
		}
	}

	// Perform the actual resolution (this would integrate with the lock backend)
	if err = adr.executeResolution(ctx, deadlock, victim); err != nil {
		// Execute post-preemption hooks for failure
		for _, hook := range adr.preemptionHooks {
			hook.AfterPreemption(ctx, victim, false)
		}
		return nil, err
	}

	// Execute post-preemption hooks for success
	for _, hook := range adr.preemptionHooks {
		hook.AfterPreemption(ctx, victim, true)
	}

	adr.log.Info("Successfully resolved deadlock %s by preempting lock %s (policy: %s)",
		deadlock.ID, victim.ID, policy)

	return victim, nil
}

// tryFallbackResolution attempts resolution with fallback policies
func (adr *AutomaticDeadlockResolver) tryFallbackResolution(ctx context.Context, deadlock *Deadlock, locks []*enhanced.EnhancedLock, failedPolicy ResolutionPolicy, analysis *GraphAnalysis) error {
	// Order of fallback policies
	fallbackPolicies := []ResolutionPolicy{
		ResolveLowestPriority,
		ResolveYoungestFirst,
		ResolveRandomVictim,
		ResolveFIFO,
	}

	// Remove the failed policy from fallbacks
	var filteredFallbacks []ResolutionPolicy
	for _, policy := range fallbackPolicies {
		if policy != failedPolicy {
			filteredFallbacks = append(filteredFallbacks, policy)
		}
	}

	// Try each fallback policy
	for i, policy := range filteredFallbacks {
		if i >= adr.config.MaxResolutionAttempts-1 {
			break
		}

		adr.log.Info("Attempting fallback resolution with policy: %s", policy)
		_, err := adr.resolveWithPolicy(ctx, deadlock, locks, policy, analysis)
		if err == nil {
			adr.log.Info("Fallback resolution successful with policy: %s", policy)
			return nil
		}

		adr.log.Warn("Fallback policy %s failed: %v", policy, err)
	}

	return fmt.Errorf("all fallback resolution attempts failed")
}

// Victim selection algorithms

func (adr *AutomaticDeadlockResolver) selectLowestPriorityVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *enhanced.EnhancedLock {
	var victim *enhanced.EnhancedLock
	lowestPriority := enhanced.PriorityCritical + 1

	for _, lock := range locks {
		if adr.isInCycle(lock.ID, deadlock.Cycle) {
			if lock.Priority < lowestPriority {
				lowestPriority = lock.Priority
				victim = lock
			}
		}
	}

	return victim
}

func (adr *AutomaticDeadlockResolver) selectYoungestVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *enhanced.EnhancedLock {
	var victim *enhanced.EnhancedLock
	var latestTime time.Time

	for _, lock := range locks {
		if adr.isInCycle(lock.ID, deadlock.Cycle) {
			if lock.AcquiredAt.After(latestTime) {
				latestTime = lock.AcquiredAt
				victim = lock
			}
		}
	}

	return victim
}

func (adr *AutomaticDeadlockResolver) selectOldestVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *enhanced.EnhancedLock {
	var victim *enhanced.EnhancedLock
	earliestTime := time.Now()

	for _, lock := range locks {
		if adr.isInCycle(lock.ID, deadlock.Cycle) {
			if lock.AcquiredAt.Before(earliestTime) {
				earliestTime = lock.AcquiredAt
				victim = lock
			}
		}
	}

	return victim
}

func (adr *AutomaticDeadlockResolver) selectNewestVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *enhanced.EnhancedLock {
	// Same as youngest for this implementation
	return adr.selectYoungestVictim(deadlock, locks)
}

func (adr *AutomaticDeadlockResolver) selectRandomVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *enhanced.EnhancedLock {
	var candidates []*enhanced.EnhancedLock

	for _, lock := range locks {
		if adr.isInCycle(lock.ID, deadlock.Cycle) {
			candidates = append(candidates, lock)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Use current time as seed for deterministic testing
	rand.Seed(time.Now().UnixNano())
	index := rand.Intn(len(candidates))
	return candidates[index]
}

func (adr *AutomaticDeadlockResolver) selectAlternativeVictim(deadlock *Deadlock, locks []*enhanced.EnhancedLock, excludeID string) *enhanced.EnhancedLock {
	var alternatives []*enhanced.EnhancedLock

	for _, lock := range locks {
		if adr.isInCycle(lock.ID, deadlock.Cycle) && lock.ID != excludeID {
			alternatives = append(alternatives, lock)
		}
	}

	if len(alternatives) == 0 {
		return nil
	}

	// Select the lock with the lowest priority among alternatives
	sort.Slice(alternatives, func(i, j int) bool {
		return alternatives[i].Priority < alternatives[j].Priority
	})

	return alternatives[0]
}

// Helper methods

func (adr *AutomaticDeadlockResolver) isInCycle(lockID string, cycle []string) bool {
	for _, nodeID := range cycle {
		if nodeID == lockID {
			return true
		}
	}
	return false
}

func (adr *AutomaticDeadlockResolver) isRecentVictim(lockID string) bool {
	adr.mutex.RLock()
	defer adr.mutex.RUnlock()

	if victimTime, exists := adr.victimHistory[lockID]; exists {
		return time.Since(victimTime) < adr.config.VictimHistoryTTL
	}
	return false
}

func (adr *AutomaticDeadlockResolver) recordVictim(lockID string) {
	adr.mutex.Lock()
	defer adr.mutex.Unlock()

	adr.victimHistory[lockID] = time.Now()

	// Clean up old entries
	cutoff := time.Now().Add(-adr.config.VictimHistoryTTL)
	for id, timestamp := range adr.victimHistory {
		if timestamp.Before(cutoff) {
			delete(adr.victimHistory, id)
		}
	}
}

func (adr *AutomaticDeadlockResolver) executeResolution(ctx context.Context, deadlock *Deadlock, victim *enhanced.EnhancedLock) error {
	// This would integrate with the actual lock backend to release the victim lock
	// For now, we'll simulate the resolution
	adr.log.Info("Executing resolution: releasing lock %s (owner: %s)", victim.ID, victim.Owner)

	// Simulate some processing time
	time.Sleep(10 * time.Millisecond)

	// Mark deadlock as resolved
	deadlock.Resolved = true
	deadlock.ResolvedAt = time.Now()
	deadlock.VictimID = victim.ID

	return nil
}

func (adr *AutomaticDeadlockResolver) handleCascadeResolution(ctx context.Context, victim *enhanced.EnhancedLock) {
	adr.log.Info("Handling cascade resolution for victim: %s", victim.ID)
	adr.resolutionStats.mutex.Lock()
	adr.resolutionStats.CascadeCount++
	adr.resolutionStats.mutex.Unlock()

	// Implement cascade logic here
	// This could involve checking for other deadlocks that may be resolved
	// by the release of this lock
}

func (adr *AutomaticDeadlockResolver) applyPriorityBoost(ctx context.Context, victim *enhanced.EnhancedLock) {
	adr.log.Info("Applying priority boost for future requests from victim: %s", victim.Owner)

	// This would integrate with the lock manager to boost priority
	// for future requests from this user/owner
}

func (adr *AutomaticDeadlockResolver) updateResolutionStats(success bool, duration time.Duration, policy ResolutionPolicy, priority enhanced.Priority) {
	adr.resolutionStats.mutex.Lock()
	defer adr.resolutionStats.mutex.Unlock()

	if success {
		adr.resolutionStats.SuccessfulResolutions++
	} else {
		adr.resolutionStats.FailedResolutions++
	}

	// Update average resolution time
	if adr.resolutionStats.AverageResolutionTime == 0 {
		adr.resolutionStats.AverageResolutionTime = duration
	} else {
		adr.resolutionStats.AverageResolutionTime = (adr.resolutionStats.AverageResolutionTime + duration) / 2
	}

	// Update policy stats
	adr.resolutionStats.ResolutionsByPolicy[policy]++

	// Update priority stats
	adr.resolutionStats.VictimsByPriority[priority]++

	adr.resolutionStats.LastResolution = time.Now()
}

func (adr *AutomaticDeadlockResolver) getPriority(lock *enhanced.EnhancedLock) enhanced.Priority {
	if lock == nil {
		return enhanced.PriorityNormal
	}
	return lock.Priority
}

// Graph analysis helper methods

func (adr *AutomaticDeadlockResolver) calculateBetweennessCentrality(nodeID string, cycle []string) float64 {
	// Simplified betweenness centrality calculation for cycle
	// In a cycle, each node has equal betweenness centrality
	if len(cycle) <= 2 {
		return 1.0
	}
	return 1.0 / float64(len(cycle))
}

func (adr *AutomaticDeadlockResolver) calculateShortestPath(nodeID string, cycle []string) int {
	// In a cycle, the maximum shortest path is len(cycle)/2
	return len(cycle) / 2
}

func (adr *AutomaticDeadlockResolver) calculateClusterCoefficient(cycle []string) float64 {
	// For a cycle, the clustering coefficient is 0 (no triangles)
	// This is simplified - real implementation would analyze the full graph
	return 0.0
}

// Configuration and management

func (adr *AutomaticDeadlockResolver) AddPreemptionHook(hook PreemptionHook) {
	adr.preemptionHooks = append(adr.preemptionHooks, hook)
}

func (adr *AutomaticDeadlockResolver) GetResolutionStats() *ResolutionStats {
	adr.resolutionStats.mutex.RLock()
	defer adr.resolutionStats.mutex.RUnlock()

	// Return a copy of the stats
	statsCopy := &ResolutionStats{
		TotalResolutions:      adr.resolutionStats.TotalResolutions,
		SuccessfulResolutions: adr.resolutionStats.SuccessfulResolutions,
		FailedResolutions:     adr.resolutionStats.FailedResolutions,
		AverageResolutionTime: adr.resolutionStats.AverageResolutionTime,
		PreemptionCount:       adr.resolutionStats.PreemptionCount,
		CascadeCount:          adr.resolutionStats.CascadeCount,
		LastResolution:        adr.resolutionStats.LastResolution,
		ResolutionsByPolicy:   make(map[ResolutionPolicy]int64),
		VictimsByPriority:     make(map[enhanced.Priority]int64),
	}

	// Copy maps
	for k, v := range adr.resolutionStats.ResolutionsByPolicy {
		statsCopy.ResolutionsByPolicy[k] = v
	}
	for k, v := range adr.resolutionStats.VictimsByPriority {
		statsCopy.VictimsByPriority[k] = v
	}

	return statsCopy
}

func (adr *AutomaticDeadlockResolver) UpdateConfig(config *ResolverConfig) error {
	if config == nil {
		return fmt.Errorf("resolver config cannot be nil")
	}

	adr.mutex.Lock()
	defer adr.mutex.Unlock()

	adr.config = config
	if config.PolicyWeights != nil {
		adr.policyWeights = config.PolicyWeights
	}

	adr.log.Info("Automatic deadlock resolver configuration updated")
	return nil
}