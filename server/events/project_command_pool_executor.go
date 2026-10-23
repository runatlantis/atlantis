// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"runtime/debug"
	"sort"
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
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
	cancellationTracker CancellationTracker,
	pull models.PullRequest,
) command.Result {
	var results []command.ProjectResult
	mux := &sync.Mutex{}

	wg := sizedwaitgroup.New(poolSize)
	cancelledAt := -1

	for i, pCmd := range cmds {
		if cancellationTracker != nil && cancellationTracker.IsCancelled(pull) {
			cancelledAt = i
			break
		}

		wg.Add()

		if cancellationTracker != nil && cancellationTracker.IsCancelled(pull) {
			// because wg.Add() is a blocking call, we check for cancellation again after it returns
			// to avoid waiting for another goroutine to finish before we can exit early.
			// see used sizedwaitgroup package:
			// https://github.com/remeh/sizedwaitgroup/blob/b77873a44db20b1b82a5d60e698ceea3270d09d0/sizedwaitgroup.go#L50
			wg.Done()
			cancelledAt = i
			break
		}

		execute := func(cmd command.ProjectContext) {
			defer wg.Done()
			res := RunOneProjectCmd(runnerFunc, cmd)
			mux.Lock()
			results = append(results, res)
			mux.Unlock()
		}

		go execute(pCmd)
	}

	wg.Wait()

	if cancelledAt != -1 {
		for _, pCmd := range cmds[cancelledAt:] {
			results = append(results, command.ProjectResult{
				Command: pCmd.CommandName,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: fmt.Errorf("operation cancelled via `atlantis cancel` command"),
				},
				RepoRelDir:  pCmd.RepoRelDir,
				Workspace:   pCmd.Workspace,
				ProjectName: pCmd.ProjectName,
			})
		}
	}

	return command.Result{ProjectResults: results}
}

// RunProjectCmdsParallelInOrder runs up to poolSize cmds at once. A cmd starts
// only after every earlier cmd from a different execution order group has
// finished, and every cmd gets a result, returned in cmds order.
//
// Like the PR path, it stops running groups once a group with
// AbortOnExecutionOrderFail set has an error. The cmds that did not run get an
// errored result, so callers still see one result per cmd. A panic in a runner
// becomes an errored result for that cmd instead of crashing the server.
func RunProjectCmdsParallelInOrder(
	log logging.SimpleLogging,
	cmds []command.ProjectContext,
	runnerFunc prjCmdRunnerFunc,
	poolSize int,
) command.Result {
	results := make([]command.ProjectResult, len(cmds))
	wg := sizedwaitgroup.New(poolSize)
	groupStart := 0
	aborted := false
	for i, cmd := range cmds {
		if i > 0 && cmd.ExecutionOrderGroup != cmds[i-1].ExecutionOrderGroup {
			wg.Wait()
			if !aborted && cmds[i-1].AbortOnExecutionOrderFail && (command.Result{ProjectResults: results[groupStart:i]}).HasErrors() {
				log.Info("abort on execution order when failed")
				aborted = true
			}
			groupStart = i
		}
		if aborted {
			results[i] = errorProjectResult(cmd, fmt.Errorf("skipped because an earlier execution order group failed"))
			continue
		}
		wg.Add()
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					log.Err("panic running %s for project %q (dir %q, workspace %q): %v\n%s", cmd.CommandName, cmd.ProjectName, cmd.RepoRelDir, cmd.Workspace, r, debug.Stack())
					results[i] = errorProjectResult(cmd, fmt.Errorf("panic: %v", r))
				}
			}()
			results[i] = RunOneProjectCmd(runnerFunc, cmd)
		}()
	}
	wg.Wait()
	return command.Result{ProjectResults: results}
}

func errorProjectResult(cmd command.ProjectContext, err error) command.ProjectResult {
	return command.ProjectResult{
		ProjectCommandOutput: command.ProjectCommandOutput{Error: err},
		Command:              cmd.CommandName,
		SubCommand:           cmd.SubCommand,
		RepoRelDir:           cmd.RepoRelDir,
		Workspace:            cmd.Workspace,
		ProjectName:          cmd.ProjectName,
		SilencePRComments:    cmd.SilencePRComments,
	}
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
	cancellationTracker CancellationTracker,
) command.Result {
	var results []command.ProjectResult
	groups := splitByExecutionOrderGroup(cmds)
	for _, group := range groups {
		res := runProjectCmdsParallel(group, runnerFunc, poolSize, cancellationTracker, ctx.Pull)
		results = append(results, res.ProjectResults...)
		if res.HasErrors() && group[0].AbortOnExecutionOrderFail {
			ctx.Log.Info("abort on execution order when failed")
			break
		}
	}

	return command.Result{ProjectResults: results}
}

func runProjectCmdsWithCancellationTracker(
	ctx *command.Context,
	projectCmds []command.ProjectContext,
	cancellationTracker CancellationTracker,
	parallelPoolSize int,
	isParallel bool,
	runnerFunc prjCmdRunnerFunc,
) command.Result {
	if isParallel {
		ctx.Log.Info("Running commands in parallel")
	}

	groups := prepareExecutionGroups(projectCmds, isParallel)
	if cancellationTracker != nil {
		defer cancellationTracker.Clear(ctx.Pull)
	}

	var results []command.ProjectResult
	for i, group := range groups {
		if i > 0 && cancellationTracker != nil && cancellationTracker.IsCancelled(ctx.Pull) {
			ctx.Log.Info("Skipping execution order group %d and all subsequent groups due to cancellation", group[0].ExecutionOrderGroup)
			results = append(results, createCancelledResults(groups[i:])...)
			break
		}

		var groupResult command.Result
		if isParallel && len(group) > 1 {
			groupResult = runProjectCmdsParallel(group, runnerFunc, parallelPoolSize, cancellationTracker, ctx.Pull)
		} else {
			groupResult = runProjectCmds(group, runnerFunc)
		}
		results = append(results, groupResult.ProjectResults...)

		if groupResult.HasErrors() && group[0].AbortOnExecutionOrderFail && isParallel {
			ctx.Log.Info("abort on execution order when failed")
			break
		}

		if ctx.PullStatus != nil {
			for _, result := range groupResult.ProjectResults {
				for projectIdx := range ctx.PullStatus.Projects {
					if result.Workspace == ctx.PullStatus.Projects[projectIdx].Workspace &&
						result.RepoRelDir == ctx.PullStatus.Projects[projectIdx].RepoRelDir &&
						result.ProjectName == ctx.PullStatus.Projects[projectIdx].ProjectName {
						ctx.PullStatus.Projects[projectIdx].Status = result.PlanStatus()
						break
					}
				}
			}
		}
	}

	return command.Result{ProjectResults: results}
}

func prepareExecutionGroups(
	projectCmds []command.ProjectContext,
	isParallel bool,
) [][]command.ProjectContext {
	groups := splitByExecutionOrderGroup(projectCmds)
	if len(groups) == 1 && !isParallel {
		return createIndividualCommandGroups(projectCmds)
	}
	return groups
}

func createIndividualCommandGroups(projectCmds []command.ProjectContext) [][]command.ProjectContext {
	groups := make([][]command.ProjectContext, len(projectCmds))
	for i, cmd := range projectCmds {
		groups[i] = []command.ProjectContext{cmd}
	}
	return groups
}

func createCancelledResults(remainingGroups [][]command.ProjectContext) []command.ProjectResult {
	var cancelledResults []command.ProjectResult
	for _, group := range remainingGroups {
		for _, cmd := range group {
			cancelledResults = append(cancelledResults, command.ProjectResult{
				Command: cmd.CommandName,
				ProjectCommandOutput: command.ProjectCommandOutput{
					Error: fmt.Errorf("operation cancelled via `atlantis cancel` command"),
				},
				RepoRelDir:  cmd.RepoRelDir,
				Workspace:   cmd.Workspace,
				ProjectName: cmd.ProjectName,
			})
		}
	}
	return cancelledResults
}
