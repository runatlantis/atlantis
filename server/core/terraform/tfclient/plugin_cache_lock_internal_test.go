// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package tfclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// fakeTFScript is a stand-in for the terraform binary used by the plugin
// cache tests. It dispatches on its first argument the same way the client
// uses args[0] to pick the plugin cache lock mode:
//
//	init    simulates installing a provider into the shared plugin cache by
//	        writing the provider binary non-atomically (like terraform does),
//	        guarded by an atomic mkdir so overlapping writers are detected.
//	plan    simulates a plan reading/executing providers from the cache; it
//	        sleeps so tests can observe the command in flight.
//	apply   like plan, but fails when called as "apply fail".
//	version sleeps briefly (sync command path) so tests can observe the
//	        command in flight via the $STARTED marker file.
const fakeTFScript = `#!/bin/sh
cmd="$1"; shift
case "$cmd" in
  init)
    if mkdir "$TF_PLUGIN_CACHE_DIR/.writer-active" 2>/dev/null; then
      printf 'provider-binary' > "$TF_PLUGIN_CACHE_DIR/terraform-provider-aws_v5"
      sleep 0.05
      rmdir "$TF_PLUGIN_CACHE_DIR/.writer-active"
    else
      echo "OVERLAP: concurrent init wrote to the shared plugin cache"
      exit 1
    fi
    ;;
  plan)
    touch "$STARTED" 2>/dev/null
    sleep 0.2
    ;;
  apply)
    if [ "$1" = "fail" ]; then
      echo "boom" >&2
      exit 1
    fi
    sleep 0.2
    ;;
  version)
    touch "$STARTED" 2>/dev/null
    sleep 0.3
    echo "Terraform v1.5.0"
    ;;
  *)
    echo "unknown command $cmd" >&2
    exit 1
    ;;
esac
`

// newFakeTFClient returns a DefaultClient whose terraform binary is a fake
// script, plus the shared plugin cache dir.
func newFakeTFClient(t *testing.T, cacheDir string) *DefaultClient {
	t.Helper()
	binDir := t.TempDir()
	fakeTF := filepath.Join(binDir, "terraform")
	Ok(t, os.WriteFile(fakeTF, []byte(fakeTFScript), 0o600))
	Ok(t, os.Chmod(fakeTF, 0o755))
	v := version.Must(version.NewVersion("1.5.0"))
	return &DefaultClient{
		defaultVersion:          v,
		terraformPluginCacheDir: cacheDir,
		binDir:                  binDir,
		overrideTF:              fakeTF,
		usePluginCache:          true,
	}
}

func fakeTFProjectContext(t *testing.T) command.ProjectContext {
	return command.ProjectContext{Log: logging.NewNoopLogger(t), SuppressJobOutput: true}
}

func fakeTFDistribution() terraform.Distribution {
	return terraform.NewDistributionTerraform()
}

// TestLockFile_ExclusiveExcludesReadersAndWriters verifies the cache lock is a
// real reader/writer lock: while an exclusive (init) lock is held, neither a
// shared nor an exclusive acquisition succeeds, and once released a shared
// acquisition succeeds immediately.
func TestLockFile_ExclusiveExcludesReadersAndWriters(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(t.TempDir(), ".atlantis.lock")

	unlock, err := lockFile(context.Background(), logger, lockPath, true)
	Ok(t, err)

	// Shared acquisition must fail while the exclusive lock is held.
	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()
	_, err = lockFile(ctx, logger, lockPath, false)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "expected shared lock wait to time out, got %v", err)

	// Exclusive acquisition must fail too.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	_, err = lockFile(ctx2, logger, lockPath, true)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "expected exclusive lock wait to time out, got %v", err)

	unlock()

	// After release a shared lock is available right away.
	unlock2, err := lockFile(context.Background(), logger, lockPath, false)
	Ok(t, err)
	unlock2()
}

// TestLockFile_SharedAllowsConcurrentReadersButBlocksWriters verifies that
// plans/applies (shared lock) can run concurrently, but an init (exclusive)
// waits until all readers are done.
func TestLockFile_SharedAllowsConcurrentReadersButBlocksWriters(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(t.TempDir(), ".atlantis.lock")

	unlockReader1, err := lockFile(context.Background(), logger, lockPath, false)
	Ok(t, err)

	// A second reader must be able to acquire the shared lock immediately.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlockReader2, err := lockFile(ctx, logger, lockPath, false)
	Ok(t, err)

	// A writer must not be able to acquire while readers hold the lock.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel2()
	_, err = lockFile(ctx2, logger, lockPath, true)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "expected exclusive lock to wait for readers, got %v", err)

	unlockReader1()
	unlockReader2()

	// Once all readers released, a writer acquires promptly.
	unlockWriter, err := lockFile(context.Background(), logger, lockPath, true)
	Ok(t, err)
	unlockWriter()
}

// TestDefaultClient_LockPluginCache_NoCacheDirIsNoop verifies that when the
// plugin cache is disabled (no TF_PLUGIN_CACHE_DIR in the command env) no lock
// file is created and the returned unlock is a no-op.
func TestDefaultClient_LockPluginCache_NoCacheDirIsNoop(t *testing.T) {
	client := &DefaultClient{}
	unlock, err := client.lockPluginCache(logging.NewNoopLogger(t), "init", []string{"PATH=/usr/bin"})
	Ok(t, err)
	unlock()
}

// TestDefaultClient_LockPluginCache_CreatesCacheDirAndLockFile verifies the
// lock file lives inside the (possibly not yet existing) cache directory.
func TestDefaultClient_LockPluginCache_CreatesCacheDirAndLockFile(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "missing-cache")
	client := &DefaultClient{}
	environ := []string{"TF_PLUGIN_CACHE_DIR=" + cacheDir}

	unlock, err := client.lockPluginCache(logging.NewNoopLogger(t), "init", environ)
	Ok(t, err)
	_, err = os.Stat(filepath.Join(cacheDir, pluginCacheLockFileName))
	Ok(t, err)
	unlock()
}

// TestDefaultClient_LockPluginCache_InitExclusiveVsOthersShared verifies the
// lock mode chosen per terraform command: init is exclusive, everything else
// is shared.
func TestDefaultClient_LockPluginCache_InitExclusiveVsOthersShared(t *testing.T) {
	cacheDir := t.TempDir()
	logger := logging.NewNoopLogger(t)
	client := &DefaultClient{}
	environ := []string{"TF_PLUGIN_CACHE_DIR=" + cacheDir}
	lockPath := filepath.Join(cacheDir, pluginCacheLockFileName)

	unlockInit, err := client.lockPluginCache(logger, "init", environ)
	Ok(t, err)

	type result struct {
		unlock func()
		err    error
	}
	planLocked := make(chan result, 1)
	go func() {
		unlock, err := client.lockPluginCache(logger, "plan", environ)
		planLocked <- result{unlock: unlock, err: err}
	}()

	select {
	case got := <-planLocked:
		if got.unlock != nil {
			got.unlock()
		}
		unlockInit()
		t.Fatalf("plan acquired the shared plugin cache lock while init held the exclusive lock: %v", got.err)
	case <-time.After(150 * time.Millisecond):
	}

	unlockInit()
	select {
	case got := <-planLocked:
		Ok(t, got.err)
		got.unlock()
	case <-time.After(time.Second):
		t.Fatal("plan did not acquire the shared plugin cache lock after init released it")
	}

	// Sanity: plan (shared) does not block another plan.
	unlockPlan1, err := client.lockPluginCache(logger, "plan", environ)
	Ok(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	unlockPlan2, err := lockFile(ctx, logger, lockPath, false)
	Ok(t, err)
	unlockPlan2()
	unlockPlan1()
}

// waitForFile polls until path exists or the deadline passes.
func waitForFile(t *testing.T, path string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for {
		if _, err := os.Stat(path); err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %q to be created", path)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestDefaultClient_RunCommandAsyncHoldsPluginCacheLockUntilExit verifies the
// lock acquired by RunCommandAsync is held for the whole lifetime of the
// terraform process, not just until the channels are returned.
func TestDefaultClient_RunCommandAsyncHoldsPluginCacheLockUntilExit(t *testing.T) {
	cacheDir := t.TempDir()
	client := newFakeTFClient(t, cacheDir)
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(cacheDir, pluginCacheLockFileName)
	v := client.defaultVersion
	started := filepath.Join(t.TempDir(), "started")

	// Run a plan (shared lock) async; the fake binary signals via $STARTED
	// once it is in flight.
	_, outCh := client.RunCommandAsync(fakeTFProjectContext(t), t.TempDir(), []string{"plan"}, map[string]string{"STARTED": started}, fakeTFDistribution(), v, "default")
	waitForFile(t, started, 2*time.Second)

	// While the async plan runs, an exclusive (init) lock must not be grantable.
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := lockFile(ctx, logger, lockPath, true)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "init acquired the plugin cache while a plan was running: %v", err)

	// Drain the channel; once the command exits the lock must be free.
	_, err = waitCh(outCh)
	Ok(t, err)
	unlock, err := lockFile(context.Background(), logger, lockPath, true)
	Ok(t, err)
	unlock()
}

// TestDefaultClient_RunCommandAsyncReleasesPluginCacheLockWhenConsumerBreaksOnError
// verifies that when a caller stops reading the output channel after an error
// (the documented contract), the plugin cache lock is still released.
func TestDefaultClient_RunCommandAsyncReleasesPluginCacheLockWhenConsumerBreaksOnError(t *testing.T) {
	cacheDir := t.TempDir()
	client := newFakeTFClient(t, cacheDir)
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(cacheDir, pluginCacheLockFileName)

	// "apply fail" takes the shared lock (apply is not init) and exits 1.
	_, outCh := client.RunCommandAsync(fakeTFProjectContext(t), t.TempDir(), []string{"apply", "fail"}, nil, fakeTFDistribution(), client.defaultVersion, "default")

	sawError := false
	for line := range outCh {
		if line.Err != nil {
			sawError = true
			break
		}
	}
	Assert(t, sawError, "expected the command to emit an error")

	acquired := make(chan error, 1)
	go func() {
		unlock, err := lockFile(context.Background(), logger, lockPath, true)
		if err == nil {
			unlock()
		}
		acquired <- err
	}()
	select {
	case err := <-acquired:
		Ok(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("plugin cache lock was not released after the consumer stopped on command error")
	}
}

// TestDefaultClient_ConcurrentInitsSharingPluginCacheAreSerialized is the
// regression test for issue #2242: many parallel project commands running
// `terraform init` against one shared TF_PLUGIN_CACHE_DIR. Without
// coordination the inits race writing the same provider binary into the
// cache (the "text file busy"/corruption from the issue); with the fix every
// init is serialized and none overlap.
func TestDefaultClient_ConcurrentInitsSharingPluginCacheAreSerialized(t *testing.T) {
	cacheDir := t.TempDir()
	client := newFakeTFClient(t, cacheDir)
	v := client.defaultVersion

	const workers = 12
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			// Each "project" runs init in its own directory but shares the
			// plugin cache dir, exactly like parallel plans in one PR.
			projectDir := filepath.Join(t.TempDir(), fmt.Sprintf("project-%d", i))
			Ok(t, os.MkdirAll(projectDir, 0o755))
			_, err := client.RunCommandWithVersion(fakeTFProjectContext(t), projectDir, []string{"init", "-input=false"}, nil, fakeTFDistribution(), v, "default")
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		Assert(t, err == nil || !strings.Contains(err.Error(), "OVERLAP"), "concurrent init overlap detected: %v", err)
		Ok(t, err)
	}

	// The provider binary in the cache must be intact, not a torn write.
	content, err := os.ReadFile(filepath.Join(cacheDir, "terraform-provider-aws_v5"))
	Ok(t, err)
	Assert(t, string(content) == "provider-binary", "plugin cache provider binary was corrupted: %q", content)
}

// TestDefaultClient_ConcurrentPlansSharingPluginCacheStayParallel verifies the
// fix does not serialize plans: readers take a shared lock, so parallel plans
// still overlap (the parallelism needed for large parallel plan runs is kept).
// Each fake plan sleeps 0.2s; if plans were serialized 8 workers would take
// >= 1.6s, while overlapping they finish in well under half that.
func TestDefaultClient_ConcurrentPlansSharingPluginCacheStayParallel(t *testing.T) {
	cacheDir := t.TempDir()
	client := newFakeTFClient(t, cacheDir)
	v := client.defaultVersion

	const workers = 8
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	start := time.Now()
	for i := range workers {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			projectDir := filepath.Join(t.TempDir(), fmt.Sprintf("project-%d", i))
			Ok(t, os.MkdirAll(projectDir, 0o755))
			_, err := client.RunCommandWithVersion(fakeTFProjectContext(t), projectDir, []string{"plan", "-out=plan.tfplan"}, nil, fakeTFDistribution(), v, "default")
			if err != nil {
				errs <- err
			}
		}(i)
	}
	wg.Wait()
	elapsed := time.Since(start)
	close(errs)
	for err := range errs {
		Ok(t, err)
	}

	Assert(t, elapsed < time.Duration(workers)*200*time.Millisecond/2,
		"expected parallel plans to overlap on the shared plugin cache, but %d plans took %s (serialized?)", workers, elapsed)
}

// TestDefaultClient_RunCommandWithVersionSyncPathTakesSharedLock verifies the
// synchronous (non-streaming) command path also holds the plugin cache lock
// while the command runs.
func TestDefaultClient_RunCommandWithVersionSyncPathTakesSharedLock(t *testing.T) {
	cacheDir := t.TempDir()
	client := newFakeTFClient(t, cacheDir)
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(cacheDir, pluginCacheLockFileName)
	started := filepath.Join(t.TempDir(), "started")

	done := make(chan error, 1)
	go func() {
		// "version" is not a streaming command, so it takes the sync path in
		// RunCommandWithVersion; the fake binary sleeps 0.3s after creating
		// the $STARTED marker.
		_, err := client.RunCommandWithVersion(fakeTFProjectContext(t), t.TempDir(), []string{"version"}, map[string]string{"STARTED": started}, fakeTFDistribution(), client.defaultVersion, "default")
		done <- err
	}()

	// The lock is acquired before the process starts, so once the command is
	// in flight an exclusive lock must be blocked.
	waitForFile(t, started, 2*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err := lockFile(ctx, logger, lockPath, true)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "sync-path command did not hold the plugin cache lock: %v", err)

	Ok(t, <-done)
	unlock, err := lockFile(context.Background(), logger, lockPath, true)
	Ok(t, err)
	unlock()
}

// TestPluginCacheLockHonorsContextCancellation verifies lock acquisition
// gives up when the context is done instead of hanging forever.
func TestPluginCacheLockHonorsContextCancellation(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	lockPath := filepath.Join(t.TempDir(), ".atlantis.lock")

	unlock, err := lockFile(context.Background(), logger, lockPath, true)
	Ok(t, err)
	defer unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	_, err = lockFile(ctx, logger, lockPath, false)
	Assert(t, errors.Is(err, context.DeadlineExceeded), "expected lock wait to honor context cancellation, got %v", err)
}
