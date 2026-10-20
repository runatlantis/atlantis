// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestWorkflowHookRunner_StatusReadError(t *testing.T) {
	runners := []struct {
		name string
		run  func(models.WorkflowHookCommandContext, string, string, string, string) (string, string, error)
	}{
		{"pre", (runtime.DefaultPreWorkflowHookRunner{}).Run},
		{"post", (runtime.DefaultPostWorkflowHookRunner{}).Run},
	}
	for _, runner := range runners {
		t.Run(runner.name, func(t *testing.T) {
			dir := t.TempDir()
			statusPath := filepath.Join(dir, "OUTPUT_STATUS_FILE")
			// A directory passes Stat but cannot be read as a status file.
			Ok(t, os.Mkdir(statusPath, 0700))
			ctx := models.WorkflowHookCommandContext{Log: logging.NewNoopLogger(t), SuppressJobOutput: true}

			output, description, err := runner.run(ctx, "printf 'hook output'", "sh", "-c", dir)

			Equals(t, "hook output", output)
			Equals(t, "", description)
			pathErr, ok := errors.AsType[*os.PathError](err)
			Assert(t, ok, "expected status file error cause, got %v", err)
			Equals(t, statusPath, pathErr.Path)
			Equals(t, "read", pathErr.Op)
			ErrContains(t, "sh -c printf 'hook output'", err)
			ErrContains(t, "\nhook output", err)
			Assert(t, strings.Contains(err.Error(), "reading workflow hook status file"), "missing status file context: %v", err)
		})
	}
}
