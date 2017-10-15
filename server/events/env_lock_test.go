package events_test

import (
	"testing"

	. "github.com/hootsuite/atlantis/testing_util"
	"github.com/hootsuite/atlantis/server/events"
)

var repo = "repo/owner"
var env = "default"

func TestTryLock(t *testing.T) {
	locker := events.NewEnvLock()

	t.Log("the first lock should succeed")
	Equals(t, true, locker.TryLock(repo, env, 1))

	t.Log("now another lock for the same repo, env, and pull should fail")
	Equals(t, false, locker.TryLock(repo, env, 1))
}

func TestTryLockDifferentEnv(t *testing.T) {
	locker := events.NewEnvLock()

	t.Log("a lock for the same repo and pull but different env should succeed")
	Equals(t, true, locker.TryLock(repo, env, 1))
	Equals(t, true, locker.TryLock(repo, "new-env", 1))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, 1))
	Equals(t, false, locker.TryLock(repo, "new-env", 1))
}

func TestTryLockDifferentRepo(t *testing.T) {
	locker := events.NewEnvLock()

	t.Log("a lock for a different repo but the same env and pull should succeed")
	Equals(t, true, locker.TryLock(repo, env, 1))
	newRepo := "owner/newrepo"
	Equals(t, true, locker.TryLock(newRepo, env, 1))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, 1))
	Equals(t, false, locker.TryLock(newRepo, env, 1))
}

func TestTryLockDifferent1(t *testing.T) {
	locker := events.NewEnvLock()

	t.Log("a lock for a different pull but the same repo and env should succeed")
	Equals(t, true, locker.TryLock(repo, env, 1))
	new1 := 2
	Equals(t, true, locker.TryLock(repo, env, new1))

	t.Log("and both should now be locked")
	Equals(t, false, locker.TryLock(repo, env, 1))
	Equals(t, false, locker.TryLock(repo, env, new1))
}

func TestUnlock(t *testing.T) {
	locker := events.NewEnvLock()

	t.Log("unlocking should work")
	Equals(t, true, locker.TryLock(repo, env, 1))
	locker.Unlock(repo, env, 1)
	Equals(t, true, locker.TryLock(repo, env, 1))
}

func TestUnlockDifferentEnvs(t *testing.T) {
	locker := events.NewEnvLock()
	t.Log("unlocking should work for different envs")
	Equals(t, true, locker.TryLock(repo, env, 1))
	Equals(t, true, locker.TryLock(repo, "new-env", 1))
	locker.Unlock(repo, env, 1)
	locker.Unlock(repo, "new-env", 1)
	Equals(t, true, locker.TryLock(repo, env, 1))
	Equals(t, true, locker.TryLock(repo, "new-env", 1))
}

func TestUnlockDifferentRepos(t *testing.T) {
	locker := events.NewEnvLock()
	t.Log("unlocking should work for different repos")
	Equals(t, true, locker.TryLock(repo, env, 1))
	newRepo := "owner/newrepo"
	Equals(t, true, locker.TryLock(newRepo, env, 1))
	locker.Unlock(repo, env, 1)
	locker.Unlock(newRepo, env, 1)
	Equals(t, true, locker.TryLock(repo, env, 1))
	Equals(t, true, locker.TryLock(newRepo, env, 1))
}

func TestUnlockDifferentPulls(t *testing.T) {
	locker := events.NewEnvLock()
	t.Log("unlocking should work for different 1s")
	Equals(t, true, locker.TryLock(repo, env, 1))
	new1 := 2
	Equals(t, true, locker.TryLock(repo, env, new1))
	locker.Unlock(repo, env, 1)
	locker.Unlock(repo, env, new1)
	Equals(t, true, locker.TryLock(repo, env, 1))
	Equals(t, true, locker.TryLock(repo, env, new1))
}
