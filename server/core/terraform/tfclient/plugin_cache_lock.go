// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package tfclient

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/runatlantis/atlantis/server/logging"
)

const (
	// pluginCacheLockRetryInterval is how often we retry a non-blocking lock
	// attempt while another process holds the plugin cache lock.
	pluginCacheLockRetryInterval = 100 * time.Millisecond
	// pluginCacheLockWarnAfter is how long we wait before logging a warning
	// (and every interval thereafter) that we are still waiting on the lock.
	pluginCacheLockWarnAfter = 30 * time.Second
	// pluginCacheLockTimeout bounds how long a terraform command waits for
	// the plugin cache lock before giving up. It is deliberately generous:
	// with many parallel projects every init is serialized on this lock, so a
	// wait can legitimately be long. Failing a plan because of a lock timeout
	// is worse than waiting.
	pluginCacheLockTimeout = time.Hour
)

// lockFile acquires an advisory lock on the file at path and returns a
// function that releases it. exclusive takes a write lock; otherwise a
// shared (read) lock is taken so that concurrent readers can proceed while
// a writer waits. The lock is held on the open file description, so it is
// released by the kernel even if this process dies, and it coordinates
// separate Atlantis processes that share the same cache directory on the
// same host (note that flock is not reliable across all network file
// systems). Acquisition is retried until ctx is done, logging progress so a
// stuck lock is visible instead of silently hanging a plan.
func lockFile(ctx context.Context, log logging.SimpleLogging, path string, exclusive bool) (func(), error) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	started := time.Now()
	waiting := false
	nextWarning := pluginCacheLockWarnAfter
	for {
		locked, err := tryLockFile(file, exclusive)
		if err != nil {
			_ = file.Close()
			return nil, err
		}
		if locked {
			return func() {
				_ = unlockFile(file)
				_ = file.Close()
			}, nil
		}

		waited := time.Since(started)
		if !waiting {
			log.Debug("waiting for terraform plugin cache lock %q", path)
			waiting = true
		} else if waited >= nextWarning {
			log.Warn("still waiting for terraform plugin cache lock %q after %s", path, waited.Round(time.Second))
			nextWarning += pluginCacheLockWarnAfter
		}

		timer := time.NewTimer(pluginCacheLockRetryInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			_ = file.Close()
			return nil, fmt.Errorf("waiting for terraform plugin cache lock %q: %w", path, ctx.Err())
		case <-timer.C:
		}
	}
}
