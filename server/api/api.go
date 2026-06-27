package api

import (
	"github.com/runatlantis/atlantis/server/api/models"
	"github.com/runatlantis/atlantis/server/events/command"
)

//go:generate oapi-codegen -generate types,skip-prune -package models -o models/models.gen.go ../../openapi/atlantis.yaml

func BuildPlanResult(result *command.Result) models.PlanResponse {
	if result == nil {
		noResponse := "Unexpected no response"
		return models.PlanResponse{
			Error: &noResponse,
		}
	}
	ret := models.PlanResponse{
		ProjectResults: make([]models.PlanProjectResponse, len(result.ProjectResults)),
	}
	if result.Error != nil {
		errMsg := result.Error.Error()
		ret.Error = &errMsg
	}
	for i, projectResult := range result.ProjectResults {
		entry := models.PlanProjectResponse{}
		if projectResult.Error != nil {
			errMsg := projectResult.Error.Error()
			entry.Error = &errMsg
		}
		if projectResult.PlanSuccess == nil {
			entry.Success = false
		} else {
			entry.ProjectName = projectResult.ProjectName
			entry.RepoRelDir = projectResult.RepoRelDir
			entry.TerraformOutput = &projectResult.PlanSuccess.TerraformOutput
			entry.Workspace = projectResult.Workspace
		}
		ret.ProjectResults[i] = entry
	}
	return ret
}
