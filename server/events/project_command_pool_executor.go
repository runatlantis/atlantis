package events

import (
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

// assignEffectiveExecutionGroups assigns the EffectiveExecutionOrderGroup for each project.
// For non-destroy commands, EffectiveExecutionOrderGroup remains 0.
// For destroy commands, it uses DestroyExecutionOrderGroup if defined, otherwise ExecutionOrderGroup.
func assignEffectiveExecutionGroups(cmds []command.ProjectContext, isDestroy bool) {
	for i := range cmds {
		if isDestroy && cmds[i].DestroyExecutionOrderGroup != nil {
			cmds[i].EffectiveExecutionOrderGroup = *cmds[i].DestroyExecutionOrderGroup
		} else {
			cmds[i].EffectiveExecutionOrderGroup = cmds[i].ExecutionOrderGroup
		}
	}
}

func splitByExecutionOrderGroup(cmds []command.ProjectContext) [][]command.ProjectContext {
	groups := make(map[int][]command.ProjectContext)
	for _, cmd := range cmds {
		// After assignEffectiveExecutionGroups has been called, EffectiveExecutionOrderGroup
		// will contain the correct group (either from DestroyExecutionOrderGroup for destroy
		// commands, or from ExecutionOrderGroup for normal commands)
		groupKey := cmd.EffectiveExecutionOrderGroup
		groups[groupKey] = append(groups[groupKey], cmd)
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
	isDestroy := false
	for _, c := range cmds {
		if c.IsDestroy {
			isDestroy = true
			break
		}
	}
	assignEffectiveExecutionGroups(cmds, isDestroy)

	groups := splitByExecutionOrderGroup(cmds)
	for _, group := range groups {
		res := runProjectCmdsParallel(group, runnerFunc, poolSize)
		results = append(results, res.ProjectResults...)
		if res.HasErrors() && group[0].AbortOnExecutionOrderFail {
			break
		}
	}

	return command.Result{ProjectResults: results}
}
