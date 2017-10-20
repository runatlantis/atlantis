package events

import (
	"fmt"
	"sync"
)

//go:generate pegomock generate --use-experimental-model-gen --package mocks -o mocks/mock_env_locker.go EnvLocker
type EnvLocker interface {
	TryLock(repoFullName string, env string, pullNum int) bool
	Unlock(repoFullName, env string, pullNum int)
}

// EnvLock is used to prevent multiple runs and commands from occurring at the same time for a single
// repo, pull, and environment
type EnvLock struct {
	mutex sync.Mutex
	locks map[string]interface{}
}

func NewEnvLock() *EnvLock {
	return &EnvLock{
		locks: make(map[string]interface{}),
	}
}

// TryLock returns true if you acquired the lock and false if someone else already has the lock
func (c *EnvLock) TryLock(repoFullName string, env string, pullNum int) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.key(repoFullName, env, pullNum)
	if _, ok := c.locks[key]; !ok {
		c.locks[key] = true
		return true
	}
	return false
}

// Unlock unlocks the repo and environment
func (c *EnvLock) Unlock(repoFullName, env string, pullNum int) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.locks, c.key(repoFullName, env, pullNum))
}

func (c *EnvLock) key(repo string, env string, pull int) string {
	return fmt.Sprintf("%s/%s/%d", repo, env, pull)
}
