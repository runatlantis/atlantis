// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
package events_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

var repo = "repo/owner"
var workspace = "default"

func TestTryLock(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()

	// The first lock should succeed.
	unlockFn, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)

	// Now another lock for the same repo, workspace, and pull should fail
	_, err = locker.TryLock(repo, workspace, 1)
	ErrEquals(t, "the default workspace is currently locked by another"+
		" command that is running for this pull request–"+
		"wait until the previous command is complete and try again", err)

	// Unlock should work.
	unlockFn()
	_, err = locker.TryLock(repo, workspace, 1)
	Ok(t, err)
}

func TestTryLockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()

	t.Log("a lock for the same repo and pull but different workspace should succeed")
	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	_, err = locker.TryLock(repo, "new-workspace", 1)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, workspace, 1)
	Assert(t, err != nil, "exp err")
	_, err = locker.TryLock(repo, "new-workspace", 1)
	Assert(t, err != nil, "exp err")
}

func TestTryLockDifferentRepo(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()

	t.Log("a lock for a different repo but the same workspace and pull should succeed")
	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	newRepo := "owner/newrepo"
	_, err = locker.TryLock(newRepo, workspace, 1)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, workspace, 1)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(newRepo, workspace, 1)
	ErrContains(t, "currently locked", err)
}

func TestTryLockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()

	t.Log("a lock for a different pull but the same repo and workspace should succeed")
	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	newPull := 2
	_, err = locker.TryLock(repo, workspace, newPull)
	Ok(t, err)

	t.Log("and both should now be locked")
	_, err = locker.TryLock(repo, workspace, 1)
	ErrContains(t, "currently locked", err)
	_, err = locker.TryLock(repo, workspace, newPull)
	ErrContains(t, "currently locked", err)
}

func TestUnlock(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()

	t.Log("unlocking should work")
	unlockFn, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	unlockFn()
	_, err = locker.TryLock(repo, workspace, 1)
	Ok(t, err)
}

func TestUnlockDifferentWorkspaces(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()
	t.Log("unlocking should work for different workspaces")
	unlockFn1, err1 := locker.TryLock(repo, workspace, 1)
	Ok(t, err1)
	unlockFn2, err2 := locker.TryLock(repo, "new-workspace", 1)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	_, err = locker.TryLock(repo, "new-workspace", 1)
	Ok(t, err)
}

func TestUnlockDifferentRepos(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()
	t.Log("unlocking should work for different repos")
	unlockFn1, err1 := locker.TryLock(repo, workspace, 1)
	Ok(t, err1)
	newRepo := "owner/newrepo"
	unlockFn2, err2 := locker.TryLock(newRepo, workspace, 1)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	_, err = locker.TryLock(newRepo, workspace, 1)
	Ok(t, err)
}

func TestUnlockDifferentPulls(t *testing.T) {
	locker := events.NewDefaultAtlantisWorkspaceLocker()
	t.Log("unlocking should work for different pulls")
	unlockFn1, err1 := locker.TryLock(repo, workspace, 1)
	Ok(t, err1)
	newPull := 2
	unlockFn2, err2 := locker.TryLock(repo, workspace, newPull)
	Ok(t, err2)
	unlockFn1()
	unlockFn2()

	_, err := locker.TryLock(repo, workspace, 1)
	Ok(t, err)
	_, err = locker.TryLock(repo, workspace, newPull)
	Ok(t, err)
}
