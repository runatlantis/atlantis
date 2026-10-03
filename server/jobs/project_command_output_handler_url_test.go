// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package jobs_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectOutputPullRequestURL(t *testing.T) {
	for _, requestURL := range []string{"", "https://github.example.com/org/repo/pull/1", "https://gitlab.example.com/group/repo/-/merge_requests/1"} {
		for _, hook := range []bool{false, true} {
			ctx := createTestProjectCmdContext(t)
			ctx.Pull.URL = requestURL
			lines := make(chan *jobs.ProjectCmdOutputLine, 1)
			handler := jobs.NewAsyncProjectCommandOutputHandler(lines, logging.NewNoopLogger(t))
			key := jobs.PullInfo{PullNum: ctx.Pull.Num, Repo: ctx.BaseRepo.Name, RepoFullName: ctx.BaseRepo.FullName}
			if hook {
				handler.SendWorkflowHook(models.WorkflowHookCommandContext{BaseRepo: ctx.BaseRepo, Pull: ctx.Pull, HookID: ctx.JobID, HookStepName: "pre_workflow", HookDescription: "Workflow hook"}, "output", false)
			} else {
				key.Workspace, key.Path, key.ProjectName = ctx.Workspace, ctx.RepoRelDir, ctx.ProjectName
				handler.Send(ctx, "output", false)
			}
			line := <-lines
			Equals(t, requestURL, line.JobInfo.PullRequestURL)
			Equals(t, key, line.JobInfo.PullInfo)
			lines <- line
			close(lines)
			handler.Handle()
			mappings := handler.GetPullToJobMapping()
			Equals(t, 1, len(mappings))
			Equals(t, key, mappings[0].Pull)
			Equals(t, requestURL, mappings[0].JobIDInfos[0].PullRequestURL)
			handler.CleanUp(key)
			Equals(t, 0, len(handler.GetPullToJobMapping()))
			Equals(t, jobs.OutputBuffer{}, handler.(*jobs.AsyncProjectCommandOutputHandler).GetProjectOutputBuffer(ctx.JobID))
		}
	}
}
