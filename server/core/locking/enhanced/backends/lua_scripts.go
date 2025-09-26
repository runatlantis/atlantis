package backends

import (
	"context"
	"crypto/sha1"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

// LuaScript represents a reusable Lua script with caching
type LuaScript struct {
	Source string
	SHA    string
	script *redis.Script
}

// ScriptManager manages and caches Lua scripts for atomic Redis operations
type ScriptManager struct {
	scripts map[string]*LuaScript
	client  redis.UniversalClient
	mutex   sync.RWMutex
}

// NewScriptManager creates a new Lua script manager
func NewScriptManager(client redis.UniversalClient) *ScriptManager {
	sm := &ScriptManager{
		scripts: make(map[string]*LuaScript),
		client:  client,
	}
	sm.initializeScripts()
	return sm
}

// initializeScripts loads all predefined scripts
func (sm *ScriptManager) initializeScripts() {
	scripts := map[string]string{
		"acquire_lock":           acquireLockScript,
		"release_lock":           releaseLockScript,
		"refresh_lock":           refreshLockScript,
		"transfer_lock":          transferLockScript,
		"cleanup_expired":        cleanupExpiredScript,
		"queue_operations":       queueOperationsScript,
		"distributed_acquire":    distributedAcquireScript,
		"health_check":           healthCheckScript,
		"batch_operations":       batchOperationsScript,
		"priority_queue_ops":     priorityQueueOpsScript,
	}

	for name, source := range scripts {
		sm.RegisterScript(name, source)
	}
}

// RegisterScript registers a new Lua script
func (sm *ScriptManager) RegisterScript(name, source string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Generate SHA1 hash for the script
	h := sha1.New()
	h.Write([]byte(source))
	sha := fmt.Sprintf("%x", h.Sum(nil))

	script := &LuaScript{
		Source: source,
		SHA:    sha,
		script: redis.NewScript(source),
	}

	sm.scripts[name] = script
}

// Execute runs a registered script
func (sm *ScriptManager) Execute(ctx context.Context, scriptName string, keys []string, args ...interface{}) (interface{}, error) {
	sm.mutex.RLock()
	script, exists := sm.scripts[scriptName]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	return script.script.Run(ctx, sm.client, keys, args...).Result()
}

// ExecuteWithFallback tries EVALSHA first, then falls back to EVAL
func (sm *ScriptManager) ExecuteWithFallback(ctx context.Context, scriptName string, keys []string, args ...interface{}) (interface{}, error) {
	sm.mutex.RLock()
	script, exists := sm.scripts[scriptName]
	sm.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	// Try EVALSHA first (more efficient)
	result, err := sm.client.EvalSha(ctx, script.SHA, keys, args...).Result()
	if err != nil {
		if err.Error() == "NOSCRIPT No matching script. Please use EVAL." {
			// Script not cached on server, use EVAL
			return sm.client.Eval(ctx, script.Source, keys, args...).Result()
		}
		return nil, err
	}

	return result, nil
}

// LoadScripts preloads all scripts to Redis server for better performance
func (sm *ScriptManager) LoadScripts(ctx context.Context) error {
	sm.mutex.RLock()
	scripts := make([]*LuaScript, 0, len(sm.scripts))
	for _, script := range sm.scripts {
		scripts = append(scripts, script)
	}
	sm.mutex.RUnlock()

	for _, script := range scripts {
		_, err := sm.client.ScriptLoad(ctx, script.Source).Result()
		if err != nil {
			return fmt.Errorf("failed to load script %s: %w", script.SHA, err)
		}
	}

	return nil
}

// Lua script definitions with advanced atomic operations

const acquireLockScript = `
-- Enhanced atomic lock acquisition with clustering support
-- KEYS[1]: lock key
-- KEYS[2]: queue key
-- KEYS[3]: cluster node key
-- ARGV[1]: lock data (JSON)
-- ARGV[2]: TTL in seconds
-- ARGV[3]: priority (0-3)
-- ARGV[4]: queue enabled ("true"/"false")
-- ARGV[5]: cluster mode ("true"/"false")
-- ARGV[6]: node ID
-- ARGV[7]: max queue size

local lockKey = KEYS[1]
local queueKey = KEYS[2]
local nodeKey = KEYS[3]
local lockData = ARGV[1]
local ttl = tonumber(ARGV[2])
local priority = tonumber(ARGV[3])
local queueEnabled = ARGV[4] == "true"
local clusterMode = ARGV[5] == "true"
local nodeId = ARGV[6]
local maxQueueSize = tonumber(ARGV[7]) or 1000

-- Check if lock already exists
local existing = redis.call('GET', lockKey)
if existing then
    if queueEnabled then
        -- Check queue size limit
        local queueSize = redis.call('ZCARD', queueKey)
        if queueSize >= maxQueueSize then
            return {false, "queue_full", queueSize}
        end

        -- Add to priority queue with timestamp for FIFO within same priority
        local score = (4 - priority) * 1000000 + redis.call('TIME')[1]
        redis.call('ZADD', queueKey, score, lockData)

        -- Set queue TTL to prevent memory leaks
        redis.call('EXPIRE', queueKey, 3600)

        return {false, "queued", queueSize + 1}
    end
    return {false, "exists", 0}
end

-- Cluster coordination check
if clusterMode then
    -- Check if another node in cluster holds conflicting lock
    local clusterNodes = redis.call('SMEMBERS', 'atlantis:cluster:nodes')
    for i = 1, #clusterNodes do
        local nodeCheckKey = 'atlantis:node:' .. clusterNodes[i] .. ':locks'
        local nodeLocks = redis.call('SISMEMBER', nodeCheckKey, lockKey)
        if nodeLocks == 1 and clusterNodes[i] ~= nodeId then
            return {false, "cluster_conflict", clusterNodes[i]}
        end
    end

    -- Register lock with this node
    redis.call('SADD', 'atlantis:node:' .. nodeId .. ':locks', lockKey)
    redis.call('EXPIRE', 'atlantis:node:' .. nodeId .. ':locks', 7200)
end

-- Acquire the lock
if ttl > 0 then
    redis.call('SETEX', lockKey, ttl, lockData)
else
    redis.call('SET', lockKey, lockData)
end

-- Update cluster state
if clusterMode then
    redis.call('HSET', 'atlantis:cluster:state', nodeId, redis.call('TIME')[1])
    redis.call('EXPIRE', 'atlantis:cluster:state', 300)
end

-- Publish lock acquired event with metadata
local eventData = cjson.encode({
    action = "acquired",
    key = lockKey,
    node = nodeId,
    priority = priority,
    timestamp = redis.call('TIME')[1]
})
redis.call('PUBLISH', 'atlantis:lock:events', eventData)

return {true, "acquired", 0}
`

const releaseLockScript = `
-- Enhanced atomic lock release with queue processing and cluster coordination
-- KEYS[1]: lock key
-- KEYS[2]: queue key
-- KEYS[3]: node key
-- ARGV[1]: lock ID to verify ownership
-- ARGV[2]: cluster mode ("true"/"false")
-- ARGV[3]: node ID

local lockKey = KEYS[1]
local queueKey = KEYS[2]
local nodeKey = KEYS[3]
local lockId = ARGV[1]
local clusterMode = ARGV[2] == "true"
local nodeId = ARGV[3]

-- Get current lock and verify ownership
local currentLock = redis.call('GET', lockKey)
if not currentLock then
    return {false, "not_found", nil}
end

-- Parse lock data to verify ownership
local lockData = cjson.decode(currentLock)
if lockData.id ~= lockId then
    return {false, "not_owner", lockData.owner}
end

-- Release the lock
redis.call('DEL', lockKey)

-- Remove from cluster node tracking
if clusterMode then
    redis.call('SREM', 'atlantis:node:' .. nodeId .. ':locks', lockKey)
end

-- Process queue if enabled
local nextLock = nil
local queueProcessed = false

-- Get next highest priority request from queue
local queued = redis.call('ZRANGE', queueKey, 0, 0, 'WITHSCORES')
if #queued > 0 then
    local nextLockData = queued[1]
    local nextScore = queued[2]

    -- Remove from queue
    redis.call('ZREM', queueKey, nextLockData)

    -- Parse and update the queued lock
    nextLock = cjson.decode(nextLockData)
    nextLock.state = "acquired"
    nextLock.acquired_at = redis.call('TIME')[1]

    -- Acquire lock for queued request
    local ttl = 0
    if nextLock.expires_at then
        ttl = nextLock.expires_at - nextLock.acquired_at
    end

    if ttl > 0 then
        redis.call('SETEX', lockKey, ttl, cjson.encode(nextLock))
    else
        redis.call('SET', lockKey, cjson.encode(nextLock))
    end

    -- Update cluster tracking for new lock
    if clusterMode then
        redis.call('SADD', 'atlantis:node:' .. nodeId .. ':locks', lockKey)
    end

    queueProcessed = true

    -- Publish lock transferred event
    local transferEvent = cjson.encode({
        action = "transferred",
        key = lockKey,
        from = lockData.owner,
        to = nextLock.owner,
        node = nodeId,
        timestamp = redis.call('TIME')[1]
    })
    redis.call('PUBLISH', 'atlantis:lock:events', transferEvent)
else
    -- Publish lock released event
    local releaseEvent = cjson.encode({
        action = "released",
        key = lockKey,
        owner = lockData.owner,
        node = nodeId,
        timestamp = redis.call('TIME')[1]
    })
    redis.call('PUBLISH', 'atlantis:lock:events', releaseEvent)
end

-- Clean up empty queue
if redis.call('ZCARD', queueKey) == 0 then
    redis.call('DEL', queueKey)
end

-- Update cluster state
if clusterMode then
    redis.call('HSET', 'atlantis:cluster:state', nodeId, redis.call('TIME')[1])
end

return {true, queueProcessed and "transferred" or "released", nextLock}
`

const refreshLockScript = `
-- Atomic lock refresh with TTL extension
-- KEYS[1]: lock key
-- ARGV[1]: lock ID to verify ownership
-- ARGV[2]: extension in seconds
-- ARGV[3]: max TTL limit

local lockKey = KEYS[1]
local lockId = ARGV[1]
local extension = tonumber(ARGV[2])
local maxTTL = tonumber(ARGV[3]) or 7200 -- 2 hours default max

local currentLock = redis.call('GET', lockKey)
if not currentLock then
    return {false, "not_found"}
end

local lockData = cjson.decode(currentLock)
if lockData.id ~= lockId then
    return {false, "not_owner"}
end

-- Calculate new TTL
local currentTTL = redis.call('TTL', lockKey)
local newTTL = math.min(currentTTL + extension, maxTTL)

-- Update expiration in lock data
local currentTime = redis.call('TIME')[1]
lockData.expires_at = currentTime + newTTL
lockData.version = (lockData.version or 1) + 1

-- Update the lock with new TTL
redis.call('SETEX', lockKey, newTTL, cjson.encode(lockData))

-- Publish refresh event
local refreshEvent = cjson.encode({
    action = "refreshed",
    key = lockKey,
    owner = lockData.owner,
    new_ttl = newTTL,
    timestamp = currentTime
})
redis.call('PUBLISH', 'atlantis:lock:events', refreshEvent)

return {true, "refreshed", newTTL}
`

const transferLockScript = `
-- Atomic lock ownership transfer
-- KEYS[1]: lock key
-- ARGV[1]: current lock ID
-- ARGV[2]: current owner
-- ARGV[3]: new owner
-- ARGV[4]: cluster mode

local lockKey = KEYS[1]
local lockId = ARGV[1]
local currentOwner = ARGV[2]
local newOwner = ARGV[3]
local clusterMode = ARGV[4] == "true"

local currentLock = redis.call('GET', lockKey)
if not currentLock then
    return {false, "not_found"}
end

local lockData = cjson.decode(currentLock)
if lockData.id ~= lockId or lockData.owner ~= currentOwner then
    return {false, "not_owner"}
end

-- Transfer ownership
lockData.owner = newOwner
lockData.version = (lockData.version or 1) + 1
lockData.transferred_at = redis.call('TIME')[1]

-- Preserve existing TTL
local ttl = redis.call('TTL', lockKey)
if ttl > 0 then
    redis.call('SETEX', lockKey, ttl, cjson.encode(lockData))
else
    redis.call('SET', lockKey, cjson.encode(lockData))
end

-- Publish transfer event
local transferEvent = cjson.encode({
    action = "ownership_transferred",
    key = lockKey,
    from = currentOwner,
    to = newOwner,
    timestamp = lockData.transferred_at
})
redis.call('PUBLISH', 'atlantis:lock:events', transferEvent)

return {true, "transferred", newOwner}
`

const cleanupExpiredScript = `
-- Batch cleanup of expired locks with cluster awareness
-- KEYS[1]: lock pattern
-- ARGV[1]: current timestamp
-- ARGV[2]: batch size
-- ARGV[3]: cluster mode
-- ARGV[4]: node ID

local pattern = KEYS[1]
local currentTime = tonumber(ARGV[1])
local batchSize = tonumber(ARGV[2]) or 100
local clusterMode = ARGV[3] == "true"
local nodeId = ARGV[4]

local cursor = "0"
local cleaned = 0
local keys = {}

repeat
    local result = redis.call('SCAN', cursor, 'MATCH', pattern, 'COUNT', batchSize)
    cursor = result[1]
    local foundKeys = result[2]

    for i = 1, #foundKeys do
        table.insert(keys, foundKeys[i])
    end
until cursor == "0" or #keys >= batchSize

-- Process found keys
for i = 1, #keys do
    local key = keys[i]
    local lockData = redis.call('GET', key)

    if lockData then
        local lock = cjson.decode(lockData)
        local isExpired = false

        -- Check TTL expiration
        local ttl = redis.call('TTL', key)
        if ttl == -1 or ttl == -2 then
            isExpired = true
        elseif lock.expires_at and currentTime > lock.expires_at then
            isExpired = true
        end

        if isExpired then
            redis.call('DEL', key)
            cleaned = cleaned + 1

            -- Remove from cluster tracking
            if clusterMode and nodeId then
                redis.call('SREM', 'atlantis:node:' .. nodeId .. ':locks', key)
            end

            -- Publish cleanup event
            local cleanupEvent = cjson.encode({
                action = "expired_cleanup",
                key = key,
                owner = lock.owner,
                node = nodeId,
                timestamp = currentTime
            })
            redis.call('PUBLISH', 'atlantis:lock:events', cleanupEvent)
        end
    end
end

return {cleaned, #keys, cursor ~= "0"}
`

const queueOperationsScript = `
-- Advanced queue operations with priority handling
-- ARGV[1]: operation type ("push", "pop", "peek", "remove", "size", "cleanup")
-- KEYS[1]: queue key
-- Additional ARGV for operation-specific parameters

local operation = ARGV[1]
local queueKey = KEYS[1]

if operation == "push" then
    local requestData = ARGV[2]
    local priority = tonumber(ARGV[3])
    local maxSize = tonumber(ARGV[4]) or 1000

    local currentSize = redis.call('ZCARD', queueKey)
    if currentSize >= maxSize then
        return {false, "queue_full", currentSize}
    end

    local score = (4 - priority) * 1000000 + redis.call('TIME')[1]
    redis.call('ZADD', queueKey, score, requestData)
    redis.call('EXPIRE', queueKey, 3600)

    return {true, "queued", currentSize + 1}

elseif operation == "pop" then
    local items = redis.call('ZRANGE', queueKey, 0, 0)
    if #items == 0 then
        return {false, "empty", nil}
    end

    local item = items[1]
    redis.call('ZREM', queueKey, item)

    -- Clean up empty queue
    if redis.call('ZCARD', queueKey) == 0 then
        redis.call('DEL', queueKey)
    end

    return {true, "popped", item}

elseif operation == "peek" then
    local items = redis.call('ZRANGE', queueKey, 0, 0)
    if #items == 0 then
        return {false, "empty", nil}
    end
    return {true, "peeked", items[1]}

elseif operation == "remove" then
    local requestData = ARGV[2]
    local removed = redis.call('ZREM', queueKey, requestData)

    if redis.call('ZCARD', queueKey) == 0 then
        redis.call('DEL', queueKey)
    end

    return {removed > 0, removed > 0 and "removed" or "not_found", removed}

elseif operation == "size" then
    local size = redis.call('ZCARD', queueKey)
    return {true, "size", size}

elseif operation == "cleanup" then
    local maxAge = tonumber(ARGV[2]) or 3600
    local currentTime = redis.call('TIME')[1]
    local cutoffScore = currentTime - maxAge

    local removed = redis.call('ZREMRANGEBYSCORE', queueKey, 0, cutoffScore)

    if redis.call('ZCARD', queueKey) == 0 then
        redis.call('DEL', queueKey)
    end

    return {true, "cleaned", removed}
end

return {false, "unknown_operation", nil}
`

const distributedAcquireScript = `
-- Distributed lock acquisition across Redis cluster
-- KEYS[1]: lock key
-- KEYS[2]: cluster consensus key
-- ARGV[1]: lock data
-- ARGV[2]: node ID
-- ARGV[3]: cluster size
-- ARGV[4]: consensus threshold (majority)

local lockKey = KEYS[1]
local consensusKey = KEYS[2]
local lockData = ARGV[1]
local nodeId = ARGV[2]
local clusterSize = tonumber(ARGV[3])
local threshold = tonumber(ARGV[4])

-- Check if lock exists locally
local existing = redis.call('GET', lockKey)
if existing then
    return {false, "exists_locally", nil}
end

-- Register vote for this lock acquisition
local voteKey = consensusKey .. ':votes'
redis.call('HSET', voteKey, nodeId, lockData)
redis.call('EXPIRE', voteKey, 30)

-- Count votes
local votes = redis.call('HLEN', voteKey)
if votes >= threshold then
    -- Consensus reached, acquire lock
    redis.call('SET', lockKey, lockData)
    redis.call('DEL', voteKey)

    -- Mark as consensus leader
    redis.call('SETEX', consensusKey .. ':leader', 60, nodeId)

    return {true, "consensus_acquired", votes}
else
    -- Wait for more votes
    return {false, "waiting_consensus", votes}
end
`

const healthCheckScript = `
-- Redis backend health check with performance metrics
-- ARGV[1]: check type ("basic", "extended", "performance")

local checkType = ARGV[1] or "basic"
local result = {}

-- Basic connectivity test
result.ping = "pong"
result.timestamp = redis.call('TIME')[1]

if checkType == "basic" then
    return cjson.encode(result)
end

-- Extended checks
if checkType == "extended" or checkType == "performance" then
    -- Memory info
    local memory = redis.call('MEMORY', 'USAGE', 'atlantis:health:check')
    result.memory_usage = memory

    -- Get database size
    result.dbsize = redis.call('DBSIZE')

    -- Check replication status
    local info = redis.call('INFO', 'replication')
    result.replication_info = info

    -- Check if clustering is enabled
    local cluster_info = redis.call('INFO', 'cluster')
    result.cluster_info = cluster_info
end

if checkType == "performance" then
    -- Performance benchmarks
    local start_time = redis.call('TIME')

    -- Test set/get performance
    for i = 1, 10 do
        redis.call('SET', 'atlantis:perf:test:' .. i, 'test_value')
        redis.call('GET', 'atlantis:perf:test:' .. i)
        redis.call('DEL', 'atlantis:perf:test:' .. i)
    end

    local end_time = redis.call('TIME')
    local duration = (end_time[1] - start_time[1]) * 1000000 + (end_time[2] - start_time[2])
    result.performance_test_microseconds = duration

    -- Get slowlog for analysis
    local slowlog = redis.call('SLOWLOG', 'GET', 5)
    result.recent_slow_queries = #slowlog
end

return cjson.encode(result)
`

const batchOperationsScript = `
-- Batch operations for improved performance
-- ARGV[1]: operation type ("multi_get", "multi_set", "multi_del", "pipeline")
-- KEYS: variable number based on operation

local operation = ARGV[1]

if operation == "multi_get" then
    local results = {}
    for i = 1, #KEYS do
        local value = redis.call('GET', KEYS[i])
        table.insert(results, {KEYS[i], value})
    end
    return results

elseif operation == "multi_set" then
    local success = 0
    for i = 1, #KEYS do
        local key = KEYS[i]
        local value = ARGV[i + 1]
        local ttl = tonumber(ARGV[i + 1 + #KEYS])

        if ttl and ttl > 0 then
            redis.call('SETEX', key, ttl, value)
        else
            redis.call('SET', key, value)
        end
        success = success + 1
    end
    return success

elseif operation == "multi_del" then
    if #KEYS > 0 then
        return redis.call('DEL', unpack(KEYS))
    end
    return 0

elseif operation == "conditional_set" then
    local key = KEYS[1]
    local value = ARGV[2]
    local condition_key = ARGV[3]
    local condition_value = ARGV[4]

    local current = redis.call('GET', condition_key)
    if current == condition_value then
        redis.call('SET', key, value)
        return {true, "set"}
    end
    return {false, "condition_failed"}
end

return {false, "unknown_operation"}
`

const priorityQueueOpsScript = `
-- Advanced priority queue operations with analytics
-- KEYS[1]: queue key
-- ARGV[1]: operation ("enqueue", "dequeue", "reorder", "analytics", "bulk_ops")

local operation = ARGV[1]
local queueKey = KEYS[1]

if operation == "enqueue" then
    local items = {}
    local count = tonumber(ARGV[2])

    for i = 1, count do
        local data = ARGV[2 + i]
        local priority = tonumber(ARGV[2 + count + i])
        local score = (4 - priority) * 1000000 + redis.call('TIME')[1]
        table.insert(items, {score, data})
    end

    for i = 1, #items do
        redis.call('ZADD', queueKey, items[i][1], items[i][2])
    end

    redis.call('EXPIRE', queueKey, 3600)
    return redis.call('ZCARD', queueKey)

elseif operation == "dequeue" then
    local count = tonumber(ARGV[2]) or 1
    local items = redis.call('ZRANGE', queueKey, 0, count - 1)

    if #items > 0 then
        for i = 1, #items do
            redis.call('ZREM', queueKey, items[i])
        end
    end

    return items

elseif operation == "analytics" then
    local total = redis.call('ZCARD', queueKey)
    local priorities = {}

    -- Count by priority ranges
    for p = 0, 3 do
        local min_score = (4 - p - 1) * 1000000
        local max_score = (4 - p) * 1000000 - 1
        local count = redis.call('ZCOUNT', queueKey, min_score, max_score)
        priorities[p] = count
    end

    -- Get oldest item
    local oldest = redis.call('ZRANGE', queueKey, 0, 0, 'WITHSCORES')
    local oldest_timestamp = nil
    if #oldest > 0 then
        oldest_timestamp = oldest[2] % 1000000
    end

    return {
        total = total,
        priorities = priorities,
        oldest_timestamp = oldest_timestamp
    }
end

return nil
`