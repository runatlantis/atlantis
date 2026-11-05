// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
)

func (a *APIController) StateShow(w http.ResponseWriter, r *http.Request) {
	if !a.requireAPICommand(w, r, command.StateShow) {
		return
	}
	request, ctx, code, err := a.apiParseAndValidate(r)
	if err != nil {
		a.apiReportLegacyError(w, code, err)
		return
	}
	ctx.TeamAllowlistChecker = a.TeamAllowlistChecker
	if err := command.ValidateInspectionArgs(command.StateShow, request.Arguments); err != nil {
		a.apiReportLegacyError(w, http.StatusBadRequest, err)
		return
	}
	if len(request.Projects) == 0 && len(request.Paths) == 0 {
		a.apiReportLegacyError(w, http.StatusBadRequest, fmt.Errorf("specify at least one project or path"))
		return
	}
	if err := a.apiSetup(ctx, command.StateShow); err != nil {
		a.apiReportLegacyError(w, apiErrorStatusCode(err), err)
		return
	}
	defer a.cleanupNonPRWorkingDir(ctx)
	build := func(ctx *command.Context, cc *events.CommentCommand) ([]command.ProjectContext, error) {
		cc.Flags = slices.Clone(request.Arguments)
		return a.ProjectCommandBuilder.BuildStateShowCommands(ctx, cc)
	}
	cmds, ccs, err := request.getCommands(ctx, command.StateShow, build)
	if events.IsIgnoredTargetedDir(err) {
		ctx.CommandSkipped = true
		a.getAPIMiddleware().Responder.writeJSON(w, http.StatusOK, command.Result{ProjectResults: []command.ProjectResult{}})
		return
	}
	if err != nil {
		a.apiReportLegacyError(w, apiErrorStatusCode(err), err)
		return
	}
	result := command.Result{ProjectResults: []command.ProjectResult{}}
	for i, cmd := range cmds {
		if err := a.runPreHooks(ctx, ccs[i]); err != nil {
			a.apiReportLegacyError(w, http.StatusInternalServerError, err)
			return
		}
		result.ProjectResults = append(result.ProjectResults, events.RunOneProjectCmd(a.ProjectStateShowCommandRunner.StateShow, cmd))
		a.runPostHooks(ctx, ccs[i:i+1])
	}
	status := http.StatusOK
	if result.HasErrors() {
		status = http.StatusInternalServerError
	}
	a.getAPIMiddleware().Responder.writeJSON(w, status, result)
}
