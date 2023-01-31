package terraform

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/testsuite"
)

type testCommandBuilder struct {
	t          *testing.T
	version    *version.Version
	path       string
	subCommand *SubCommand

	resp *exec.Cmd
	err  error
}

func (t *testCommandBuilder) Build(ctx context.Context, v *version.Version, path string, subCommand *SubCommand) (*exec.Cmd, error) {
	assert.Equal(t.t, t.version, v)
	assert.Equal(t.t, t.path, path)
	assert.Equal(t.t, t.subCommand, subCommand)

	return t.resp, t.err
}

func TestDefaultClient_RunCommandAsync_Success(t *testing.T) {

	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"

	cmd := NewSubCommand(Plan)
	echoCommand := exec.Command("sh", "-c", "echo hello")

	testCommandBuilder := &testCommandBuilder{
		t:          t,
		version:    nil,
		path:       path,
		subCommand: cmd,
		resp:       echoCommand,
		err:        nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}

	testFunc := func(ctx context.Context) (string, error) {
		buf := &bytes.Buffer{}
		r := &RunCommandRequest{
			RootPath:          path,
			SubCommand:        cmd,
			AdditionalEnvVars: map[string]string{},
		}
		err := client.RunCommand(ctx, r, RunOptions{
			StdOut: buf,
		})
		return buf.String(), err
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, "hello\n", out)
}

func TestDefaultClient_RunCommandAsync_StderrOutput(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	echoCommand := exec.Command("sh", "-c", "echo stderr >&2")

	cmd := NewSubCommand(Plan)
	testCommandBuilder := &testCommandBuilder{
		t:          t,
		version:    nil,
		path:       path,
		subCommand: cmd,
		resp:       echoCommand,
		err:        nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}
	testFunc := func(ctx context.Context) (string, error) {
		buf := &bytes.Buffer{}
		r := &RunCommandRequest{
			RootPath:          path,
			SubCommand:        cmd,
			AdditionalEnvVars: map[string]string{},
		}
		err := client.RunCommand(ctx, r, RunOptions{
			StdOut: buf,
			StdErr: buf,
		})
		return buf.String(), err
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, "stderr\n", out)
}

func TestDefaultClient_RunCommandAsync_EnvValues(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	echoCommand := exec.Command("sh", "-c", "echo $env_key")

	cmd := NewSubCommand(Plan)
	testCommandBuilder := &testCommandBuilder{
		t:          t,
		version:    nil,
		path:       path,
		subCommand: cmd,
		resp:       echoCommand,
		err:        nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}
	testFunc := func(ctx context.Context) (string, error) {
		buf := &bytes.Buffer{}
		r := &RunCommandRequest{
			RootPath:          path,
			SubCommand:        cmd,
			AdditionalEnvVars: map[string]string{"env_key": "env_val"},
		}
		err := client.RunCommand(ctx, r, RunOptions{
			StdOut: buf,
			StdErr: buf,
		})
		return buf.String(), err
	}
	env.RegisterActivity(testFunc)
	resp, err := env.ExecuteActivity(testFunc)
	assert.NoError(t, err)

	var out string
	assert.Nil(t, resp.Get(&out))
	assert.Equal(t, "env_val\n", out)
}

func TestDefaultClient_RunCommandAsync_ExitOne(t *testing.T) {
	ts := testsuite.WorkflowTestSuite{}
	env := ts.NewTestActivityEnvironment()

	path := "some/path"
	echoCommand := exec.Command("sh", "-c", "echo dying && exit 1")

	cmd := NewSubCommand(Plan)
	testCommandBuilder := &testCommandBuilder{
		t:          t,
		version:    nil,
		path:       path,
		subCommand: cmd,
		resp:       echoCommand,
		err:        nil,
	}
	client := &AsyncClient{
		CommandBuilder: testCommandBuilder,
	}

	testFunc := func(ctx context.Context) (string, error) {
		buf := &bytes.Buffer{}
		r := &RunCommandRequest{
			RootPath:          path,
			SubCommand:        cmd,
			AdditionalEnvVars: map[string]string{},
		}
		err := client.RunCommand(ctx, r, RunOptions{
			StdOut: buf,
			StdErr: buf,
		})
		return buf.String(), err
	}

	env.RegisterActivity(testFunc)
	_, err := env.ExecuteActivity(testFunc)
	assert.ErrorContains(t, err, `exit status 1`)
}

// TempDir creates a temporary directory and returns its path along
// with a cleanup function to be called via defer, ex:
//
//	dir, cleanup := TempDir()
//	defer cleanup()
func TempDir(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()
	return tmpDir, func() {
		os.RemoveAll(tmpDir) // nolint: errcheck
	}
}
