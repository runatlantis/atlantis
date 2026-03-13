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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
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
	"go.uber.org/mock/gomock"
)

func TestCleanUpPullWorkspaceErr(t *testing.T) {
	t.Log("when workspace.Delete returns an error, we return it")
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	w := mocks.NewMockWorkingDir(ctrl)
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	pce := events.PullClosedExecutor{
		WorkingDir:         w,
		PullClosedTemplate: &events.PullClosedEventTemplate{},
		Database:           db,
	}
	expErr := errors.New("err")
	w.EXPECT().Delete(logger, testdata.GithubRepo, testdata.Pull).Return(expErr)
	actualErr := pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning workspace: err", actualErr.Error())
}

func TestCleanUpPullUnlockErr(t *testing.T) {
	t.Log("when locker.UnlockByPull returns an error, we return it")
	logger := logging.NewNoopLogger(t)
	ctrl := gomock.NewController(t)
	w := mocks.NewMockWorkingDir(ctrl)
	l := lockmocks.NewMockLocker(ctrl)
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	pce := events.PullClosedExecutor{
		Locker:             l,
		WorkingDir:         w,
		Database:           db,
		PullClosedTemplate: &events.PullClosedEventTemplate{},
	}
	expErr := errors.New("err")
	w.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	l.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, expErr)
	actualErr := pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning up locks: err", actualErr.Error())
}

func TestCleanUpPullNoLocks(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	t.Log("when there are no locks to clean up, we don't comment")
	RegisterMockTestingT(t)
	ctrl := gomock.NewController(t)
	w := mocks.NewMockWorkingDir(ctrl)
	l := lockmocks.NewMockLocker(ctrl)
	cp := vcsmocks.NewMockClient()
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	pce := events.PullClosedExecutor{
		Locker:     l,
		VCSClient:  cp,
		WorkingDir: w,
		Database:   db,
	}
	w.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	l.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, nil)
	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Ok(t, err)
	cp.VerifyWasCalled(Never()).CreateComment(Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]())
}

func TestCleanUpPullComments(t *testing.T) {
	logger := logging.NewNoopLogger(t)
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
					Project:   models.NewProject("owner/repo", "", ""),
					Workspace: "default",
				},
			},
			"- dir: `.` workspace: `default`",
		},
		{
			"single lock, named project",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "", "projectname"),
					Workspace: "default",
				},
			},
			// TODO: Should project name be included in output?
			"- dir: `.` workspace: `default`",
		},
		{
			"single lock, non-empty path",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path", ""),
					Workspace: "default",
				},
			},
			"- dir: `path` workspace: `default`",
		},
		{
			"single path, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path", ""),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path", ""),
					Workspace: "workspace2",
				},
			},
			"- dir: `path` workspaces: `workspace1`, `workspace2`",
		},
		{
			"multiple paths, multiple workspaces",
			[]models.ProjectLock{
				{
					Project:   models.NewProject("owner/repo", "path", ""),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path", ""),
					Workspace: "workspace2",
				},
				{
					Project:   models.NewProject("owner/repo", "path2", ""),
					Workspace: "workspace1",
				},
				{
					Project:   models.NewProject("owner/repo", "path2", ""),
					Workspace: "workspace2",
				},
			},
			"- dir: `path` workspaces: `workspace1`, `workspace2`\n- dir: `path2` workspaces: `workspace1`, `workspace2`",
		},
	}
	for _, c := range cases {
		func() {
			ctrl := gomock.NewController(t)
			w := mocks.NewMockWorkingDir(ctrl)
			cp := vcsmocks.NewMockClient()
			l := lockmocks.NewMockLocker(ctrl)
			tmp := t.TempDir()
			db, err := boltdb.New(tmp)
			t.Cleanup(func() {
				db.Close()
			})
			Ok(t, err)
			pce := events.PullClosedExecutor{
				Locker:     l,
				VCSClient:  cp,
				WorkingDir: w,
				Database:   db,
			}
			t.Log("testing: " + c.Description)
			w.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
			l.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(c.Locks, nil)
			err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
			Ok(t, err)
			_, _, _, comment, _ := cp.VerifyWasCalledOnce().CreateComment(
				Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()

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
			panic(fmt.Errorf("failed to create temp file: %w", err))
		}
		path := f.Name()
		f.Close() // nolint: errcheck

		// Open the database.
		boltDB, err := bolt.Open(path, 0600, nil)
		if err != nil {
			panic(fmt.Errorf("could not start bolt DB: %w", err))
		}
		if err := boltDB.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists([]byte(pullsBucketName)); err != nil {
				return fmt.Errorf("failed to create bucket: %w", err)
			}
			return nil
		}); err != nil {
			panic(fmt.Errorf("could not create bucket: %w", err))
		}
		database, _ := boltdb.NewWithDB(boltDB, lockBucket, configBucket)
		result := []command.ProjectResult{
			{
				RepoRelDir:  testdata.GithubRepo.FullName,
				Workspace:   "default",
				ProjectName: *testdata.Project.Name,
			},
		}

		// Create a new record for pull
		_, err = database.UpdatePullWithResults(testdata.Pull, result)
		Ok(t, err)

		gmockCtrl := gomock.NewController(t)
		workingDir := mocks.NewMockWorkingDir(gmockCtrl)
		locker := lockmocks.NewMockLocker(gmockCtrl)
		client := vcsmocks.NewMockClient()
		logger := loggermocks.NewMockSimpleLogging()

		pullClosedExecutor := events.PullClosedExecutor{
			Locker:                   locker,
			WorkingDir:               workingDir,
			Database:                 database,
			VCSClient:                client,
			PullClosedTemplate:       &events.PullClosedEventTemplate{},
			LogStreamResourceCleaner: prjCmdOutHandler,
		}

		locks := []models.ProjectLock{
			{
				Project:   models.NewProject(testdata.GithubRepo.FullName, "", ""),
				Workspace: "default",
			},
		}
		workingDir.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
		locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(locks, nil)

		// Clean up.
		err = pullClosedExecutor.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
		Ok(t, err)

		close(prjCmdOutput)
		_, _, _, comment, _ := client.VerifyWasCalledOnce().CreateComment(
			Any[logging.SimpleLogging](), Any[models.Repo](), Any[int](), Any[string](), Any[string]()).GetCapturedArguments()
		expectedComment := "Locks and plans deleted for the projects and workspaces modified in this pull request:\n\n" + "- dir: `.` workspace: `default`"
		Equals(t, expectedComment, comment)

		// Assert log streaming resources are cleaned up.
		dfPrjCmdOutputHandler := prjCmdOutHandler.(*jobs.AsyncProjectCommandOutputHandler)
		assert.Empty(t, dfPrjCmdOutputHandler.GetProjectOutputBuffer(ctx.PullInfo()))
		assert.Empty(t, dfPrjCmdOutputHandler.GetReceiverBufferForPull(ctx.PullInfo()))
	})
}

func TestCleanUpPullWithCorrectJobContext(t *testing.T) {
	t.Log("CleanUpPull should call LogStreamResourceCleaner.CleanUp with complete PullInfo including RepoFullName and Path")
	logger := logging.NewNoopLogger(t)

	// Create mocks
	ctrl := gomock.NewController(t)
	workingDir := mocks.NewMockWorkingDir(ctrl)
	locker := lockmocks.NewMockLocker(ctrl)
	RegisterMockTestingT(t)
	client := vcsmocks.NewMockClient()
	resourceCleaner := mocks.NewMockResourceCleaner(ctrl)

	// Create temporary database
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)

	// Create test data with multiple projects to verify all fields are populated correctly
	testProjects := []command.ProjectResult{
		{
			RepoRelDir:  "path/to/project1",
			Workspace:   "default",
			ProjectName: "project1",
		},
		{
			RepoRelDir:  "path/to/project2",
			Workspace:   "staging",
			ProjectName: "project2",
		},
	}

	// Add pull status to database
	_, err = db.UpdatePullWithResults(testdata.Pull, testProjects)
	Ok(t, err)

	// Create executor
	pce := events.PullClosedExecutor{
		Locker:                   locker,
		VCSClient:                client,
		WorkingDir:               workingDir,
		Database:                 db,
		PullClosedTemplate:       &events.PullClosedEventTemplate{},
		LogStreamResourceCleaner: resourceCleaner,
	}

	// Setup mock expectations BEFORE execution
	workingDir.EXPECT().Delete(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, nil)

	var capturedArgs []jobs.PullInfo
	resourceCleaner.EXPECT().CleanUp(gomock.Any()).Times(2).Do(func(pullInfo jobs.PullInfo) {
		capturedArgs = append(capturedArgs, pullInfo)
	})

	// Execute CleanUpPull
	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Ok(t, err)

	// Verify first project's PullInfo
	expectedPullInfo1 := jobs.PullInfo{
		PullNum:      testdata.Pull.Num,
		Repo:         testdata.Pull.BaseRepo.Name,
		RepoFullName: testdata.Pull.BaseRepo.FullName,
		ProjectName:  "project1",
		Path:         "path/to/project1",
		Workspace:    "default",
	}
	Equals(t, expectedPullInfo1, capturedArgs[0])

	// Verify second project's PullInfo
	expectedPullInfo2 := jobs.PullInfo{
		PullNum:      testdata.Pull.Num,
		Repo:         testdata.Pull.BaseRepo.Name,
		RepoFullName: testdata.Pull.BaseRepo.FullName,
		ProjectName:  "project2",
		Path:         "path/to/project2",
		Workspace:    "staging",
	}
	Equals(t, expectedPullInfo2, capturedArgs[1])
}
