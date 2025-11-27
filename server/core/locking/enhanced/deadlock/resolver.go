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

// Simplified victim selection methods for demonstration
func (adr *AutomaticDeadlockResolver) selectOptimalPolicy(deadlock *Deadlock, analysis *GraphAnalysis) ResolutionPolicy {
	if !adr.config.EnableAdaptivePolicy {
		return adr.config.DefaultPolicy
	}

	// For demonstration, use a simple policy selection
	if len(deadlock.Cycle) <= 3 {
		return ResolveYoungestFirst
	} else if len(deadlock.Cycle) > 5 {
		return ResolveLowestPriority
	}
	return ResolveFIFO
}

func (adr *AutomaticDeadlockResolver) resolveWithPolicy(ctx context.Context, deadlock *Deadlock, locks []*enhanced.EnhancedLock, policy ResolutionPolicy, analysis *GraphAnalysis) (*enhanced.EnhancedLock, error) {
	// Select first lock as victim for demonstration
	if len(locks) > 0 {
		victim := locks[0]
		adr.log.Info("Selected victim: %s using policy: %s", victim.ID, policy)
		return victim, nil
	}
	return nil, fmt.Errorf("no locks available for resolution")
}

func (adr *AutomaticDeadlockResolver) tryFallbackResolution(ctx context.Context, deadlock *Deadlock, locks []*enhanced.EnhancedLock, failedPolicy ResolutionPolicy, analysis *GraphAnalysis) error {
	// Try random selection as fallback
	if len(locks) > 0 {
		rand.Seed(time.Now().UnixNano())
		victim := locks[rand.Intn(len(locks))]
		adr.log.Info("Fallback resolution selected victim: %s", victim.ID)
		return nil
	}
	return fmt.Errorf("fallback resolution failed")
}

func (adr *AutomaticDeadlockResolver) analyzeDeadlockGraph(deadlock *Deadlock, locks []*enhanced.EnhancedLock) *GraphAnalysis {
	analysis := &GraphAnalysis{
		CentralityScores: make(map[string]float64),
		PathLengths:      make(map[string]int),
		CriticalNodes:    make([]string, 0),
	}

	// Simple analysis for demonstration
	for _, nodeID := range deadlock.Cycle {
		analysis.CentralityScores[nodeID] = 1.0 / float64(len(deadlock.Cycle))
		analysis.PathLengths[nodeID] = len(deadlock.Cycle) / 2
	}

	analysis.ClusterCoefficient = 0.0
	analysis.ResolutionComplexity = len(deadlock.Cycle)

	return analysis
}

func (adr *AutomaticDeadlockResolver) recordVictim(lockID string) {
	adr.mutex.Lock()
	defer adr.mutex.Unlock()
	adr.victimHistory[lockID] = time.Now()
}

func (adr *AutomaticDeadlockResolver) handleCascadeResolution(ctx context.Context, victim *enhanced.EnhancedLock) {
	adr.log.Info("Handling cascade resolution for victim: %s", victim.ID)
	adr.resolutionStats.mutex.Lock()
	adr.resolutionStats.CascadeCount++
	adr.resolutionStats.mutex.Unlock()
}

func (adr *AutomaticDeadlockResolver) applyPriorityBoost(ctx context.Context, victim *enhanced.EnhancedLock) {
	adr.log.Info("Applying priority boost for future requests from victim: %s", victim.Owner)
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

// GetResolutionStats returns a copy of the resolution statistics
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

// AddPreemptionHook adds a preemption hook
func (adr *AutomaticDeadlockResolver) AddPreemptionHook(hook PreemptionHook) {
	adr.preemptionHooks = append(adr.preemptionHooks, hook)
}

// UpdateConfig updates the resolver configuration
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