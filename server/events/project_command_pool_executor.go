package events

import (
	"fmt"
	"sort"
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/command"
)

type prjCmdRunnerFunc func(ctx command.ProjectContext) command.ProjectResult

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
			res := runnerFunc(pCmd)
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
		res := runnerFunc(pCmd)

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

func runProjectCmdsWithCancellationTracker(
	ctx *command.Context,
	projectCmds []command.ProjectContext,
	cancellationTracker CancellationTracker,
	parallelPoolSize int,
	isParallel bool,
	runnerFunc func(command.ProjectContext) command.ProjectResult,
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

		groupResult := runGroup(group, runnerFunc, isParallel, parallelPoolSize)
		results = append(results, groupResult.ProjectResults...)

		if groupResult.HasErrors() && group[0].AbortOnExecutionOrderFail && isParallel {
			ctx.Log.Info("abort on execution order when failed")
			break
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
				Command:     cmd.CommandName,
				Error:       fmt.Errorf("operation cancelled"),
				RepoRelDir:  cmd.RepoRelDir,
				Workspace:   cmd.Workspace,
				ProjectName: cmd.ProjectName,
			})
		}
	}
	return cancelledResults
}

func runGroup(
	group []command.ProjectContext,
	runnerFunc func(command.ProjectContext) command.ProjectResult,
	isParallel bool,
	parallelPoolSize int,
) command.Result {
	if isParallel && len(group) > 1 {
		return runProjectCmdsParallel(group, runnerFunc, parallelPoolSize)
	}
	return runProjectCmds(group, runnerFunc)
}
