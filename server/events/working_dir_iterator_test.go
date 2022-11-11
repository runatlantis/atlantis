package events_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v45/github"
	"github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"
	eventmocks "github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/vcs"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestListCurrentWorkingDirPulls(t *testing.T) {

	mockGHClient := vcsmocks.NewMockGithubPullRequestGetter()
	mockEventParser := eventmocks.NewMockEventParsing()
	log := logging.NewNoopCtxLogger(t)

	t.Run("repos subdir not exist", func(t *testing.T) {

		baseDir := t.TempDir()

		subject := &events.FileWorkDirIterator{
			Log:          log,
			GithubClient: mockGHClient,
			EventParser:  mockEventParser,
			DataDir:      baseDir,
		}

		pulls, err := subject.ListCurrentWorkingDirPulls()

		assert.Nil(t, err)
		assert.Empty(t, pulls)
	})

	t.Run("pull not found", func(t *testing.T) {

		baseDir := t.TempDir()

		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo1", "1", "default"), os.ModePerm)

		pullNotFound := &vcs.PullRequestNotFound{Err: errors.New("error")}

		pegomock.When(mockGHClient.GetPullRequestFromName("repo1", "nish", 1)).ThenReturn(nil, pullNotFound)

		subject := &events.FileWorkDirIterator{
			Log:          log,
			GithubClient: mockGHClient,
			EventParser:  mockEventParser,
			DataDir:      baseDir,
		}

		pulls, err := subject.ListCurrentWorkingDirPulls()

		assert.NoError(t, err)
		assert.Empty(t, pulls)
	})

	t.Run("1 pull returned", func(t *testing.T) {

		pullNum := 1

		expectedGithubPull := &github.PullRequest{
			Number: &pullNum,
		}
		expectedInternalPull := models.PullRequest{
			Num: pullNum,
		}

		baseDir := t.TempDir()

		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo1", "1", "default"), os.ModePerm)

		pegomock.When(mockGHClient.GetPullRequestFromName("repo1", "nish", 1)).ThenReturn(expectedGithubPull, nil)
		pegomock.When(mockEventParser.ParseGithubPull(expectedGithubPull)).ThenReturn(expectedInternalPull, models.Repo{}, models.Repo{}, nil)

		subject := &events.FileWorkDirIterator{
			Log:          log,
			GithubClient: mockGHClient,
			EventParser:  mockEventParser,
			DataDir:      baseDir,
		}

		pulls, err := subject.ListCurrentWorkingDirPulls()

		assert.Nil(t, err)
		assert.Len(t, pulls, 1)
		assert.Contains(t, pulls, expectedInternalPull)
	})

	t.Run("2 pulls same repo", func(t *testing.T) {

		pullNum1 := 1

		expectedGithubPull1 := &github.PullRequest{
			Number: &pullNum1,
		}
		expectedInternalPull1 := models.PullRequest{
			Num: pullNum1,
		}

		pullNum2 := 2

		expectedGithubPull2 := &github.PullRequest{
			Number: &pullNum2,
		}
		expectedInternalPull2 := models.PullRequest{
			Num: pullNum2,
		}

		baseDir := t.TempDir()

		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo1", "1", "default"), os.ModePerm)
		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo1", "2", "default"), os.ModePerm)

		pegomock.When(mockGHClient.GetPullRequestFromName("repo1", "nish", pullNum1)).ThenReturn(expectedGithubPull1, nil)
		pegomock.When(mockGHClient.GetPullRequestFromName("repo1", "nish", pullNum2)).ThenReturn(expectedGithubPull2, nil)
		pegomock.When(mockEventParser.ParseGithubPull(expectedGithubPull1)).ThenReturn(expectedInternalPull1, models.Repo{}, models.Repo{}, nil)
		pegomock.When(mockEventParser.ParseGithubPull(expectedGithubPull2)).ThenReturn(expectedInternalPull2, models.Repo{}, models.Repo{}, nil)

		subject := &events.FileWorkDirIterator{
			Log:          log,
			GithubClient: mockGHClient,
			EventParser:  mockEventParser,
			DataDir:      baseDir,
		}

		pulls, err := subject.ListCurrentWorkingDirPulls()

		assert.Nil(t, err)
		assert.Len(t, pulls, 2)
		assert.Contains(t, pulls, expectedInternalPull1)
		assert.Contains(t, pulls, expectedInternalPull2)
	})

	t.Run("2 pulls multiple repos", func(t *testing.T) {

		pullNum1 := 1

		expectedGithubPull1 := &github.PullRequest{
			Number: &pullNum1,
		}
		expectedInternalPull1 := models.PullRequest{
			Num: pullNum1,
		}

		pullNum2 := 2

		expectedGithubPull2 := &github.PullRequest{
			Number: &pullNum2,
		}
		expectedInternalPull2 := models.PullRequest{
			Num: pullNum2,
		}

		baseDir := t.TempDir()

		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo1", "1", "default"), os.ModePerm)
		_ = os.MkdirAll(filepath.Join(baseDir, "repos", "nish", "repo2", "2", "default"), os.ModePerm)

		pegomock.When(mockGHClient.GetPullRequestFromName("repo1", "nish", pullNum1)).ThenReturn(expectedGithubPull1, nil)
		pegomock.When(mockGHClient.GetPullRequestFromName("repo2", "nish", pullNum2)).ThenReturn(expectedGithubPull2, nil)
		pegomock.When(mockEventParser.ParseGithubPull(expectedGithubPull1)).ThenReturn(expectedInternalPull1, models.Repo{}, models.Repo{}, nil)
		pegomock.When(mockEventParser.ParseGithubPull(expectedGithubPull2)).ThenReturn(expectedInternalPull2, models.Repo{}, models.Repo{}, nil)

		subject := &events.FileWorkDirIterator{
			Log:          log,
			GithubClient: mockGHClient,
			EventParser:  mockEventParser,
			DataDir:      baseDir,
		}

		pulls, err := subject.ListCurrentWorkingDirPulls()

		assert.Nil(t, err)
		assert.Len(t, pulls, 2)
		assert.Contains(t, pulls, expectedInternalPull1)
		assert.Contains(t, pulls, expectedInternalPull2)
	})

}
