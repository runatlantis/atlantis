package link_test

import (
	httpurl "net/url"
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github/link"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
	"github.com/stretchr/testify/assert"
)

func Test_BuildDownloadLinkFromArchive(t *testing.T) {
	expectedURL := "https://github.com/testowner/testrepo/legacy.zip/refs/heads/main//testowner-testrepo-a1b2c3d?archive=zip&token=testtoken123"
	testRoot := terraform.Root{
		Path: "test/path",
	}
	testRepo := github.Repo{
		Owner: "testowner",
		Name:  "testrepo",
	}
	testRevision := "a1b2c3d"
	archiveURL, err := httpurl.Parse("https://github.com/testowner/testrepo/legacy.zip/refs/heads/main?token=testtoken123")
	assert.NoError(t, err)
	linkBuilder := link.Builder{}
	downloadLink := linkBuilder.BuildDownloadLinkFromArchive(archiveURL, testRoot, testRepo, testRevision)
	assert.Equal(t, expectedURL, downloadLink)
}

func Test_BuildDownloadLinkFromArchive_NoToken(t *testing.T) {
	expectedURL := "https://github.com/testowner/testrepo/legacy.zip/refs/heads/main//testowner-testrepo-a1b2c3d?archive=zip"
	testRoot := terraform.Root{
		Path: "/test/path",
	}
	testRepo := github.Repo{
		Owner: "testowner",
		Name:  "testrepo",
	}
	testRevision := "a1b2c3d"
	archiveURL, err := httpurl.Parse("https://github.com/testowner/testrepo/legacy.zip/refs/heads/main")
	assert.NoError(t, err)
	linkBuilder := link.Builder{}
	downloadLink := linkBuilder.BuildDownloadLinkFromArchive(archiveURL, testRoot, testRepo, testRevision)
	assert.Equal(t, expectedURL, downloadLink)
}
