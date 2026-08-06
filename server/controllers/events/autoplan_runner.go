package events

import "github.com/runatlantis/atlantis/server/events/command"

func (e *VCSEventsController) runAutoplan(request autoplanRequest) {
	defer e.AutoplanRuns.release(request)
	e.CommandRunner.RunAutoplanCommand(request.baseRepo, request.headRepo, request.pull, request.user)
	for {
		pendingEvents := e.AutoplanRuns.pendingEvents(request)
		if pendingEvents == 0 {
			return
		}
		livePull, err := e.LivePullHeadFetcher.GetLivePullIdentity(command.ProjectContext{Log: e.Logger, Pull: request.pull})
		if err != nil {
			e.Logger.Err("fetching live pull request for autoplan: %s", err)
			return
		}
		next, started, retry := e.AutoplanRuns.scheduleNext(request, pendingEvents, livePull)
		if started {
			go e.runAutoplan(next)
		}
		if !retry {
			return
		}
	}
}
