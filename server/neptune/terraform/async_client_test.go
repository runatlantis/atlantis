package terraform

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testCommandBuilder struct {
	t       *testing.T
	version *version.Version
	path    string
	args    []string

	resp *exec.Cmd
	err  error
}

func (t *testCommandBuilder) Build(v *version.Version, path string, args []string) (*exec.Cmd, error) {
	assert.Equal(t.t, t.version, v)
	assert.Equal(t.t, t.path, path)
	assert.Equal(t.t, t.args, args)

	return t.resp, t.err
}

func TestDefaultClient_RunCommandAsync_Success(t *testing.T) {

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	jobID := "1234"
	echoCommand := exec.Command("sh", "-c", "echo hello")
	// ctx := context.Background()

	testCommandBuilder := &testCommandBuilder{
		t:       t,
		version: nil,
		path:    path,
		args:    args,
		resp:    echoCommand,
		err:     nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}

	testFunc := func(ctx context.Context) (string, error) {
		ch := client.RunCommand(ctx, jobID, path, args, map[string]string{}, nil)
		return waitCh(ch)
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, "hello", out)
}

// Our implementation is bottlenecked on large output due to the way we pipe each line.
func TestDefaultClient_RunCommandAsync_BigOutput(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	jobID := "1234"

	// set up big file to test limitations.
	tmp, cleanup := TempDir(t)
	defer cleanup()

	filename := filepath.Join(tmp, "data")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	assert.Nil(t, err)

	var exp string
	for i := 0; i < 1024; i++ {
		s := strings.Repeat("0", 10) + "\n"
		exp += s
		_, err = f.WriteString(s)
		assert.Nil(t, err)
	}

	cmdStr := fmt.Sprintf("cat %s", filename)
	cat := exec.Command("sh", "-c", cmdStr)

	testCommandBuilder := &testCommandBuilder{
		t:       t,
		version: nil,
		path:    path,
		args:    args,
		resp:    cat,
		err:     nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}

	testFunc := func(ctx context.Context) (string, error) {
		ch := client.RunCommand(ctx, jobID, path, args, map[string]string{}, nil)
		return waitCh(ch)
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, strings.TrimRight(exp, "\n"), out)
}

func TestDefaultClient_RunCommandAsync_StderrOutput(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	jobID := "1234"
	echoCommand := exec.Command("sh", "-c", "echo stderr >&2")

	testCommandBuilder := &testCommandBuilder{
		t:       t,
		version: nil,
		path:    path,
		args:    args,
		resp:    echoCommand,
		err:     nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}
	testFunc := func(ctx context.Context) (string, error) {
		ch := client.RunCommand(ctx, jobID, path, args, map[string]string{}, nil)
		return waitCh(ch)
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, "stderr", out)
}

func TestDefaultClient_RunCommandAsync_ExitOne(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	args := []string{
		"ARG1=$ARG1",
	}
	jobID := "1234"
	echoCommand := exec.Command("sh", "-c", "echo dying && exit 1")

	testCommandBuilder := &testCommandBuilder{
		t:       t,
		version: nil,
		path:    path,
		args:    args,
		resp:    echoCommand,
		err:     nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}

	testFunc := func(ctx context.Context) (string, error) {
		ch := client.RunCommand(ctx, jobID, path, args, map[string]string{}, nil)
		return waitCh(ch)
	}
	env.RegisterActivity(testFunc)
	_, err := env.ExecuteActivity(testFunc)
	assert.ErrorContains(t, err, fmt.Sprintf(`running "/bin/sh -c echo dying && exit 1" in %q: exit status 1`, path))
}

func waitCh(ch <-chan Line) (string, error) {
	var ls []string
	for line := range ch {
		if line.Err != nil {
			return strings.Join(ls, "\n"), line.Err
		}
		ls = append(ls, line.Line)
	}
	return strings.Join(ls, "\n"), nil
}

// TempDir creates a temporary directory and returns its path along
// with a cleanup function to be called via defer, ex:
//   dir, cleanup := TempDir()
//   defer cleanup()
func TempDir(t *testing.T) (string, func()) {
	tmpDir, err := ioutil.TempDir("", "")
	assert.Nil(t, err)
	return tmpDir, func() {
		os.RemoveAll(tmpDir) // nolint: errcheck
	}
}
