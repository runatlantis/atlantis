package main

import (
	"encoding/json"
	"fmt"

	"github.com/hootsuite/atlantis/logging"
)

type StashPullRequestContext struct {
	owner                 string
	repoName              string
	number                int
	pullRequestLink       string
	terraformApplier      string
	terraformApplierEmail string
}

type StashPRClient struct {
	client *StashClient
}

func (s *StashPRClient) LockState(log *logging.SimpleLogger, ctx *StashPullRequestContext, path string) StashLockResponse {
	state := map[string]string{
		"pull_request_id":         fmt.Sprintf("%d", ctx.number),
		"pull_request_link":       ctx.pullRequestLink,
		"owner_name":              ctx.owner,
		"repo_name":               ctx.repoName,
		"terraform_applier":       ctx.terraformApplier,
		"terraform_applier_email": ctx.terraformApplierEmail,
	}
	stateJson, _ := json.Marshal(state) // todo: swallowing error here since don't want to construct StashLockResponse, but I think that StashLockResponse should not encapsulate the idea of an error
	return s.client.LockState(log, path, stateJson)
}

func (s *StashPRClient) UnlockState(log *logging.SimpleLogger, ctx *StashPullRequestContext, path string) StashUnlockResponse {
	state := map[string]string{
		"pull_request_link":       ctx.pullRequestLink,
		"terraform_applier":       ctx.terraformApplier,
		"repo_name":               ctx.repoName,
		"terraform_applier_email": ctx.terraformApplierEmail,
	}
	stateJson, _ := json.Marshal(state) // todo: swallowing error here since don't want to construct StashLockResponse, but I think that StashLockResponse should not encapsulate the idea of an error
	return s.client.UnlockState(log, path, stateJson)
}
