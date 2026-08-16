// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/runatlantis/atlantis/server/core/boltdb"
	coredb "github.com/runatlantis/atlantis/server/core/db"
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

func TestCleanUpPullPublicationClaimRecovery(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	workingDir := mocks.NewMockWorkingDir()
	resourceCleaner := mocks.NewMockResourceCleaner()
	store := &countingPlanStore{}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	_, err = database.UpdatePullWithResults(testdata.Pull, []command.ProjectResult{{
		Workspace: "default", RepoRelDir: "project-a", ProjectName: "a",
	}})
	Ok(t, err)
	Ok(t, database.AcquirePlanPublicationClaim(testdata.Pull, "publisher"))
	notifyingDatabase := &notifyBusyClaimDatabase{Database: database, busy: make(chan struct{})}
	lockerController := gomock.NewController(t)
	locker := lockmocks.NewMockLocker(lockerController)
	locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, nil)
	pce := events.PullClosedExecutor{
		Locker:                   locker,
		WorkingDir:               workingDir,
		Database:                 notifyingDatabase,
		PlanStore:                store,
		LogStreamResourceCleaner: resourceCleaner,
	}

	done := make(chan error, 1)
	go func() { done <- pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull) }()
	<-notifyingDatabase.busy
	select {
	case err := <-done:
		t.Fatalf("cleanup returned while publisher still held claim: %v", err)
	default:
	}
	workingDir.VerifyWasCalled(Never()).Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
	resourceCleaner.VerifyWasCalled(Never()).CleanUp(Any[jobs.PullInfo]())
	Equals(t, 0, store.deleteForPullCalls)
	status, statusErr := database.GetPullStatus(testdata.Pull)
	Ok(t, statusErr)
	Assert(t, status != nil, "busy cleanup must not delete pull status")
	secondOwnerErr := database.AcquirePlanPublicationClaim(testdata.Pull, "second-owner")
	Assert(t, errors.Is(secondOwnerErr, coredb.ErrPlanPublicationBusy), "cleanup must not clear the publisher claim, got %v", secondOwnerErr)

	Ok(t, database.ReleasePlanPublicationClaim(testdata.Pull, "publisher"))
	err = <-done

	Ok(t, err)
	workingDir.VerifyWasCalledOnce().Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
	resourceCleaner.VerifyWasCalledOnce().CleanUp(Any[jobs.PullInfo]())
	Equals(t, 1, store.deleteForPullCalls)
	status, statusErr = database.GetPullStatus(testdata.Pull)
	Ok(t, statusErr)
	Assert(t, status != nil, "successful cleanup must retain a durable close tombstone")
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
	Equals(t, "", status.Projects[0].PlanGeneration)
	Ok(t, database.AcquirePlanPublicationClaim(testdata.Pull, "after-cleanup"))
	Ok(t, database.ReleasePlanPublicationClaim(testdata.Pull, "after-cleanup"))
}

func TestCleanUpPullEarlyErrorReleasesPublicationClaim(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	workingDir := mocks.NewMockWorkingDir()
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	pce := events.PullClosedExecutor{
		WorkingDir: workingDir,
		Database:   database,
	}
	workspaceErr := errors.New("disk unavailable")
	When(workingDir.Delete(logger, testdata.GithubRepo, testdata.Pull)).ThenReturn(workspaceErr)

	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)

	Equals(t, "cleaning workspace: disk unavailable", err.Error())
	workingDir.VerifyWasCalledOnce().Delete(logger, testdata.GithubRepo, testdata.Pull)
	Ok(t, database.AcquirePlanPublicationClaim(testdata.Pull, "retry-owner"))
	Ok(t, database.ReleasePlanPublicationClaim(testdata.Pull, "retry-owner"))
}

func TestCleanUpPullCancelsActivePlanGenerationBeforeDestructiveWork(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	workingDir := mocks.NewMockWorkingDir()
	resourceCleaner := mocks.NewMockResourceCleaner()
	store := &countingPlanStore{}
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	project := models.ProjectStatus{
		Workspace:   "default",
		RepoRelDir:  "project-a",
		ProjectName: "a",
	}
	_, err = database.BeginPlanGeneration(testdata.Pull, []models.ProjectStatus{project}, "generation-a")
	Ok(t, err)
	lockerController := gomock.NewController(t)
	locker := lockmocks.NewMockLocker(lockerController)
	locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, nil)
	pce := events.PullClosedExecutor{
		Locker:                   locker,
		WorkingDir:               workingDir,
		Database:                 database,
		PlanStore:                store,
		LogStreamResourceCleaner: resourceCleaner,
	}

	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)

	Ok(t, err)
	workingDir.VerifyWasCalledOnce().Delete(Any[logging.SimpleLogging](), Any[models.Repo](), Any[models.PullRequest]())
	resourceCleaner.VerifyWasCalledOnce().CleanUp(Any[jobs.PullInfo]())
	Equals(t, 1, store.deleteForPullCalls)
	status, statusErr := database.GetPullStatus(testdata.Pull)
	Ok(t, statusErr)
	Assert(t, status != nil, "active generation cleanup must retain a close tombstone")
	Equals(t, models.ErroredPlanStatus, status.Projects[0].Status)
	Equals(t, "", status.Projects[0].PlanGeneration)
	_, staleErr := database.CompletePlanGeneration(testdata.Pull, "generation-a", []command.ProjectResult{{
		Command: command.Plan, Workspace: project.Workspace, RepoRelDir: project.RepoRelDir, ProjectName: project.ProjectName,
		ProjectCommandOutput: command.ProjectCommandOutput{PlanSuccess: &models.PlanSuccess{}},
	}})
	Assert(t, coredb.IsPlanGenerationObsolete(staleErr), "stale generation must remain obsolete after close, got %v", staleErr)
	Ok(t, database.AcquirePlanPublicationClaim(testdata.Pull, "retry-owner"))
	Ok(t, database.ReleasePlanPublicationClaim(testdata.Pull, "retry-owner"))
}

func TestCleanUpPullRetainsClaimWhenFinalCommentPublicationIsAmbiguous(t *testing.T) {
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	workingDir := mocks.NewMockWorkingDir()
	database, err := boltdb.New(t.TempDir())
	Ok(t, err)
	t.Cleanup(func() { database.Close() })
	_, err = database.UpdatePullWithResults(testdata.Pull, []command.ProjectResult{{
		Command:     command.Plan,
		Workspace:   "default",
		RepoRelDir:  "project-a",
		ProjectName: "a",
		ProjectCommandOutput: command.ProjectCommandOutput{
			PlanSuccess: &models.PlanSuccess{},
		},
	}})
	Ok(t, err)
	lockerController := gomock.NewController(t)
	locker := lockmocks.NewMockLocker(lockerController)
	locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return([]models.ProjectLock{{
		Pull:      testdata.Pull,
		Workspace: "default",
		Project:   models.NewProject(testdata.GithubRepo.FullName, "project-a", "a"),
	}}, nil)
	vcsClient := vcsmocks.NewMockClient()
	commentErr := errors.New("ambiguous VCS timeout")
	When(vcsClient.CreateComment(Any[logging.SimpleLogging](), Eq(testdata.GithubRepo), Eq(testdata.Pull.Num), Any[string](), Eq(""))).ThenReturn(commentErr)
	pce := events.PullClosedExecutor{
		Locker:                   locker,
		VCSClient:                vcsClient,
		WorkingDir:               workingDir,
		Database:                 database,
		LogStreamResourceCleaner: mocks.NewMockResourceCleaner(),
	}

	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)

	Equals(t, commentErr, err)
	status, statusErr := database.GetPullStatus(testdata.Pull)
	Ok(t, statusErr)
	Assert(t, status != nil, "claim-aware cleanup must retain a close tombstone before final publication")
	Equals(t, models.DiscardedPlanStatus, status.Projects[0].Status)
	competingErr := database.AcquirePlanPublicationClaim(testdata.Pull, "reopened-pull")
	Assert(t, errors.Is(competingErr, coredb.ErrPlanPublicationBusy), "ambiguous final close comment must retain claim, got %v", competingErr)
}

func TestCleanUpPullWorkspaceErr(t *testing.T) {
	t.Log("when workspace.Delete returns an error, we return it")
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	w := mocks.NewMockWorkingDir()
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
	err = errors.New("err")
	When(w.Delete(logger, testdata.GithubRepo, testdata.Pull)).ThenReturn(err)
	actualErr := pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning workspace: err", actualErr.Error())
}

func TestCleanUpPullWorkspaceErrStillDeletesExternalPlans(t *testing.T) {
	t.Log("when workspace.Delete fails, external plans are still cleaned up")
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	w := mocks.NewMockWorkingDir()
	tmp := t.TempDir()
	db, err := boltdb.New(tmp)
	t.Cleanup(func() {
		db.Close()
	})
	Ok(t, err)
	store := &countingPlanStore{}
	pce := events.PullClosedExecutor{
		WorkingDir:         w,
		PullClosedTemplate: &events.PullClosedEventTemplate{},
		Database:           db,
		PlanStore:          store,
	}
	When(w.Delete(logger, testdata.GithubRepo, testdata.Pull)).ThenReturn(errors.New("disk full"))
	actualErr := pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning workspace: disk full", actualErr.Error())
	Equals(t, 1, store.deleteForPullCalls)
}

func TestCleanUpPullUnlockErr(t *testing.T) {
	t.Log("when locker.UnlockByPull returns an error, we return it")
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)
	w := mocks.NewMockWorkingDir()
	ctrl := gomock.NewController(t)
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
	err = errors.New("err")
	l.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, err)
	actualErr := pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Equals(t, "cleaning up locks: err", actualErr.Error())
}

func TestCleanUpPullNoLocks(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	t.Log("when there are no locks to clean up, we don't comment")
	RegisterMockTestingT(t)
	w := mocks.NewMockWorkingDir()
	ctrl := gomock.NewController(t)
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
			w := mocks.NewMockWorkingDir()
			cp := vcsmocks.NewMockClient()
			ctrl := gomock.NewController(t)
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

		workingDir := mocks.NewMockWorkingDir()
		gmockCtrl := gomock.NewController(t)
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
	RegisterMockTestingT(t)
	logger := logging.NewNoopLogger(t)

	// Create mocks
	workingDir := mocks.NewMockWorkingDir()
	gmockCtrl := gomock.NewController(t)
	locker := lockmocks.NewMockLocker(gmockCtrl)
	client := vcsmocks.NewMockClient()
	resourceCleaner := mocks.NewMockResourceCleaner()

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

	// Setup mock expectations
	locker.EXPECT().UnlockByPull(testdata.GithubRepo.FullName, testdata.Pull.Num).Return(nil, nil)

	// Execute CleanUpPull
	err = pce.CleanUpPull(logger, testdata.GithubRepo, testdata.Pull)
	Ok(t, err)

	// Verify ResourceCleaner.CleanUp was called twice (once for each project)
	resourceCleaner.VerifyWasCalled(Times(2)).CleanUp(Any[jobs.PullInfo]())

	// Get the captured arguments to verify they contain all required fields
	capturedArgs := resourceCleaner.VerifyWasCalled(Times(2)).CleanUp(Any[jobs.PullInfo]()).GetAllCapturedArguments()

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

type countingPlanStore struct {
	deleteForPullCalls int
}

type notifyBusyClaimDatabase struct {
	coredb.Database
	busy chan struct{}
	once sync.Once
}

func (d *notifyBusyClaimDatabase) AcquirePlanPublicationClaim(pull models.PullRequest, token string) error {
	err := d.Database.AcquirePlanPublicationClaim(pull, token)
	if errors.Is(err, coredb.ErrPlanPublicationBusy) {
		d.once.Do(func() { close(d.busy) })
	}
	return err
}

func (s *countingPlanStore) Save(command.ProjectContext, string) error   { return nil }
func (s *countingPlanStore) Load(command.ProjectContext, string) error   { return nil }
func (s *countingPlanStore) Remove(command.ProjectContext, string) error { return nil }
func (s *countingPlanStore) ListWorkspaces(string, string, int) ([]string, error) {
	return nil, nil
}
func (s *countingPlanStore) RestorePlans(string, string, string, int) error { return nil }
func (s *countingPlanStore) DeleteForPull(string, string, int) error {
	s.deleteForPullCalls++
	return nil
}
func (s *countingPlanStore) DeletePlanForProject(string, string, int, string, string, string) error {
	return nil
}
