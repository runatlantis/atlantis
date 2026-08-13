// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"fmt"
	"sort"
	"sync"

	"github.com/remeh/sizedwaitgroup"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
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
	return runApplyCmdsWithOptionalReplan(ctx, projectCmds, cancellationTracker, parallelPoolSize, isParallel, projectCmdRunners{
		apply: runnerFunc,
	})
}

// projectCmdRunners holds the runners used during apply. plan/policyCheck are
// optional and only used when replan_between_execution_order_groups is enabled.
type projectCmdRunners struct {
	apply       prjCmdRunnerFunc
	plan        prjCmdRunnerFunc
	policyCheck prjCmdRunnerFunc
}

func runApplyCmdsWithOptionalReplan(
	ctx *command.Context,
	projectCmds []command.ProjectContext,
	cancellationTracker CancellationTracker,
	parallelPoolSize int,
	isParallel bool,
	runners projectCmdRunners,
) command.Result {
	if isParallel {
		ctx.Log.Info("Running commands in parallel")
	}

	multiEOGroups := len(splitByExecutionOrderGroup(projectCmds)) > 1
	groups := prepareExecutionGroups(projectCmds, isParallel)
	if cancellationTracker != nil {
		defer cancellationTracker.Clear(ctx.Pull)
	}

	replanEnabled := runners.plan != nil && anyReplanEnabled(projectCmds)

	var results []command.ProjectResult
	for i, group := range groups {
		if i > 0 && cancellationTracker != nil && cancellationTracker.IsCancelled(ctx.Pull) {
			ctx.Log.Info("Skipping execution order group %d and all subsequent groups due to cancellation", group[0].ExecutionOrderGroup)
			results = append(results, createCancelledResults(groups[i:])...)
			break
		}

		if i > 0 && replanEnabled {
			refreshedGroup, failedIdx, abort := replanGroupBeforeApply(ctx, group, multiEOGroups, runners)
			if abort {
				// Do not apply this group or later ones after a hard replan abort.
				results = append(results, applyErrorsForProjectContexts(refreshedGroup,
					"apply aborted after mid-apply replan failure for a prior project in the execution order")...)
				break
			}
			results = append(results, applyErrorsForIndexes(refreshedGroup, failedIdx,
				"skipping apply after mid-apply replan failure; fix the plan error and re-run apply")...)
			group = filterFailedReplans(refreshedGroup, failedIdx)
			if len(group) == 0 {
				continue
			}
		}

		var groupResult command.Result
		if isParallel && len(group) > 1 {
			groupResult = runProjectCmdsParallel(group, runners.apply, parallelPoolSize, cancellationTracker, ctx.Pull)
		} else {
			groupResult = runProjectCmds(group, runners.apply)
		}
		results = append(results, groupResult.ProjectResults...)
		updatePullStatusFromResults(ctx, groupResult.ProjectResults)

		if !groupResult.HasErrors() {
			continue
		}

		if replanEnabled {
			// Later groups must not refresh/apply against a partially failed earlier group.
			ctx.Log.Info("aborting later execution order groups after apply failure while replan_between_execution_order_groups is enabled")
			if i+1 < len(groups) {
				results = append(results, applyErrorsForProjectContexts(flattenProjectGroups(groups[i+1:]),
					"apply skipped because an earlier execution order group failed while replan_between_execution_order_groups was enabled")...)
			}
			break
		}

		// Historical abort_on_execution_order_fail behavior: stop running later
		// groups, but do not synthesize results for projects that never ran.
		if group[0].AbortOnExecutionOrderFail && isParallel {
			ctx.Log.Info("abort on execution order when failed")
			break
		}
	}

	return command.Result{ProjectResults: results}
}

func anyReplanEnabled(cmds []command.ProjectContext) bool {
	for _, cmd := range cmds {
		if cmd.ReplanBetweenExecutionOrderGroups {
			return true
		}
	}
	return false
}

func applyErrorsForIndexes(group []command.ProjectContext, indexes map[int]struct{}, msg string) []command.ProjectResult {
	if len(indexes) == 0 {
		return nil
	}
	var out []command.ProjectResult
	for i, cmd := range group {
		if _, ok := indexes[i]; !ok {
			continue
		}
		out = append(out, applyErrorResult(cmd, msg))
	}
	return out
}

func applyErrorsForProjectContexts(cmds []command.ProjectContext, msg string) []command.ProjectResult {
	out := make([]command.ProjectResult, 0, len(cmds))
	for _, cmd := range cmds {
		out = append(out, applyErrorResult(cmd, msg))
	}
	return out
}

func applyErrorResult(cmd command.ProjectContext, msg string) command.ProjectResult {
	return command.ProjectResult{
		Command: command.Apply,
		ProjectCommandOutput: command.ProjectCommandOutput{
			Error: fmt.Errorf("%s", msg),
		},
		RepoRelDir:  cmd.RepoRelDir,
		Workspace:   cmd.Workspace,
		ProjectName: cmd.ProjectName,
	}
}

func flattenProjectGroups(groups [][]command.ProjectContext) []command.ProjectContext {
	var out []command.ProjectContext
	for _, group := range groups {
		out = append(out, group...)
	}
	return out
}

// replanGroupBeforeApply refreshes planfiles for projects that need it.
// Plan/policy outputs are intentionally not returned as ProjectResults so they
// do not pollute apply PR comments or the apply DB status update.
func replanGroupBeforeApply(
	ctx *command.Context,
	group []command.ProjectContext,
	multiEOGroups bool,
	runners projectCmdRunners,
) (refreshed []command.ProjectContext, failedIdx map[int]struct{}, abort bool) {
	refreshed = make([]command.ProjectContext, len(group))
	copy(refreshed, group)
	failedIdx = make(map[int]struct{})

	for j, cmd := range group {
		if !shouldReplanBeforeApply(cmd, multiEOGroups) {
			continue
		}
		if len(cmd.PlanSteps) == 0 {
			ctx.Log.Warn("skipping mid-apply replan for %s/%s: no plan steps available", cmd.RepoRelDir, cmd.ProjectName)
			continue
		}

		ctx.Log.Info("Replanning %s (project %q) before apply because earlier execution order groups were applied",
			cmd.RepoRelDir, cmd.ProjectName)

		planCtx := cmd
		planCtx.CommandName = command.Plan
		planCtx.Steps = cmd.PlanSteps
		planCtx.ExpectedPlanHash = ""
		planCtx.MidApplyReplan = true
		// Mid-apply refresh must not flip PR commit statuses to "plan".
		planCtx.SuppressVCSStatus = true
		planCtx.SuppressJobOutput = true
		planOut := runners.plan(planCtx)
		updatePullStatusFromResults(ctx, []command.ProjectResult{{
			ProjectCommandOutput: planOut,
			Command:              command.Plan,
			RepoRelDir:           cmd.RepoRelDir,
			Workspace:            cmd.Workspace,
			ProjectName:          cmd.ProjectName,
		}})

		if planOut.Error != nil || planOut.Failure != "" {
			ctx.Log.Err("mid-apply replan failed for %s (project %q): %v %s",
				cmd.RepoRelDir, cmd.ProjectName, planOut.Error, planOut.Failure)
			failedIdx[j] = struct{}{}
			if cmd.AbortOnExecutionOrderFail {
				abort = true
			}
			continue
		}

		if len(cmd.PolicyCheckSteps) > 0 && runners.policyCheck != nil {
			policyCtx := cmd
			policyCtx.CommandName = command.PolicyCheck
			policyCtx.Steps = cmd.PolicyCheckSteps
			policyCtx.SuppressVCSStatus = true
			policyCtx.SuppressJobOutput = true
			policyOut := runners.policyCheck(policyCtx)
			updatePullStatusFromResults(ctx, []command.ProjectResult{{
				ProjectCommandOutput: policyOut,
				Command:              command.PolicyCheck,
				RepoRelDir:           cmd.RepoRelDir,
				Workspace:            cmd.Workspace,
				ProjectName:          cmd.ProjectName,
			}})
			if policyOut.Error != nil || policyOut.Failure != "" {
				ctx.Log.Err("mid-apply policy check failed for %s (project %q): %v %s",
					cmd.RepoRelDir, cmd.ProjectName, policyOut.Error, policyOut.Failure)
				failedIdx[j] = struct{}{}
				if cmd.AbortOnExecutionOrderFail {
					abort = true
				}
				continue
			}
		}

		// Clear the hash captured from the pre-apply planfile so apply
		// re-hashes the refreshed plan produced above.
		refreshed[j].ExpectedPlanHash = ""
	}

	return refreshed, failedIdx, abort
}

func filterFailedReplans(group []command.ProjectContext, failedIdx map[int]struct{}) []command.ProjectContext {
	if len(failedIdx) == 0 {
		return group
	}
	out := make([]command.ProjectContext, 0, len(group)-len(failedIdx))
	for i, cmd := range group {
		if _, failed := failedIdx[i]; failed {
			continue
		}
		out = append(out, cmd)
	}
	return out
}

func shouldReplanBeforeApply(cmd command.ProjectContext, multiEOGroups bool) bool {
	if !cmd.ReplanBetweenExecutionOrderGroups {
		return false
	}
	if multiEOGroups {
		return true
	}
	// Same execution_order_group applied sequentially: only refresh projects
	// that declare depends_on (dependency graphs without distinct EO groups).
	return len(cmd.DependsOn) > 0
}

func updatePullStatusFromResults(ctx *command.Context, results []command.ProjectResult) {
	if ctx.PullStatus == nil {
		return
	}
	for _, result := range results {
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
