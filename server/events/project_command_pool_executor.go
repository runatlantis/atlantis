package events

import (
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/models"
)

type prjCmdRunnerFunc func(ctx models.ProjectCommandContext, parallel models.ParallelCommand) models.ProjectResult

func runProjectCmdsParallel(
	cmds []models.ProjectCommandContext,
	runnerFunc prjCmdRunnerFunc,
	poolSize int,
) CommandResult {
	var results []models.ProjectResult
	mux := &sync.Mutex{}

	wg := sizedwaitgroup.New(poolSize)
	for idx, pCmd := range cmds {
		pCmd := pCmd
		var execute func()
		wg.Add()

		execute = func() {
			defer wg.Done()
			res := runnerFunc(pCmd, models.ParallelCommand{idx, true})
			mux.Lock()
			results = append(results, res)
			mux.Unlock()
		}

		go execute()
	}

	wg.Wait()
	return CommandResult{ProjectResults: results}
}

func runProjectCmds(
	cmds []models.ProjectCommandContext,
	runnerFunc prjCmdRunnerFunc,
) CommandResult {
	var results []models.ProjectResult
	for _, pCmd := range cmds {
		res := runnerFunc(pCmd, models.ParallelCommand{})

		results = append(results, res)
	}
	return CommandResult{ProjectResults: results}
}
