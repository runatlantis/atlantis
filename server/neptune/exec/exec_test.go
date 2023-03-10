package exec_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/logging"
	subprocess_exec "github.com/runatlantis/atlantis/server/neptune/exec"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCmd_RunWithNewProcessGroup(t *testing.T) {
	subprocessCmd := subprocess_exec.Command(logging.NewNoopCtxLogger(t), "sleep", "1")
	err := subprocessCmd.RunWithNewProcessGroup(context.Background())
	assert.NoError(t, err)
}

func TestCmd_RunWithNewProcessGroup_CanceledCtx(t *testing.T) {
	subprocessCmd := subprocess_exec.Command(logging.NewNoopCtxLogger(t), "sleep", "10")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := subprocessCmd.RunWithNewProcessGroup(ctx)
	assert.ErrorContains(t, err, "context canceled")
}
