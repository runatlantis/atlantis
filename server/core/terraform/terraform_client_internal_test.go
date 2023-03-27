package terraform

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/stretchr/testify/assert"
)

// Test that it executes successfully
func TestDefaultClient_Synchronous_RunCommandWithVersion(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	logger := logging.NewNoopCtxLogger(t)
	echoCommand := exec.Command("sh", "-c", "echo hello")

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		RequestCtx: context.TODO(),
		Log:        logger,
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	asyncClient := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}
	client := &DefaultClient{
		commandBuilder: mockBuilder,
		AsyncClient:    asyncClient,
	}
	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)

	customEnvVars := map[string]string{}
	out, err := client.RunCommandWithVersion(ctx, prjCtx, path, args, customEnvVars, nil, workspace)
	Ok(t, err)
	Equals(t, "hello\n", out)
}

func TestVersionLoader_buildsURL(t *testing.T) {
	v, _ := version.NewVersion("0.15.0")

	destPath := "some/path"
	fullURL := fmt.Sprintf("https://releases.hashicorp.com/terraform/0.15.0/terraform_0.15.0_%s_%s.zip?checksum=file:https://releases.hashicorp.com/terraform/0.15.0/terraform_0.15.0_SHA256SUMS", runtime.GOOS, runtime.GOARCH)

	RegisterMockTestingT(t)

	mockDownloader := mocks.NewMockDownloader()

	subject := VersionLoader{
		downloader:  mockDownloader,
		downloadURL: "https://releases.hashicorp.com",
	}

	t.Run("success", func(t *testing.T) {
		When(mockDownloader.GetAny(EqString(destPath), EqString(fullURL))).ThenReturn(nil)
		binPath, err := subject.LoadVersion(v, destPath)

		mockDownloader.VerifyWasCalledOnce().GetAny(EqString(destPath), EqString(fullURL))

		Ok(t, err)

		Assert(t, binPath.Resolve() == filepath.Join(destPath, "terraform"), "expected binpath")
	})

	t.Run("error", func(t *testing.T) {

		When(mockDownloader.GetAny(EqString(destPath), EqString(fullURL))).ThenReturn(fmt.Errorf("err"))
		_, err := subject.LoadVersion(v, destPath)

		Assert(t, err != nil, "err is expected")
	})
}

// Test that it returns an error on error.
func TestDefaultClient_Synchronous_RunCommandWithVersion_Error(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	logger := logging.NewNoopCtxLogger(t)
	echoCommand := exec.Command("sh", "-c", "echo dying && exit 1")

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		RequestCtx: context.TODO(),
		Log:        logger,
		BaseRepo: models.Repo{
			FullName: "owner/repo",
			Owner:    "owner",
			Name:     "repo",
		},
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	asyncClient := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}

	client := &DefaultClient{
		commandBuilder: mockBuilder,
		AsyncClient:    asyncClient,
	}

	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)
	out, err := client.RunCommandWithVersion(ctx, prjCtx, path, args, map[string]string{}, nil, workspace)
	assert.ErrorContains(t, err, fmt.Sprintf(`echo dying && exit 1" in %q: exit status 1`, path))
	// Test that we still get our output.
	Equals(t, "dying\n", out)
}
