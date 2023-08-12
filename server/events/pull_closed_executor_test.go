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

package events_test

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"

	. "github.com/petergtz/pegomock/v4"
	lockmocks "github.com/runatlantis/atlantis/server/core/locking/mocks"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/models/testdata"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	loggermocks "github.com/runatlantis/atlantis/server/logging/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestCleanUpPullWorkspaceErr(t *testing.T) {
	t.Log("when workspace.Delete returns an error, we return it")
	RegisterMockTestingT(t)
	w := mocks.NewMockWorkingDir()
	tmp := t.TempDir()
	db, err := db.New(tmp, false)
	Ok(t, err)
	pce := events.PullClosedExecutor{
		WorkingDir:         w,
		PullClosedTemplate: &events.PullClosedEventTemplate{},
		Backend:            db,
	}
	err = errors.New("err")
	When(w.Delete(testdata.GithubRepo, testdata.Pull)).ThenReturn(err)
	actualErr := pce.CleanUpPull(testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning workspace: err", actualErr.Error())
}

func TestCleanUpPullUnlockErr(t *testing.T) {
	t.Log("when locker.UnlockByPull returns an error, we return it")
	RegisterMockTestingT(t)
	w := mocks.NewMockWorkingDir()
	l := lockmocks.NewMockLocker()
	tmp := t.TempDir()
	db, err := db.New(tmp, false)
	Ok(t, err)
	pce := events.PullClosedExecutor{
		Locker:             l,
		WorkingDir:         w,
		Backend:            db,
		PullClosedTemplate: &events.PullClosedEventTemplate{},
	}
	err = errors.New("err")
	When(l.UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num, true)).ThenReturn(nil, nil, err)
	actualErr := pce.CleanUpPull(testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning up locks: err", actualErr.Error())
}

func TestCleanUpPullNoLocks(t *testing.T) {
	t.Log("when there are no locks to clean up, we don't comment")
	RegisterMockTestingT(t)
	w := mocks.NewMockWorkingDir()
	l := lockmocks.NewMockLocker()
	cp := vcsmocks.NewMockClient()
	tmp := t.TempDir()
	db, err := db.New(tmp, false)
	Ok(t, err)
	pce := events.PullClosedExecutor{
		Locker:     l,
		VCSClient:  cp,
		WorkingDir: w,
		Backend:    db,
	}
	When(l.UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num, true)).ThenReturn(nil, nil, nil)
	err = pce.CleanUpPull(testdata.GithubRepo, testdata.Pull)
	Ok(t, err)
	cp.VerifyWasCalled(Never()).CreateComment(Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestCleanUpPullComments(t *testing.T) {
	t.Log("should comment correctly")
	RegisterMockTestingT(t)
	cases := []struct {
		Description string
		Locks       []models.ProjectLock
		Exp         string
	}{
		{
			"single lock, empty path",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", ""),
					Workspace: "default",
				},
			},
			"- dir: `.` workspace: `default`",
		},
		{
			"single lock, non-empty path",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "default",
				},
			},
			"- dir: `path` workspace: `default`",
		},
		{
			"single path, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace2",
				},
			},
			"- dir: `path` workspaces: `workspace1`, `workspace2`",
		},
		{
			"multiple paths, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path"),
					Workspace: "workspace2",
				},
				{
					Project:   models.NewProject("owner/repo", "path2"),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path2"),
					Workspace: "workspace2",
				},
			},
			"- dir: `path` workspaces: `workspace1`, `workspace2`\n- dir: `path2` workspaces: `workspace1`, `workspace2`",
		},
	}
	for _, c := range cases {
		func() {
			w := mocks.NewMockWorkingDir()
			cp := vcsmocks.NewMockClient()
			l := lockmocks.NewMockLocker()
			tmp := t.TempDir()
			db, err := db.New(tmp, false)
			Ok(t, err)
			pce := events.PullClosedExecutor{
				Locker:     l,
				VCSClient:  cp,
				WorkingDir: w,
				Backend:    db,
			}
			t.Log("testing: " + c.Description)
			When(l.UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num, true)).ThenReturn(c.Locks, nil, nil)
			err = pce.CleanUpPull(testdata.GithubRepo, testdata.Pull)
			Ok(t, err)
			_, _, comment, _ := cp.VerifyWasCalledOnce().CreateComment(Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()

			expected := "Locks and plans deleted for the projects and workspaces modified in this pull request:\n\n" + c.Exp
			Equals(t, expected, comment)
		}()
	}
}

func TestCleanUpLogStreaming(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	RegisterMockTestingT(t)

	t.Run("Should Clean Up Log Streaming Resources When PR is closed", func(t *testing.T) {

		// Create Log streaming resources
		prjCmdOutput := make(chan *jobs.ProjectCmdOutputLine)
		prjCmdOutHandler := jobs.NewAsyncProjectCommandOutputHandler(prjCmdOutput, logger)
		ctx := command.ProjectContext{
			BaseRepo:    testdata.GithubRepo,
			Pull:        testdata.Pull,
			ProjectName: *testdata.Project.Name,
			Workspace:   "default",
		}

		go prjCmdOutHandler.Handle()
		prjCmdOutHandler.Send(ctx, "Test Message", false)

		// Create boltdb and add pull request.
		var lockBucket = "bucket"
		var configBucket = "configBucket"
		var pullsBucketName = "pulls"

		f, err := os.CreateTemp("", "")
		if err != nil {
			panic(errors.Wrap(err, "failed to create temp file"))
		}
		path := f.Name()
		f.Close() // nolint: errcheck

		// Open the database.
		boltDB, err := bolt.Open(path, 0600, nil)
		if err != nil {
			panic(errors.Wrap(err, "could not start bolt DB"))
		}
		if err := boltDB.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists([]byte(pullsBucketName)); err != nil {
				return errors.Wrap(err, "failed to create bucket")
			}
			return nil
		}); err != nil {
			panic(errors.Wrap(err, "could not create bucket"))
		}
		db, _ := db.NewWithDB(boltDB, lockBucket, configBucket, false)
		result := []command.ProjectResult{
			{
				RepoRelDir:  testdata.GithubRepo.FullName,
				Workspace:   "default",
				ProjectName: *testdata.Project.Name,
			},
		}

		// Create a new record for pull
		_, err = db.UpdatePullWithResults(testdata.Pull, result)
		Ok(t, err)

		workingDir := mocks.NewMockWorkingDir()
		locker := lockmocks.NewMockLocker()
		client := vcsmocks.NewMockClient()
		logger := loggermocks.NewMockSimpleLogging()

		pullClosedExecutor := events.PullClosedExecutor{
			Locker:                   locker,
			WorkingDir:               workingDir,
			Backend:                  db,
			VCSClient:                client,
			PullClosedTemplate:       &events.PullClosedEventTemplate{},
			LogStreamResourceCleaner: prjCmdOutHandler,
			Logger:                   logger,
		}

		locks := []models.ProjectLock{
			{
				Project:   models.NewProject(testdata.GithubRepo.FullName, ""),
				Workspace: "default",
			},
		}
		When(locker.UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num, true)).ThenReturn(locks, nil, nil)

		// Clean up.
		err = pullClosedExecutor.CleanUpPull(testdata.GithubRepo, testdata.Pull)
		Ok(t, err)

		close(prjCmdOutput)
		_, _, comment, _ := client.VerifyWasCalledOnce().CreateComment(Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()
		expectedComment := "Locks and plans deleted for the projects and workspaces modified in this pull request:\n\n" + "- dir: `.` workspace: `default`"
		Equals(t, expectedComment, comment)

		// Assert log streaming resources are cleaned up.
		dfPrjCmdOutputHandler := prjCmdOutHandler.(*jobs.AsyncProjectCommandOutputHandler)
		assert.Empty(t, dfPrjCmdOutputHandler.GetProjectOutputBuffer(ctx.PullInfo()))
		assert.Empty(t, dfPrjCmdOutputHandler.GetReceiverBufferForPull(ctx.PullInfo()))
	})
}
