// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"sort"
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/command"
)

type prjCmdRunnerFunc func(ctx command.ProjectContext) command.ProjectCommandOutput

func RunOneProjectCmd(
	runnerFunc prjCmdRunnerFunc,
	cmd command.ProjectContext,
) command.ProjectResult {
	projectCommandOutput := runnerFunc(cmd)

	return command.ProjectResult{
		ProjectCommandOutput: projectCommandOutput,
		Command:              cmd.CommandName,
		SubCommand:           cmd.SubCommand,
		RepoRelDir:           cmd.RepoRelDir,
		Workspace:            cmd.Workspace,
		ProjectName:          cmd.ProjectName,
		SilencePRComments:    cmd.SilencePRComments,
	}
}

func runProjectCmdsParallel(
	cmds []command.ProjectContext,
	runnerFunc prjCmdRunnerFunc,
	poolSize int,
) command.Result {
	var results []command.ProjectResult
	mux := &sync.Mutex{}

	wg := sizedwaitgroup.New(poolSize)
	for _, pCmd := range cmds {
		pCmd := pCmd
		var execute func()
		wg.Add()

		execute = func() {
			defer wg.Done()
			res := RunOneProjectCmd(runnerFunc, pCmd)
			mux.Lock()
			results = append(results, res)
			mux.Unlock()
		}

		go execute()
	}

	wg.Wait()
	return command.Result{ProjectResults: results}
}

func runProjectCmds(
	cmds []command.ProjectContext,
	runnerFunc prjCmdRunnerFunc,
) command.Result {
	var results []command.ProjectResult
	for _, pCmd := range cmds {
		res := RunOneProjectCmd(runnerFunc, pCmd)

		results = append(results, res)
	}
	return command.Result{ProjectResults: results}
}

func splitByExecutionOrderGroup(cmds []command.ProjectContext) [][]command.ProjectContext {
	groups := make(map[int][]command.ProjectContext)
	for _, cmd := range cmds {
		groups[cmd.ExecutionOrderGroup] = append(groups[cmd.ExecutionOrderGroup], cmd)
	}

	var groupKeys []int
	for k := range groups {
		groupKeys = append(groupKeys, k)
	}
	sort.Ints(groupKeys)

	var res [][]command.ProjectContext
	for _, group := range groupKeys {
		res = append(res, groups[group])
	}
	return res
}

func runProjectCmdsParallelGroups(
	ctx *command.Context,
	cmds []command.ProjectContext,
	runnerFunc prjCmdRunnerFunc,
	poolSize int,
) command.Result {
	var results []command.ProjectResult
	groups := splitByExecutionOrderGroup(cmds)
	for _, group := range groups {
		res := runProjectCmdsParallel(group, runnerFunc, poolSize)
		results = append(results, res.ProjectResults...)
		if res.HasErrors() && group[0].AbortOnExecutionOrderFail {
			ctx.Log.Info("abort on execution order when failed")
			break
		}
	}

	return command.Result{ProjectResults: results}
}
