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
	pluginCacheLockRetryInterval = 100 * time.Millisecond
	pluginCacheLockWarnAfter     = 30 * time.Second
	// TODO: Expose a setting if legitimate waits exceed this fixed ceiling.
	pluginCacheLockTimeout = time.Hour
)

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
