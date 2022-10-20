package file

import "sync"

// RWLock is a wrapper around sync.RWMutex for to easily allow sharing a mutex across
// the codebase
type RWLock struct {
	sync.RWMutex
}
