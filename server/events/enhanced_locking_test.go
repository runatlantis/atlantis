package events

import (
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/locking"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimal mockLocker for TryLockWithRetry
// always returns LockAcquired: false, CurrLock.Pull.Num: 2
// so the memory lock logic is exercised but no panic

type mockLocker struct {
	callCount int
}

func (m *mockLocker) TryLock(project models.Project, workspace string, pull models.PullRequest, user models.User) (locking.TryLockResponse, error) {
	m.callCount++
	if m.callCount == 1 {
		return locking.TryLockResponse{
			LockAcquired: true,
			LockKey:      "lock-key",
		}, nil
	}
	return locking.TryLockResponse{
		LockAcquired: false,
		CurrLock:     models.ProjectLock{Pull: models.PullRequest{Num: 1}},
	}, nil
}
func (m *mockLocker) Unlock(lockKey string) (*models.ProjectLock, error) { return nil, nil }
func (m *mockLocker) List() (map[string]models.ProjectLock, error)       { return nil, nil }
func (m *mockLocker) UnlockByPull(repoFullName string, pullNum int) ([]models.ProjectLock, error) {
	return nil, nil
}
func (m *mockLocker) GetLock(lockKey string) (*models.ProjectLock, error) { return nil, nil }

// minimal mockVCSClient for MarkdownPullLink
// always returns a dummy link

type mockVCSClient struct{}

func (m *mockVCSClient) MarkdownPullLink(pull models.PullRequest) (string, error) {
	return "dummy-link", nil
}
func (m *mockVCSClient) CreateComment(log logging.SimpleLogging, repo models.Repo, pullNum int, comment string, command string) error {
	return nil
}
func (m *mockVCSClient) DiscardReviews(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest) error {
	return nil
}
func (m *mockVCSClient) UpdateStatus(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest, status models.CommitStatus, src, description, url string) error {
	return nil
}
func (m *mockVCSClient) MergePull(log logging.SimpleLogging, pull models.PullRequest, opts models.PullRequestOptions) error {
	return nil
}
func (m *mockVCSClient) GetModifiedFiles(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return nil, nil
}
func (m *mockVCSClient) GetPullRequest(repo models.Repo, num int) (models.PullRequest, error) {
	return models.PullRequest{}, nil
}
func (m *mockVCSClient) IsMergeable(repo models.Repo, pull models.PullRequest) (bool, error) {
	return true, nil
}
func (m *mockVCSClient) ApprovePull(log logging.SimpleLogging, repo models.Repo, pullNum int) error {
	return nil
}
func (m *mockVCSClient) UnapprovePull(log logging.SimpleLogging, repo models.Repo, pullNum int) error {
	return nil
}
func (m *mockVCSClient) GetPullLabels(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest) ([]string, error) {
	return nil, nil
}
func (m *mockVCSClient) AddPullLabel(repo models.Repo, pullNum int, label string) error { return nil }
func (m *mockVCSClient) RemovePullLabel(repo models.Repo, pullNum int, label string) error {
	return nil
}
func (m *mockVCSClient) GetUserName() (string, error) { return "", nil }
func (m *mockVCSClient) GetTeamNamesForUser(log logging.SimpleLogging, repo models.Repo, user models.User) ([]string, error) {
	return nil, nil
}
func (m *mockVCSClient) GetCloneURL(log logging.SimpleLogging, hostType models.VCSHostType, cloneURL string) (string, error) {
	return cloneURL, nil
}
func (m *mockVCSClient) GetFileContent(log logging.SimpleLogging, pull models.PullRequest, filePath string) (bool, []byte, error) {
	return true, []byte{}, nil
}
func (m *mockVCSClient) HidePrevCommandComments(log logging.SimpleLogging, repo models.Repo, pullNum int, command string, workspace string) error {
	return nil
}
func (m *mockVCSClient) PullIsApproved(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest) (models.ApprovalStatus, error) {
	return models.ApprovalStatus{}, nil
}
func (m *mockVCSClient) PullIsMergeable(log logging.SimpleLogging, repo models.Repo, pull models.PullRequest, branch string, requiredStatuses []string) (bool, error) {
	return true, nil
}
func (m *mockVCSClient) ReactToComment(log logging.SimpleLogging, repo models.Repo, pullNum int, commentID int64, reaction string) error {
	return nil
}
func (m *mockVCSClient) SupportsSingleFileDownload(repo models.Repo) bool { return true }

func TestEnhancedLockingSystem_ProtectWorkingDir(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	enhancedLocking := NewEnhancedLockingSystem(
		nil, nil, nil, logger, nil, false, false, 3, 1,
	)
	repoFullName := "test/repo"
	pullNum := 1
	workspace := "default"
	cancel := enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, workspace)
	assert.True(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, workspace))
	cancel()
	enhancedLocking.CleanupWorkingDirProtection(repoFullName, pullNum, workspace)
	assert.False(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, workspace))
}

func TestEnhancedLockingSystem_CleanupAllLocks(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	enhancedLocking := NewEnhancedLockingSystem(
		nil, nil, nil, logger, nil, false, false, 3, 1,
	)
	repoFullName := "test/repo"
	pullNum := 1
	enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, "default")
	enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, "staging")
	err := enhancedLocking.CleanupAllLocks(repoFullName, pullNum)
	require.NoError(t, err)
	assert.False(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "default"))
	assert.False(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "staging"))
}

func TestEnhancedLockingSystem_MemoryLockPreventsRaceConditions(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	mock := &mockLocker{}
	mockVCS := &mockVCSClient{}
	enhancedLocking := NewEnhancedLockingSystem(
		mock, nil, mockVCS, logger, nil, false, false, 3, 1,
	)
	project := models.Project{RepoFullName: "test/repo", Path: "."}
	workspace := "default"
	pull1 := models.PullRequest{Num: 1}
	pull2 := models.PullRequest{Num: 2}
	user := models.User{Username: "testuser"}

	// Channel to coordinate the test
	firstLockAcquired := make(chan bool, 1)
	secondAttemptDone := make(chan *TryLockResponse, 1)
	secondErrorDone := make(chan error, 1)

	// First goroutine: acquire lock and hold it
	go func() {
		resp, err := enhancedLocking.TryLockWithRetry(project, workspace, pull1, user)
		if err != nil {
			firstLockAcquired <- false
			return
		}
		firstLockAcquired <- true

		// Hold the lock for a short time to ensure second attempt happens while locked
		time.Sleep(100 * time.Millisecond)

		// Release the lock
		if resp.UnlockFn != nil {
			_ = resp.UnlockFn()
		}
	}()

	// Second goroutine: try to acquire lock while first is held
	go func() {
		// Wait a bit to ensure first lock is acquired
		time.Sleep(50 * time.Millisecond)

		resp2, err2 := enhancedLocking.TryLockWithRetry(project, workspace, pull2, user)
		secondAttemptDone <- resp2
		secondErrorDone <- err2
	}()

	// Wait for results with timeout
	select {
	case firstSuccess := <-firstLockAcquired:
		assert.True(t, firstSuccess, "First lock should be acquired successfully")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for first lock acquisition")
	}

	select {
	case resp2 := <-secondAttemptDone:
		err2 := <-secondErrorDone
		require.NoError(t, err2)
		assert.False(t, resp2.LockAcquired)
		assert.Contains(t, resp2.LockFailureReason, "Another operation is in progress for this project/workspace")
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for second lock attempt")
	}
}

func TestEnhancedLockingSystem_QueueKeyGeneration(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	enhancedLocking := NewEnhancedLockingSystem(
		nil, nil, nil, logger, nil, false, false, 3, 1,
	)
	project := models.Project{RepoFullName: "test/repo", Path: "terraform"}
	workspace := "default"
	key := enhancedLocking.memoryLockKey(project, workspace)
	expectedKey := "memory:test/repo:terraform:default"
	assert.Equal(t, expectedKey, key)
}

func TestEnhancedLockingSystem_WorkingDirProtectionIsolation(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	enhancedLocking := NewEnhancedLockingSystem(
		nil, nil, nil, logger, nil, false, false, 3, 1,
	)
	repoFullName := "test/repo"
	pullNum := 1
	enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, "default")
	enhancedLocking.ProtectWorkingDir(repoFullName, pullNum, "staging")
	assert.True(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "default"))
	assert.True(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "staging"))
	enhancedLocking.CleanupWorkingDirProtection(repoFullName, pullNum, "default")
	assert.False(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "default"))
	assert.True(t, enhancedLocking.IsWorkingDirProtected(repoFullName, pullNum, "staging"))
}
