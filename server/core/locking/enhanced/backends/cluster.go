package backends

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/runatlantis/atlantis/server/core/locking/enhanced"
	"github.com/runatlantis/atlantis/server/logging"
)

// ClusterManager handles Redis cluster operations and coordination
type ClusterManager struct {
	client        redis.UniversalClient
	log           logging.SimpleLogging
	scriptMgr     *ScriptManager
	nodeID        string
	clusterConfig *ClusterConfig

	// Cluster state
	nodes         map[string]*ClusterNode
	nodesMutex    sync.RWMutex
	leaderID      string
	leaderMutex   sync.RWMutex

	// Health monitoring
	healthMonitor *HealthMonitor

	// Background tasks
	stopChan      chan struct{}
	running       bool
	mutex         sync.Mutex
}

// ClusterConfig configures cluster behavior
type ClusterConfig struct {
	NodeID               string        `mapstructure:"node_id"`
	HeartbeatInterval    time.Duration `mapstructure:"heartbeat_interval"`
	NodeTimeout          time.Duration `mapstructure:"node_timeout"`
	LeaderElectionTimeout time.Duration `mapstructure:"leader_election_timeout"`
	ConsensusThreshold   int           `mapstructure:"consensus_threshold"`
	EnableLeaderElection bool          `mapstructure:"enable_leader_election"`
	EnableConsensus      bool          `mapstructure:"enable_consensus"`
	MaxClusterSize       int           `mapstructure:"max_cluster_size"`
	ReplicationFactor    int           `mapstructure:"replication_factor"`
}

// DefaultClusterConfig returns default cluster configuration
func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		NodeID:               generateNodeID(),
		HeartbeatInterval:    30 * time.Second,
		NodeTimeout:          90 * time.Second,
		LeaderElectionTimeout: 10 * time.Second,
		ConsensusThreshold:   2, // Majority for 3-node cluster
		EnableLeaderElection: true,
		EnableConsensus:      true,
		MaxClusterSize:       7, // Odd number for consensus
		ReplicationFactor:    3,
	}
}

// ClusterNode represents a node in the Redis cluster
type ClusterNode struct {
	ID              string                 `json:"id"`
	Address         string                 `json:"address"`
	Status          ClusterNodeStatus      `json:"status"`
	LastHeartbeat   time.Time              `json:"last_heartbeat"`
	Version         string                 `json:"version"`
	Capabilities    []string               `json:"capabilities"`
	Metadata        map[string]interface{} `json:"metadata"`
	LockCount       int64                  `json:"lock_count"`
	QueueCount      int64                  `json:"queue_count"`
	IsLeader        bool                   `json:"is_leader"`
	LeaderPriority  int                    `json:"leader_priority"`
}

// ClusterNodeStatus represents the status of a cluster node
type ClusterNodeStatus string

const (
	NodeStatusActive      ClusterNodeStatus = "active"
	NodeStatusSuspected   ClusterNodeStatus = "suspected"
	NodeStatusFailed      ClusterNodeStatus = "failed"
	NodeStatusJoining     ClusterNodeStatus = "joining"
	NodeStatusLeaving     ClusterNodeStatus = "leaving"
	NodeStatusMaintenance ClusterNodeStatus = "maintenance"
)

// ClusterState represents the overall cluster state
type ClusterState struct {
	ClusterID     string                    `json:"cluster_id"`
	Nodes         map[string]*ClusterNode   `json:"nodes"`
	LeaderID      string                    `json:"leader_id"`
	LeaderTerm    int64                     `json:"leader_term"`
	LastUpdated   time.Time                 `json:"last_updated"`
	TotalLocks    int64                     `json:"total_locks"`
	HealthStatus  string                    `json:"health_status"`
	Configuration map[string]interface{}    `json:"configuration"`
}

// NewClusterManager creates a new cluster manager
func NewClusterManager(client redis.UniversalClient, config *ClusterConfig, log logging.SimpleLogging) *ClusterManager {
	if config == nil {
		config = DefaultClusterConfig()
	}

	cm := &ClusterManager{
		client:        client,
		log:           log,
		clusterConfig: config,
		nodeID:        config.NodeID,
		nodes:         make(map[string]*ClusterNode),
		stopChan:      make(chan struct{}),
		scriptMgr:     NewScriptManager(client),
	}

	// Initialize health monitor
	healthConfig := DefaultHealthConfig()
	healthConfig.CheckInterval = config.HeartbeatInterval / 2
	cm.healthMonitor = NewHealthMonitor(client, healthConfig, log)

	return cm
}

// Start initializes and starts the cluster manager
func (cm *ClusterManager) Start(ctx context.Context) error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if cm.running {
		return fmt.Errorf("cluster manager is already running")
	}

	cm.log.Info("Starting cluster manager for node %s", cm.nodeID)

	// Start health monitor
	if err := cm.healthMonitor.Start(ctx); err != nil {
		return fmt.Errorf("failed to start health monitor: %w", err)
	}

	// Load cluster scripts
	if err := cm.scriptMgr.LoadScripts(ctx); err != nil {
		cm.log.Warn("Failed to preload cluster scripts: %v", err)
	}

	// Join or create cluster
	if err := cm.joinCluster(ctx); err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	// Start background tasks
	go cm.heartbeatLoop(ctx)
	go cm.leaderElectionLoop(ctx)
	go cm.clusterMaintenanceLoop(ctx)

	cm.running = true
	cm.log.Info("Cluster manager started successfully")
	return nil
}

// Stop gracefully shuts down the cluster manager
func (cm *ClusterManager) Stop() error {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	if !cm.running {
		return nil
	}

	cm.log.Info("Stopping cluster manager for node %s", cm.nodeID)

	// Leave cluster
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := cm.leaveCluster(ctx); err != nil {
		cm.log.Warn("Failed to gracefully leave cluster: %v", err)
	}

	// Stop background tasks
	close(cm.stopChan)

	// Stop health monitor
	cm.healthMonitor.Stop()

	cm.running = false
	cm.log.Info("Cluster manager stopped")
	return nil
}

// joinCluster joins an existing cluster or creates a new one
func (cm *ClusterManager) joinCluster(ctx context.Context) error {
	node := &ClusterNode{
		ID:              cm.nodeID,
		Address:         cm.getNodeAddress(),
		Status:          NodeStatusJoining,
		LastHeartbeat:   time.Now(),
		Version:         cm.getNodeVersion(),
		Capabilities:    cm.getNodeCapabilities(),
		Metadata:        cm.getNodeMetadata(),
		LeaderPriority:  cm.calculateLeaderPriority(),
	}

	// Register node in cluster
	nodeData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal node data: %w", err)
	}

	clusterKey := "atlantis:cluster:nodes"
	nodeKey := fmt.Sprintf("atlantis:cluster:node:%s", cm.nodeID)

	// Use Lua script for atomic cluster join
	joinScript := `
		local clusterKey = KEYS[1]
		local nodeKey = KEYS[2]
		local nodeData = ARGV[1]
		local maxNodes = tonumber(ARGV[2])
		local nodeId = ARGV[3]

		-- Check if cluster is full
		local currentNodes = redis.call('SCARD', clusterKey)
		if currentNodes >= maxNodes then
			return {false, "cluster_full", currentNodes}
		end

		-- Check if node already exists
		local exists = redis.call('SISMEMBER', clusterKey, nodeId)
		if exists == 1 then
			-- Node rejoining, update data
			redis.call('SET', nodeKey, nodeData)
			redis.call('EXPIRE', nodeKey, 300)
			return {true, "rejoined", currentNodes}
		end

		-- Add new node
		redis.call('SADD', clusterKey, nodeId)
		redis.call('SET', nodeKey, nodeData)
		redis.call('EXPIRE', nodeKey, 300)

		-- Initialize node state
		redis.call('SET', nodeKey .. ':state', 'joining')
		redis.call('EXPIRE', nodeKey .. ':state', 300)

		-- Publish join event
		local joinEvent = cjson.encode({
			action = "node_joined",
			node_id = nodeId,
			timestamp = redis.call('TIME')[1]
		})
		redis.call('PUBLISH', 'atlantis:cluster:events', joinEvent)

		return {true, "joined", currentNodes + 1}
	`

	cm.scriptMgr.RegisterScript("cluster_join", joinScript)

	result, err := cm.scriptMgr.Execute(ctx, "cluster_join",
		[]string{clusterKey, nodeKey},
		string(nodeData), cm.clusterConfig.MaxClusterSize, cm.nodeID)
	if err != nil {
		return fmt.Errorf("failed to join cluster: %w", err)
	}

	resultSlice := result.([]interface{})
	success := resultSlice[0].(int64) == 1
	status := resultSlice[1].(string)
	nodeCount := resultSlice[2].(int64)

	if !success {
		return fmt.Errorf("failed to join cluster: %s", status)
	}

	cm.log.Info("Node %s %s cluster with %d total nodes", cm.nodeID, status, nodeCount)

	// Update node status to active
	node.Status = NodeStatusActive
	nodeData, _ = json.Marshal(node)
	cm.client.Set(ctx, nodeKey, string(nodeData), 5*time.Minute)

	// Sync cluster state
	return cm.syncClusterState(ctx)
}

// leaveCluster gracefully leaves the cluster
func (cm *ClusterManager) leaveCluster(ctx context.Context) error {
	clusterKey := "atlantis:cluster:nodes"
	nodeKey := fmt.Sprintf("atlantis:cluster:node:%s", cm.nodeID)

	// Use Lua script for atomic cluster leave
	leaveScript := `
		local clusterKey = KEYS[1]
		local nodeKey = KEYS[2]
		local nodeId = ARGV[1]

		-- Check if node exists
		local exists = redis.call('SISMEMBER', clusterKey, nodeId)
		if exists == 0 then
			return {false, "not_member"}
		end

		-- Remove from cluster
		redis.call('SREM', clusterKey, nodeId)
		redis.call('DEL', nodeKey)
		redis.call('DEL', nodeKey .. ':state')
		redis.call('DEL', nodeKey .. ':locks')

		-- Publish leave event
		local leaveEvent = cjson.encode({
			action = "node_left",
			node_id = nodeId,
			timestamp = redis.call('TIME')[1]
		})
		redis.call('PUBLISH', 'atlantis:cluster:events', leaveEvent)

		local remainingNodes = redis.call('SCARD', clusterKey)
		return {true, "left", remainingNodes}
	`

	cm.scriptMgr.RegisterScript("cluster_leave", leaveScript)

	result, err := cm.scriptMgr.Execute(ctx, "cluster_leave",
		[]string{clusterKey, nodeKey}, cm.nodeID)
	if err != nil {
		return fmt.Errorf("failed to leave cluster: %w", err)
	}

	resultSlice := result.([]interface{})
	success := resultSlice[0].(int64) == 1
	if success {
		remainingNodes := resultSlice[2].(int64)
		cm.log.Info("Node %s left cluster, %d nodes remaining", cm.nodeID, remainingNodes)
	}

	return nil
}

// syncClusterState synchronizes the local cluster state with Redis
func (cm *ClusterManager) syncClusterState(ctx context.Context) error {
	clusterKey := "atlantis:cluster:nodes"

	// Get all cluster nodes
	nodeIDs, err := cm.client.SMembers(ctx, clusterKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get cluster nodes: %w", err)
	}

	cm.nodesMutex.Lock()
	defer cm.nodesMutex.Unlock()

	// Clear current state
	cm.nodes = make(map[string]*ClusterNode)

	// Fetch each node's data
	for _, nodeID := range nodeIDs {
		nodeKey := fmt.Sprintf("atlantis:cluster:node:%s", nodeID)
		nodeData, err := cm.client.Get(ctx, nodeKey).Result()
		if err != nil {
			if err == redis.Nil {
				// Node data expired, remove from cluster
				cm.client.SRem(ctx, clusterKey, nodeID)
				continue
			}
			cm.log.Warn("Failed to get data for node %s: %v", nodeID, err)
			continue
		}

		var node ClusterNode
		if err := json.Unmarshal([]byte(nodeData), &node); err != nil {
			cm.log.Warn("Failed to unmarshal data for node %s: %v", nodeID, err)
			continue
		}

		cm.nodes[nodeID] = &node
	}

	cm.log.Debug("Synced cluster state: %d nodes", len(cm.nodes))
	return nil
}

// heartbeatLoop sends periodic heartbeats to maintain cluster membership
func (cm *ClusterManager) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(cm.clusterConfig.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			if err := cm.sendHeartbeat(ctx); err != nil {
				cm.log.Warn("Failed to send heartbeat: %v", err)
			}
		}
	}
}

// sendHeartbeat sends a heartbeat to update node status
func (cm *ClusterManager) sendHeartbeat(ctx context.Context) error {
	nodeKey := fmt.Sprintf("atlantis:cluster:node:%s", cm.nodeID)

	// Update node data with current status
	node := &ClusterNode{
		ID:             cm.nodeID,
		Address:        cm.getNodeAddress(),
		Status:         NodeStatusActive,
		LastHeartbeat:  time.Now(),
		Version:        cm.getNodeVersion(),
		Capabilities:   cm.getNodeCapabilities(),
		Metadata:       cm.getNodeMetadata(),
		LockCount:      cm.getLockCount(ctx),
		QueueCount:     cm.getQueueCount(ctx),
		IsLeader:       cm.isLeader(),
		LeaderPriority: cm.calculateLeaderPriority(),
	}

	nodeData, err := json.Marshal(node)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat data: %w", err)
	}

	// Update node data with TTL
	if err := cm.client.SetEx(ctx, nodeKey, string(nodeData), 2*cm.clusterConfig.HeartbeatInterval).Err(); err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	// Sync cluster state periodically
	if time.Now().Unix()%60 == 0 { // Every minute
		if err := cm.syncClusterState(ctx); err != nil {
			cm.log.Warn("Failed to sync cluster state: %v", err)
		}
	}

	return nil
}

// leaderElectionLoop handles leader election
func (cm *ClusterManager) leaderElectionLoop(ctx context.Context) {
	if !cm.clusterConfig.EnableLeaderElection {
		return
	}

	ticker := time.NewTicker(cm.clusterConfig.LeaderElectionTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			if err := cm.electLeader(ctx); err != nil {
				cm.log.Warn("Leader election failed: %v", err)
			}
		}
	}
}

// electLeader performs leader election using Redis distributed locking
func (cm *ClusterManager) electLeader(ctx context.Context) error {
	leaderKey := "atlantis:cluster:leader"
	electionKey := "atlantis:cluster:election"

	// Check if current leader is still alive
	currentLeader, err := cm.client.Get(ctx, leaderKey).Result()
	if err == nil && currentLeader != "" {
		// Verify leader is still active
		leaderNodeKey := fmt.Sprintf("atlantis:cluster:node:%s", currentLeader)
		_, err := cm.client.Get(ctx, leaderNodeKey).Result()
		if err == nil {
			// Current leader is still alive
			cm.leaderMutex.Lock()
			cm.leaderID = currentLeader
			cm.leaderMutex.Unlock()
			return nil
		}
	}

	// No leader or leader is dead, start election
	electionScript := `
		local leaderKey = KEYS[1]
		local electionKey = KEYS[2]
		local nodeId = ARGV[1]
		local priority = tonumber(ARGV[2])
		local ttl = tonumber(ARGV[3])

		-- Check if election is in progress
		local election = redis.call('GET', electionKey)
		if election then
			return {false, "election_in_progress", election}
		end

		-- Start election
		redis.call('SETEX', electionKey, ttl, nodeId)

		-- Try to become leader
		local result = redis.call('SET', leaderKey, nodeId, 'NX', 'EX', ttl * 2)
		if result then
			-- Election successful
			redis.call('DEL', electionKey)

			-- Publish leader election event
			local electionEvent = cjson.encode({
				action = "leader_elected",
				leader_id = nodeId,
				priority = priority,
				timestamp = redis.call('TIME')[1]
			})
			redis.call('PUBLISH', 'atlantis:cluster:events', electionEvent)

			return {true, "elected", nodeId}
		else
			-- Election failed
			redis.call('DEL', electionKey)
			local currentLeader = redis.call('GET', leaderKey)
			return {false, "failed", currentLeader}
		end
	`

	cm.scriptMgr.RegisterScript("leader_election", electionScript)

	priority := cm.calculateLeaderPriority()
	ttl := int64(cm.clusterConfig.LeaderElectionTimeout.Seconds())

	result, err := cm.scriptMgr.Execute(ctx, "leader_election",
		[]string{leaderKey, electionKey},
		cm.nodeID, priority, ttl)
	if err != nil {
		return fmt.Errorf("leader election script failed: %w", err)
	}

	resultSlice := result.([]interface{})
	success := resultSlice[0].(int64) == 1
	status := resultSlice[1].(string)
	leaderID := resultSlice[2].(string)

	cm.leaderMutex.Lock()
	cm.leaderID = leaderID
	cm.leaderMutex.Unlock()

	if success {
		cm.log.Info("Node %s elected as cluster leader", cm.nodeID)
	} else {
		cm.log.Debug("Leader election %s, current leader: %s", status, leaderID)
	}

	return nil
}

// clusterMaintenanceLoop performs periodic cluster maintenance
func (cm *ClusterManager) clusterMaintenanceLoop(ctx context.Context) {
	ticker := time.NewTicker(cm.clusterConfig.NodeTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopChan:
			return
		case <-ticker.C:
			cm.performClusterMaintenance(ctx)
		}
	}
}

// performClusterMaintenance cleans up failed nodes and maintains cluster health
func (cm *ClusterManager) performClusterMaintenance(ctx context.Context) {
	clusterKey := "atlantis:cluster:nodes"
	nodeIDs, err := cm.client.SMembers(ctx, clusterKey).Result()
	if err != nil {
		cm.log.Warn("Failed to get cluster nodes for maintenance: %v", err)
		return
	}

	now := time.Now()
	failedNodes := make([]string, 0)

	for _, nodeID := range nodeIDs {
		nodeKey := fmt.Sprintf("atlantis:cluster:node:%s", nodeID)
		nodeData, err := cm.client.Get(ctx, nodeKey).Result()
		if err != nil {
			if err == redis.Nil {
				// Node data expired
				failedNodes = append(failedNodes, nodeID)
				continue
			}
			cm.log.Warn("Failed to check node %s: %v", nodeID, err)
			continue
		}

		var node ClusterNode
		if err := json.Unmarshal([]byte(nodeData), &node); err != nil {
			cm.log.Warn("Failed to unmarshal node %s data: %v", nodeID, err)
			continue
		}

		// Check if node has timed out
		if now.Sub(node.LastHeartbeat) > cm.clusterConfig.NodeTimeout {
			failedNodes = append(failedNodes, nodeID)
		}
	}

	// Remove failed nodes
	for _, nodeID := range failedNodes {
		cm.log.Info("Removing failed node from cluster: %s", nodeID)
		cm.client.SRem(ctx, clusterKey, nodeID)
		cm.client.Del(ctx, fmt.Sprintf("atlantis:cluster:node:%s", nodeID))

		// Publish node failure event
		failureEvent := map[string]interface{}{
			"action":    "node_failed",
			"node_id":   nodeID,
			"timestamp": now.Unix(),
		}
		eventData, _ := json.Marshal(failureEvent)
		cm.client.Publish(ctx, "atlantis:cluster:events", string(eventData))
	}

	if len(failedNodes) > 0 {
		// Trigger leader election if leader failed
		cm.leaderMutex.RLock()
		leaderFailed := false
		for _, nodeID := range failedNodes {
			if nodeID == cm.leaderID {
				leaderFailed = true
				break
			}
		}
		cm.leaderMutex.RUnlock()

		if leaderFailed {
			cm.log.Info("Leader node failed, triggering re-election")
			cm.client.Del(ctx, "atlantis:cluster:leader")
		}

		// Sync cluster state after cleanup
		cm.syncClusterState(ctx)
	}
}

// Helper methods

func (cm *ClusterManager) getNodeAddress() string {
	// This would typically return the actual network address
	return fmt.Sprintf("atlantis-node-%s", cm.nodeID)
}

func (cm *ClusterManager) getNodeVersion() string {
	return "enhanced-v1.0.0"
}

func (cm *ClusterManager) getNodeCapabilities() []string {
	return []string{"redis", "clustering", "priority_queue", "health_monitoring", "lua_scripts"}
}

func (cm *ClusterManager) getNodeMetadata() map[string]interface{} {
	return map[string]interface{}{
		"startup_time": time.Now().Unix(),
		"cluster_config": map[string]interface{}{
			"max_cluster_size":    cm.clusterConfig.MaxClusterSize,
			"replication_factor":  cm.clusterConfig.ReplicationFactor,
			"consensus_threshold": cm.clusterConfig.ConsensusThreshold,
		},
	}
}

func (cm *ClusterManager) calculateLeaderPriority() int {
	// Higher priority = better leader candidate
	// Factors: health score, lock count (lower is better), uptime
	healthScore := 100
	if cm.healthMonitor != nil {
		metrics := cm.healthMonitor.GetMetrics()
		healthScore = int(float64(metrics.SuccessfulChecks) / float64(metrics.TotalChecks) * 100)
	}

	// Base priority on health, lower lock count is better for leader
	lockCount := cm.getLockCount(context.Background())
	priority := healthScore - int(lockCount)

	if priority < 0 {
		priority = 0
	}
	if priority > 100 {
		priority = 100
	}

	return priority
}

func (cm *ClusterManager) getLockCount(ctx context.Context) int64 {
	nodeLocksKey := fmt.Sprintf("atlantis:node:%s:locks", cm.nodeID)
	count, err := cm.client.SCard(ctx, nodeLocksKey).Result()
	if err != nil {
		return 0
	}
	return count
}

func (cm *ClusterManager) getQueueCount(ctx context.Context) int64 {
	// Count all queue entries for this node
	pattern := fmt.Sprintf("atlantis:enhanced:lock:*:queue")
	keys, err := cm.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0
	}

	var total int64
	for _, key := range keys {
		count, err := cm.client.ZCard(ctx, key).Result()
		if err == nil {
			total += count
		}
	}

	return total
}

func (cm *ClusterManager) isLeader() bool {
	cm.leaderMutex.RLock()
	defer cm.leaderMutex.RUnlock()
	return cm.leaderID == cm.nodeID
}

// Public API methods

// GetClusterState returns the current cluster state
func (cm *ClusterManager) GetClusterState(ctx context.Context) (*ClusterState, error) {
	if err := cm.syncClusterState(ctx); err != nil {
		return nil, err
	}

	cm.nodesMutex.RLock()
	defer cm.nodesMutex.RUnlock()

	cm.leaderMutex.RLock()
	leaderID := cm.leaderID
	cm.leaderMutex.RUnlock()

	var totalLocks int64
	for _, node := range cm.nodes {
		totalLocks += node.LockCount
	}

	return &ClusterState{
		ClusterID:    "atlantis-enhanced",
		Nodes:        cm.nodes,
		LeaderID:     leaderID,
		LastUpdated:  time.Now(),
		TotalLocks:   totalLocks,
		HealthStatus: cm.calculateClusterHealth(),
		Configuration: map[string]interface{}{
			"max_cluster_size":    cm.clusterConfig.MaxClusterSize,
			"consensus_threshold": cm.clusterConfig.ConsensusThreshold,
			"replication_factor":  cm.clusterConfig.ReplicationFactor,
		},
	}, nil
}

// GetClusterNodes returns all active cluster nodes
func (cm *ClusterManager) GetClusterNodes() map[string]*ClusterNode {
	cm.nodesMutex.RLock()
	defer cm.nodesMutex.RUnlock()

	// Return copy to prevent external modification
	nodes := make(map[string]*ClusterNode)
	for id, node := range cm.nodes {
		nodeCopy := *node
		nodes[id] = &nodeCopy
	}

	return nodes
}

// IsLeader returns true if this node is the cluster leader
func (cm *ClusterManager) IsLeader() bool {
	return cm.isLeader()
}

// GetLeaderID returns the current cluster leader ID
func (cm *ClusterManager) GetLeaderID() string {
	cm.leaderMutex.RLock()
	defer cm.leaderMutex.RUnlock()
	return cm.leaderID
}

// calculateClusterHealth determines overall cluster health
func (cm *ClusterManager) calculateClusterHealth() string {
	cm.nodesMutex.RLock()
	defer cm.nodesMutex.RUnlock()

	totalNodes := len(cm.nodes)
	if totalNodes == 0 {
		return "critical"
	}

	activeNodes := 0
	for _, node := range cm.nodes {
		if node.Status == NodeStatusActive {
			activeNodes++
		}
	}

	healthRatio := float64(activeNodes) / float64(totalNodes)
	switch {
	case healthRatio >= 0.9:
		return "healthy"
	case healthRatio >= 0.7:
		return "degraded"
	case healthRatio >= 0.5:
		return "warning"
	default:
		return "critical"
	}
}

// generateNodeID generates a unique node identifier
func generateNodeID() string {
	return fmt.Sprintf("node-%d", time.Now().UnixNano())
}