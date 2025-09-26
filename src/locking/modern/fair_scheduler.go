package modern

import (
	"container/heap"
	"context"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/logging"
)

// FairScheduler implements advanced fair scheduling algorithms for lock requests
type FairScheduler struct {
	config        *FairSchedulingConfig
	log           logging.SimpleLogging

	// Algorithm-specific implementations
	priorityQueue      *PriorityBasedQueue
	wrr               *WeightedRoundRobin
	lottery           *LotteryScheduler
	cfs               *CompletelyFairScheduler

	// Anti-starvation tracking
	starvationTracker  *StarvationTracker

	// Load balancing
	loadBalancer      *LoadBalancer

	// Metrics
	metrics           *SchedulerMetrics
	mutex             sync.RWMutex
}

// SchedulerMetrics tracks scheduling performance and fairness
type SchedulerMetrics struct {
	// Request metrics
	TotalRequests        int64                      `json:"total_requests"`
	ProcessedRequests    int64                      `json:"processed_requests"`
	QueuedRequests       int64                      `json:"queued_requests"`

	// Fairness metrics
	RequestsByPriority   map[enhanced.Priority]int64 `json:"requests_by_priority"`
	RequestsByUser       map[string]int64            `json:"requests_by_user"`
	RequestsByProject    map[string]int64            `json:"requests_by_project"`

	// Timing metrics
	AverageWaitTime      map[enhanced.Priority]time.Duration `json:"average_wait_time"`
	MaxWaitTime         time.Duration                      `json:"max_wait_time"`
	MinWaitTime         time.Duration                      `json:"min_wait_time"`

	// Starvation metrics
	StarvationEvents     int64                              `json:"starvation_events"`
	StarvationBoosts     int64                              `json:"starvation_boosts"`

	// Algorithm-specific metrics
	AlgorithmMetrics     map[string]interface{}             `json:"algorithm_metrics"`

	LastUpdated         time.Time                          `json:"last_updated"`
}

// PriorityBasedQueue implements traditional priority-based scheduling
type PriorityBasedQueue struct {
	queues    map[enhanced.Priority]*RequestQueue
	mutex     sync.RWMutex
}

// RequestQueue holds requests for a specific priority level
type RequestQueue struct {
	requests  []*QueuedRequest
	weights   map[string]int  // User/project weights
	lastServed map[string]time.Time
	mutex     sync.RWMutex
}

// QueuedRequest represents a queued lock request with scheduling metadata
type QueuedRequest struct {
	Request         *enhanced.EnhancedLockRequest
	QueuedAt        time.Time
	Priority        enhanced.Priority
	EffectivePriority float64  // Can be boosted for anti-starvation
	Weight          int
	StarvationBoost float64
	Attempts        int
}

// WeightedRoundRobin implements weighted round-robin scheduling
type WeightedRoundRobin struct {
	userCounters    map[string]int
	projectCounters map[string]int
	userWeights     map[string]int
	projectWeights  map[string]int
	requests        []*QueuedRequest
	mutex           sync.RWMutex
}

// LotteryScheduler implements lottery-based fair scheduling
type LotteryScheduler struct {
	totalTickets  int
	userTickets   map[string]int
	requests      []*QueuedRequest
	random        *rand.Rand
	mutex         sync.RWMutex
}

// CompletelyFairScheduler implements CFS-like scheduling for locks
type CompletelyFairScheduler struct {
	vruntime     map[string]float64  // Virtual runtime per user/project
	weights      map[string]int
	requests     []*VRuntimeRequest
	timeSlice    time.Duration
	mutex        sync.RWMutex
}

// VRuntimeRequest extends QueuedRequest with virtual runtime information
type VRuntimeRequest struct {
	*QueuedRequest
	VirtualRuntime  float64
	TimeSlice       time.Duration
	LastScheduled   time.Time
}

// StarvationTracker monitors and prevents request starvation
type StarvationTracker struct {
	config         *FairSchedulingConfig
	requestAge     map[string]time.Time
	boostHistory   map[string][]time.Time
	mutex          sync.RWMutex
}

// LoadBalancer distributes load across different dimensions
type LoadBalancer struct {
	config         *FairSchedulingConfig
	userLoad       map[string]int
	projectLoad    map[string]int
	resourceLoad   map[string]int
	windowStart    time.Time
	mutex          sync.RWMutex
}

// NewFairScheduler creates a new fair scheduler
func NewFairScheduler(config *FairSchedulingConfig, log logging.SimpleLogging) *FairScheduler {
	scheduler := &FairScheduler{
		config: config,
		log:    log,
		metrics: &SchedulerMetrics{
			RequestsByPriority: make(map[enhanced.Priority]int64),
			RequestsByUser:     make(map[string]int64),
			RequestsByProject:  make(map[string]int64),
			AverageWaitTime:    make(map[enhanced.Priority]time.Duration),
			AlgorithmMetrics:   make(map[string]interface{}),
			LastUpdated:        time.Now(),
		},
		starvationTracker: &StarvationTracker{
			config:       config,
			requestAge:   make(map[string]time.Time),
			boostHistory: make(map[string][]time.Time),
		},
	}

	// Initialize algorithm-specific schedulers
	scheduler.initSchedulers()

	return scheduler
}

// initSchedulers initializes the specific scheduling algorithms
func (fs *FairScheduler) initSchedulers() {
	// Priority-based queue
	fs.priorityQueue = &PriorityBasedQueue{
		queues: make(map[enhanced.Priority]*RequestQueue),
	}
	for priority := enhanced.PriorityLow; priority <= enhanced.PriorityCritical; priority++ {
		fs.priorityQueue.queues[priority] = &RequestQueue{
			requests:   make([]*QueuedRequest, 0),
			weights:    make(map[string]int),
			lastServed: make(map[string]time.Time),
		}
	}

	// Weighted round-robin
	fs.wrr = &WeightedRoundRobin{
		userCounters:    make(map[string]int),
		projectCounters: make(map[string]int),
		userWeights:     fs.config.UserWeights,
		projectWeights:  fs.config.ProjectWeights,
		requests:        make([]*QueuedRequest, 0),
	}

	// Lottery scheduler
	fs.lottery = &LotteryScheduler{
		userTickets: make(map[string]int),
		requests:    make([]*QueuedRequest, 0),
		random:      rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Completely fair scheduler
	fs.cfs = &CompletelyFairScheduler{
		vruntime:  make(map[string]float64),
		weights:   fs.config.UserWeights,
		requests:  make([]*VRuntimeRequest, 0),
		timeSlice: fs.config.TimeSliceDuration,
	}

	// Load balancer
	if fs.config.EnableLoadBalancing {
		fs.loadBalancer = &LoadBalancer{
			config:       fs.config,
			userLoad:     make(map[string]int),
			projectLoad:  make(map[string]int),
			resourceLoad: make(map[string]int),
			windowStart:  time.Now(),
		}
	}
}

// Schedule selects the next request to process based on the configured algorithm
func (fs *FairScheduler) Schedule(ctx context.Context) (*enhanced.EnhancedLockRequest, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	var request *enhanced.EnhancedLockRequest
	var err error

	// Apply anti-starvation before scheduling
	fs.applyAntiStarvation()

	// Apply load balancing if enabled
	if fs.loadBalancer != nil {
		fs.loadBalancer.rebalance()
	}

	// Execute the appropriate scheduling algorithm
	switch fs.config.Algorithm {
	case "priority":
		request, err = fs.schedulePriorityBased()
	case "weighted_round_robin":
		request, err = fs.scheduleWeightedRoundRobin()
	case "lottery":
		request, err = fs.scheduleLottery()
	case "cfs":
		request, err = fs.scheduleCompletelyFair()
	default:
		return nil, fmt.Errorf("unknown scheduling algorithm: %s", fs.config.Algorithm)
	}

	if err != nil {
		return nil, err
	}

	if request != nil {
		fs.updateMetrics(request)
		fs.log.Debug("Scheduled request %s using %s algorithm", request.ID, fs.config.Algorithm)
	}

	return request, nil
}

// Enqueue adds a new request to the scheduler
func (fs *FairScheduler) Enqueue(ctx context.Context, request *enhanced.EnhancedLockRequest) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	queuedRequest := &QueuedRequest{
		Request:           request,
		QueuedAt:          time.Now(),
		Priority:          request.Priority,
		EffectivePriority: float64(request.Priority),
		Weight:            fs.getWeight(request),
		StarvationBoost:   1.0,
		Attempts:          0,
	}

	// Add to starvation tracker
	fs.starvationTracker.trackRequest(request.ID, queuedRequest.QueuedAt)

	// Enqueue to appropriate algorithm
	switch fs.config.Algorithm {
	case "priority":
		return fs.enqueuePriorityBased(queuedRequest)
	case "weighted_round_robin":
		return fs.enqueueWeightedRoundRobin(queuedRequest)
	case "lottery":
		return fs.enqueueLottery(queuedRequest)
	case "cfs":
		return fs.enqueueCompletelyFair(queuedRequest)
	default:
		return fmt.Errorf("unknown scheduling algorithm: %s", fs.config.Algorithm)
	}
}

// Priority-based scheduling implementation
func (fs *FairScheduler) schedulePriorityBased() (*enhanced.EnhancedLockRequest, error) {
	// Process priorities from highest to lowest
	priorities := []enhanced.Priority{
		enhanced.PriorityCritical,
		enhanced.PriorityHigh,
		enhanced.PriorityNormal,
		enhanced.PriorityLow,
	}

	for _, priority := range priorities {
		queue := fs.priorityQueue.queues[priority]
		queue.mutex.Lock()

		if len(queue.requests) > 0 {
			// Fair selection within the same priority using weights
			selected := fs.selectFairlyWithinPriority(queue)
			if selected != nil {
				// Remove from queue
				fs.removeFromQueue(queue, selected)
				queue.mutex.Unlock()
				return selected.Request, nil
			}
		}

		queue.mutex.Unlock()
	}

	return nil, nil // No requests available
}

// selectFairlyWithinPriority selects fairly among requests of the same priority
func (fs *FairScheduler) selectFairlyWithinPriority(queue *RequestQueue) *QueuedRequest {
	if len(queue.requests) == 0 {
		return nil
	}

	// Simple round-robin among users within same priority
	userLastServed := make([]string, 0)
	for user := range queue.lastServed {
		userLastServed = append(userLastServed, user)
	}

	// Sort by last served time
	sort.Slice(userLastServed, func(i, j int) bool {
		return queue.lastServed[userLastServed[i]].Before(queue.lastServed[userLastServed[j]])
	})

	// Find request from user who was served least recently
	for _, user := range userLastServed {
		for _, req := range queue.requests {
			if req.Request.User.Username == user {
				queue.lastServed[user] = time.Now()
				return req
			}
		}
	}

	// If no users in lastServed map, take the first request
	if len(queue.requests) > 0 {
		user := queue.requests[0].Request.User.Username
		queue.lastServed[user] = time.Now()
		return queue.requests[0]
	}

	return nil
}

// removeFromQueue removes a request from the queue
func (fs *FairScheduler) removeFromQueue(queue *RequestQueue, request *QueuedRequest) {
	for i, req := range queue.requests {
		if req.Request.ID == request.Request.ID {
			queue.requests = append(queue.requests[:i], queue.requests[i+1:]...)
			break
		}
	}
}

// Weighted round-robin scheduling implementation
func (fs *FairScheduler) scheduleWeightedRoundRobin() (*enhanced.EnhancedLockRequest, error) {
	fs.wrr.mutex.Lock()
	defer fs.wrr.mutex.Unlock()

	if len(fs.wrr.requests) == 0 {
		return nil, nil
	}

	// Find user/project with highest deficit (weight - counter ratio)
	var selectedRequest *QueuedRequest
	highestDeficit := -1.0

	for _, req := range fs.wrr.requests {
		user := req.Request.User.Username
		project := req.Request.Resource.Namespace

		// Calculate user deficit
		userWeight := fs.wrr.userWeights[user]
		if userWeight == 0 {
			userWeight = 1 // Default weight
		}
		userCounter := fs.wrr.userCounters[user]
		userDeficit := float64(userWeight) - float64(userCounter)

		// Calculate project deficit
		projectWeight := fs.wrr.projectWeights[project]
		if projectWeight == 0 {
			projectWeight = 1 // Default weight
		}
		projectCounter := fs.wrr.projectCounters[project]
		projectDeficit := float64(projectWeight) - float64(projectCounter)

		// Combined deficit (could be weighted differently)
		combinedDeficit := userDeficit + projectDeficit

		if combinedDeficit > highestDeficit {
			highestDeficit = combinedDeficit
			selectedRequest = req
		}
	}

	if selectedRequest != nil {
		// Update counters
		user := selectedRequest.Request.User.Username
		project := selectedRequest.Request.Resource.Namespace

		fs.wrr.userCounters[user]++
		fs.wrr.projectCounters[project]++

		// Remove from queue
		for i, req := range fs.wrr.requests {
			if req.Request.ID == selectedRequest.Request.ID {
				fs.wrr.requests = append(fs.wrr.requests[:i], fs.wrr.requests[i+1:]...)
				break
			}
		}

		return selectedRequest.Request, nil
	}

	return nil, nil
}

// Lottery scheduling implementation
func (fs *FairScheduler) scheduleLottery() (*enhanced.EnhancedLockRequest, error) {
	fs.lottery.mutex.Lock()
	defer fs.lottery.mutex.Unlock()

	if len(fs.lottery.requests) == 0 {
		return nil, nil
	}

	// Calculate total tickets
	totalTickets := 0
	for _, req := range fs.lottery.requests {
		user := req.Request.User.Username
		tickets := fs.lottery.userTickets[user]
		if tickets == 0 {
			// Base tickets + priority bonus + starvation boost
			tickets = 10 + int(req.Priority)*5 + int(req.StarvationBoost*5)
			fs.lottery.userTickets[user] = tickets
		}
		totalTickets += tickets
	}

	// Draw lottery
	winningTicket := fs.lottery.random.Intn(totalTickets)
	currentTicket := 0

	for i, req := range fs.lottery.requests {
		user := req.Request.User.Username
		tickets := fs.lottery.userTickets[user]
		currentTicket += tickets

		if currentTicket > winningTicket {
			// Winner found
			selectedRequest := fs.lottery.requests[i]

			// Remove from queue
			fs.lottery.requests = append(fs.lottery.requests[:i], fs.lottery.requests[i+1:]...)

			return selectedRequest.Request, nil
		}
	}

	// Fallback (shouldn't happen)
	if len(fs.lottery.requests) > 0 {
		selected := fs.lottery.requests[0]
		fs.lottery.requests = fs.lottery.requests[1:]
		return selected.Request, nil
	}

	return nil, nil
}

// CFS scheduling implementation
func (fs *FairScheduler) scheduleCompletelyFair() (*enhanced.EnhancedLockRequest, error) {
	fs.cfs.mutex.Lock()
	defer fs.cfs.mutex.Unlock()

	if len(fs.cfs.requests) == 0 {
		return nil, nil
	}

	// Sort by virtual runtime (lowest first)
	sort.Slice(fs.cfs.requests, func(i, j int) bool {
		return fs.cfs.requests[i].VirtualRuntime < fs.cfs.requests[j].VirtualRuntime
	})

	// Select the request with lowest virtual runtime
	selectedRequest := fs.cfs.requests[0]

	// Update virtual runtime
	user := selectedRequest.Request.User.Username
	weight := fs.cfs.weights[user]
	if weight == 0 {
		weight = 1024 // Default weight (like Linux CFS)
	}

	// Virtual runtime increment = (actual_time / weight)
	// For simplicity, use time slice as actual time
	vruntimeIncrement := float64(fs.cfs.timeSlice.Nanoseconds()) / float64(weight)
	fs.cfs.vruntime[user] += vruntimeIncrement
	selectedRequest.VirtualRuntime = fs.cfs.vruntime[user]

	// Remove from queue
	fs.cfs.requests = fs.cfs.requests[1:]

	return selectedRequest.Request, nil
}

// applyAntiStarvation applies anti-starvation mechanisms
func (fs *FairScheduler) applyAntiStarvation() {
	fs.starvationTracker.mutex.Lock()
	defer fs.starvationTracker.mutex.Unlock()

	now := time.Now()
	threshold := fs.config.StarvationThreshold

	for requestID, queuedTime := range fs.starvationTracker.requestAge {
		if now.Sub(queuedTime) > threshold {
			// Apply starvation boost
			fs.applyStarvationBoost(requestID)
			fs.metrics.StarvationEvents++
		}
	}
}

// applyStarvationBoost applies a boost to a starving request
func (fs *FairScheduler) applyStarvationBoost(requestID string) {
	boost := fs.config.StarvationBoost

	// Apply boost to different algorithms differently
	switch fs.config.Algorithm {
	case "priority":
		fs.boostPriorityRequest(requestID, boost)
	case "weighted_round_robin":
		fs.boostWRRRequest(requestID, boost)
	case "lottery":
		fs.boostLotteryRequest(requestID, boost)
	case "cfs":
		fs.boostCFSRequest(requestID, boost)
	}

	fs.metrics.StarvationBoosts++
	fs.log.Info("Applied starvation boost %.2f to request %s", boost, requestID)
}

// Helper functions for applying boosts to different algorithms
func (fs *FairScheduler) boostPriorityRequest(requestID string, boost float64) {
	for _, queue := range fs.priorityQueue.queues {
		queue.mutex.Lock()
		for _, req := range queue.requests {
			if req.Request.ID == requestID {
				req.EffectivePriority *= boost
				req.StarvationBoost = boost
			}
		}
		queue.mutex.Unlock()
	}
}

func (fs *FairScheduler) boostWRRRequest(requestID string, boost float64) {
	fs.wrr.mutex.Lock()
	defer fs.wrr.mutex.Unlock()

	for _, req := range fs.wrr.requests {
		if req.Request.ID == requestID {
			req.Weight = int(float64(req.Weight) * boost)
			req.StarvationBoost = boost
		}
	}
}

func (fs *FairScheduler) boostLotteryRequest(requestID string, boost float64) {
	fs.lottery.mutex.Lock()
	defer fs.lottery.mutex.Unlock()

	for _, req := range fs.lottery.requests {
		if req.Request.ID == requestID {
			user := req.Request.User.Username
			currentTickets := fs.lottery.userTickets[user]
			fs.lottery.userTickets[user] = int(float64(currentTickets) * boost)
			req.StarvationBoost = boost
		}
	}
}

func (fs *FairScheduler) boostCFSRequest(requestID string, boost float64) {
	fs.cfs.mutex.Lock()
	defer fs.cfs.mutex.Unlock()

	for _, req := range fs.cfs.requests {
		if req.Request.ID == requestID {
			// Reduce virtual runtime (making it more likely to be selected)
			req.VirtualRuntime /= boost
			req.StarvationBoost = boost
		}
	}
}

// Helper methods for enqueueing to different algorithms
func (fs *FairScheduler) enqueuePriorityBased(request *QueuedRequest) error {
	queue := fs.priorityQueue.queues[request.Priority]
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	queue.requests = append(queue.requests, request)
	return nil
}

func (fs *FairScheduler) enqueueWeightedRoundRobin(request *QueuedRequest) error {
	fs.wrr.mutex.Lock()
	defer fs.wrr.mutex.Unlock()

	fs.wrr.requests = append(fs.wrr.requests, request)
	return nil
}

func (fs *FairScheduler) enqueueLottery(request *QueuedRequest) error {
	fs.lottery.mutex.Lock()
	defer fs.lottery.mutex.Unlock()

	fs.lottery.requests = append(fs.lottery.requests, request)
	return nil
}

func (fs *FairScheduler) enqueueCompletelyFair(request *QueuedRequest) error {
	fs.cfs.mutex.Lock()
	defer fs.cfs.mutex.Unlock()

	user := request.Request.User.Username
	vruntime, exists := fs.cfs.vruntime[user]
	if !exists {
		// Initialize virtual runtime to current minimum
		minVruntime := math.Inf(1)
		for _, vr := range fs.cfs.vruntime {
			if vr < minVruntime {
				minVruntime = vr
			}
		}
		if minVruntime == math.Inf(1) {
			minVruntime = 0
		}
		vruntime = minVruntime
		fs.cfs.vruntime[user] = vruntime
	}

	vruntimeRequest := &VRuntimeRequest{
		QueuedRequest:  request,
		VirtualRuntime: vruntime,
		TimeSlice:      fs.cfs.timeSlice,
		LastScheduled:  time.Time{},
	}

	fs.cfs.requests = append(fs.cfs.requests, vruntimeRequest)
	return nil
}

// getWeight calculates the weight for a request
func (fs *FairScheduler) getWeight(request *enhanced.EnhancedLockRequest) int {
	baseWeight := fs.config.PriorityWeights[fmt.Sprintf("%v", request.Priority)]
	if baseWeight == 0 {
		baseWeight = int(request.Priority) + 1
	}

	// Add user weight
	userWeight := fs.config.UserWeights[request.User.Username]
	if userWeight == 0 {
		userWeight = 1
	}

	// Add project weight
	projectWeight := fs.config.ProjectWeights[request.Resource.Namespace]
	if projectWeight == 0 {
		projectWeight = 1
	}

	return baseWeight * userWeight * projectWeight
}

// updateMetrics updates scheduler metrics
func (fs *FairScheduler) updateMetrics(request *enhanced.EnhancedLockRequest) {
	fs.metrics.ProcessedRequests++
	fs.metrics.RequestsByPriority[request.Priority]++
	fs.metrics.RequestsByUser[request.User.Username]++
	fs.metrics.RequestsByProject[request.Resource.Namespace]++
	fs.metrics.LastUpdated = time.Now()

	// Calculate wait time
	if queuedAt, exists := fs.starvationTracker.requestAge[request.ID]; exists {
		waitTime := time.Since(queuedAt)

		// Update average wait time for priority
		current := fs.metrics.AverageWaitTime[request.Priority]
		if current == 0 {
			fs.metrics.AverageWaitTime[request.Priority] = waitTime
		} else {
			fs.metrics.AverageWaitTime[request.Priority] = (current + waitTime) / 2
		}

		// Update min/max
		if fs.metrics.MaxWaitTime == 0 || waitTime > fs.metrics.MaxWaitTime {
			fs.metrics.MaxWaitTime = waitTime
		}
		if fs.metrics.MinWaitTime == 0 || waitTime < fs.metrics.MinWaitTime {
			fs.metrics.MinWaitTime = waitTime
		}

		// Remove from starvation tracker
		delete(fs.starvationTracker.requestAge, request.ID)
	}
}

// trackRequest adds a request to starvation tracking
func (st *StarvationTracker) trackRequest(requestID string, queuedTime time.Time) {
	st.mutex.Lock()
	defer st.mutex.Unlock()

	st.requestAge[requestID] = queuedTime
}

// rebalance performs load rebalancing
func (lb *LoadBalancer) rebalance() {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	now := time.Now()
	if now.Sub(lb.windowStart) > lb.config.LoadBalanceWindow {
		// Reset load counters for new window
		for k := range lb.userLoad {
			lb.userLoad[k] = 0
		}
		for k := range lb.projectLoad {
			lb.projectLoad[k] = 0
		}
		for k := range lb.resourceLoad {
			lb.resourceLoad[k] = 0
		}
		lb.windowStart = now
	}
}