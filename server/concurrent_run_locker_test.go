package server_test

import (
	"testing"

	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
)

var repo = "repo/owner"
var env = "default"
var pull = 1

func TestTryLock(t *testing.T) {
	locker := server.NewConcurrentRunLocker()

	t.Log("the first lock should succeed")
	Equals(t, true, locker.TryLock(repo, env, pull))

	t.Log("now another lock for the same repo, env, and pull should fail")
	Equals(t, false, locker.TryLock(repo, env, pull))
}

func TestTryLockDifferentEnv(t *testing.T) {
	locker := server.NewConcurrentRunLocker()

	t.Log("a lock for the same repo and pull but different env should succeed")
	Equals(t, true, locker.TryLock(repo, env, pull))
	Equals(t, true, locker.TryLock(repo, "new-env", pull))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, pull))
	Equals(t, false, locker.TryLock(repo, "new-env", pull))
}

func TestTryLockDifferentRepo(t *testing.T) {
	locker := server.NewConcurrentRunLocker()

	t.Log("a lock for a different repo but the same env and pull should succeed")
	Equals(t, true, locker.TryLock(repo, env, pull))
	newRepo := "owner/newrepo"
	Equals(t, true, locker.TryLock(newRepo, env, pull))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, pull))
	Equals(t, false, locker.TryLock(newRepo, env, pull))
}

func TestTryLockDifferentPull(t *testing.T) {
	locker := server.NewConcurrentRunLocker()

	t.Log("a lock for a different pull but the same repo and env should succeed")
	Equals(t, true, locker.TryLock(repo, env, pull))
	newPull := 2
	Equals(t, true, locker.TryLock(repo, env, newPull))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, pull))
	Equals(t, false, locker.TryLock(repo, env, newPull))
}

func TestUnlock(t *testing.T) {
	locker := server.NewConcurrentRunLocker()

	t.Log("unlocking should work")
	Equals(t, true, locker.TryLock(repo, env, pull))
	locker.Unlock(repo, env, pull)
	Equals(t, true, locker.TryLock(repo, env, pull))
}

func TestUnlockDifferentEnvs(t *testing.T) {
	locker := server.NewConcurrentRunLocker()
	t.Log("unlocking should work for different envs")
	Equals(t, true, locker.TryLock(repo, env, pull))
	Equals(t, true, locker.TryLock(repo, "new-env", pull))
	locker.Unlock(repo, env, pull)
	locker.Unlock(repo, "new-env", pull)
	Equals(t, true, locker.TryLock(repo, env, pull))
	Equals(t, true, locker.TryLock(repo, "new-env", pull))
}

func TestUnlockDifferentRepos(t *testing.T) {
	locker := server.NewConcurrentRunLocker()
	t.Log("unlocking should work for different repos")
	Equals(t, true, locker.TryLock(repo, env, pull))
	newRepo := "owner/newrepo"
	Equals(t, true, locker.TryLock(newRepo, env, pull))
	locker.Unlock(repo, env, pull)
	locker.Unlock(newRepo, env, pull)
	Equals(t, true, locker.TryLock(repo, env, pull))
	Equals(t, true, locker.TryLock(newRepo, env, pull))
}

func TestUnlockDifferentPulls(t *testing.T) {
	locker := server.NewConcurrentRunLocker()
	t.Log("unlocking should work for different pulls")
	Equals(t, true, locker.TryLock(repo, env, pull))
	newPull := 2
	Equals(t, true, locker.TryLock(repo, env, newPull))
	locker.Unlock(repo, env, pull)
	locker.Unlock(repo, env, newPull)
	Equals(t, true, locker.TryLock(repo, env, pull))
	Equals(t, true, locker.TryLock(repo, env, newPull))
}
