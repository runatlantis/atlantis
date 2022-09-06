package link_test

import (
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github/link"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/stretchr/testify/assert"
	httpurl "net/url"
	"testing"
)

func Test_BuildDownloadLinkFromArchive(t *testing.T) {
	expectedURL := "https://github.com/testowner/testrepo/legacy.zip/refs/heads/main//testowner-testrepo-a1b2c3d/test/path?archive=zip&token=testtoken123"
	testRoot := root.Root{
		Path: "test/path",
	}
	testCommit := github.Commit{
		Ref: "a1b2c3d",
	}
	testRepo := github.Repo{
		HeadCommit: testCommit,
		Owner:      "testowner",
		Name:       "testrepo",
	}
	archiveURL, err := httpurl.Parse("https://github.com/testowner/testrepo/legacy.zip/refs/heads/main?token=testtoken123")
	assert.NoError(t, err)
	linkBuilder := link.Builder{}
	downloadLink := linkBuilder.BuildDownloadLinkFromArchive(archiveURL, testRoot, testRepo)
	assert.Equal(t, expectedURL, downloadLink)
}

func Test_BuildDownloadLinkFromArchive_NoToken(t *testing.T) {
	expectedURL := "https://github.com/testowner/testrepo/legacy.zip/refs/heads/main//testowner-testrepo-a1b2c3d/test/path?archive=zip"
	testRoot := root.Root{
		Path: "/test/path",
	}
	testCommit := github.Commit{
		Ref: "a1b2c3d",
	}
	testRepo := github.Repo{
		HeadCommit: testCommit,
		Owner:      "testowner",
		Name:       "testrepo",
	}
	archiveURL, err := httpurl.Parse("https://github.com/testowner/testrepo/legacy.zip/refs/heads/main")
	assert.NoError(t, err)
	linkBuilder := link.Builder{}
	downloadLink := linkBuilder.BuildDownloadLinkFromArchive(archiveURL, testRoot, testRepo)
	assert.Equal(t, expectedURL, downloadLink)
}
