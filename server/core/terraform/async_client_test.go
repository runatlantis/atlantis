package terraform

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/core/terraform/helpers"
	"github.com/runatlantis/atlantis/server/core/terraform/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	jobmocks "github.com/runatlantis/atlantis/server/jobs/mocks"
	"github.com/stretchr/testify/assert"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultClient_RunCommandAsync_Success(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	logger := logging.NewNoopCtxLogger(t)
	echoCommand := exec.Command("sh", "-c", "echo hello")

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log:        logger,
		RequestCtx: context.TODO(),
	}

	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	client := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}

	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)
	outCh := client.RunCommandAsync(ctx, prjCtx, path, args, map[string]string{}, nil, workspace)

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "hello", out)
}

// Our implementation is bottlenecked on large output due to the way we pipe each line.
func TestDefaultClient_RunCommandAsync_BigOutput(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	logger := logging.NewNoopCtxLogger(t)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log:        logger,
		RequestCtx: context.TODO(),
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	client := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}

	// set up big file to test limitations.
	tmp, cleanup := TempDir(t)
	defer cleanup()

	filename := filepath.Join(tmp, "data")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	Ok(t, err)

	var exp string
	for i := 0; i < 1024; i++ {
		s := strings.Repeat("0", 10) + "\n"
		exp += s
		_, err = f.WriteString(s)
		Ok(t, err)
	}

	cmdStr := fmt.Sprintf("cat %s", filename)
	cat := exec.Command("sh", "-c", cmdStr)

	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(cat, nil)
	outCh := client.RunCommandAsync(ctx, prjCtx, path, args, map[string]string{}, nil, workspace)

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, strings.TrimRight(exp, "\n"), out)
}

func TestDefaultClient_RunCommandAsync_StderrOutput(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	echoCommand := exec.Command("sh", "-c", "echo stderr >&2")

	logger := logging.NewNoopCtxLogger(t)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log:        logger,
		RequestCtx: context.TODO(),
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	client := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}
	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)
	outCh := client.RunCommandAsync(ctx, prjCtx, path, args, map[string]string{}, nil, workspace)

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "stderr", out)
}

func TestDefaultClient_RunCommandAsync_ExitOne(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	echoCommand := exec.Command("sh", "-c", "echo dying && exit 1")
	logger := logging.NewNoopCtxLogger(t)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log:        logger,
		RequestCtx: context.TODO(),
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	client := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}
	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)
	outCh := client.RunCommandAsync(ctx, prjCtx, path, args, map[string]string{}, nil, workspace)

	out, err := waitCh(outCh)
	assert.ErrorContains(t, err, fmt.Sprintf(`echo dying && exit 1" in %q: exit status 1`, path))
	// Test that we still get our output.
	Equals(t, "dying", out)
}

func TestDefaultClient_RunCommandAsync_Input(t *testing.T) {
	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	workspace := "workspace"
	echoCommand := exec.Command("sh", "-c", "read a && echo $a")
	logger := logging.NewNoopCtxLogger(t)

	ctx := context.Background()
	prjCtx := command.ProjectContext{
		Log:        logger,
		RequestCtx: context.TODO(),
	}
	mockBuilder := mocks.NewMockcommandBuilder()
	projectCmdOutputHandler := jobmocks.NewMockProjectCommandOutputHandler()
	client := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          mockBuilder,
	}

	inCh := make(chan string)

	When(mockBuilder.Build(nil, workspace, path, args)).ThenReturn(echoCommand, nil)
	outCh := client.RunCommandAsyncWithInput(ctx, prjCtx, path, args, map[string]string{}, nil, workspace, inCh)
	inCh <- "echo me\n"

	out, err := waitCh(outCh)
	Ok(t, err)
	Equals(t, "echo me", out)
}

func waitCh(ch <-chan helpers.Line) (string, error) {
	var ls []string
	for line := range ch {
		if line.Err != nil {
			return strings.Join(ls, "\n"), line.Err
		}
		ls = append(ls, line.Line)
	}
	return strings.Join(ls, "\n"), nil
}
