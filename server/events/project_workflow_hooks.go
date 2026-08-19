// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events/command"
)

// wrapWithProjectWorkflowHooks wraps runner so that project-scoped pre/post
// workflow hooks run around each project command.
func wrapWithProjectWorkflowHooks(
	pre PreWorkflowHooksCommandRunner,
	post PostWorkflowHooksCommandRunner,
	runner prjCmdRunnerFunc,
) prjCmdRunnerFunc {
	if pre == nil || post == nil {
		return runner
	}
	return func(pctx command.ProjectContext) command.ProjectCommandOutput {
		if err := pre.RunPreHooksForProject(pctx); err != nil {
			return command.ProjectCommandOutput{
				Failure: fmt.Sprintf("pre-workflow hook failed: %s", err),
				Error:   err,
			}
		}

		out := runner(pctx)

		commandHasErrors := out.Error != nil || out.Failure != ""
		_ = post.RunPostHooksForProject(pctx, commandHasErrors) // nolint: errcheck

		return out
	}
}
