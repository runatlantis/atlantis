package server

import (
	"path/filepath"
	"github.com/hootsuite/atlantis/locking"
)

type BaseExecutor struct {
	github                *GithubClient
	awsConfig             *AWSConfig
	scratchDir            string
	s3Bucket              string
	sshKey                string
	ghComments            *GithubCommentRenderer
	terraform             *TerraformClient
	githubCommentRenderer *GithubCommentRenderer
	lockingBackend        locking.Backend
}

type PullRequestContext struct {
	repoFullName          string
	head                  string
	base                  string
	number                int
	pullRequestLink       string
	terraformApplier      string
	terraformApplierEmail string
}

type ExecutionResult struct {
	SetupError   Templater
	SetupFailure Templater
	PathResults  []PathResult
	Command      CommandType
}

type PathResult struct {
	Path   string
	Status string // todo: this should be an enum for success/error/failure
	Result Templater
}

type ExecutionPath struct {
	// Absolute is the full path on the OS where we will execute.
	// Will never end with a '/'.
	Absolute string
	// Relative is the path relative to the repo root.
	// Will never end with a '/'.
	Relative string
}

func NewExecutionPath(absolutePath string, relativePath string) ExecutionPath {
	return ExecutionPath{filepath.Clean(absolutePath), filepath.Clean(relativePath)}
}

func (b *BaseExecutor) updateGithubStatus(pullCtx *PullRequestContext, pathResults []PathResult) {
	// the status will be the worst result
	worstResult := b.worstResult(pathResults)
	if worstResult == "success" {
		b.github.UpdateStatus(pullCtx, SuccessStatus, "Plan Succeeded")
	} else if worstResult == "failure" {
		b.github.UpdateStatus(pullCtx, FailureStatus, "Plan Failed")
	} else {
		b.github.UpdateStatus(pullCtx, ErrorStatus, "Plan Error")
	}
}

func (b *BaseExecutor) worstResult(results []PathResult) string {
	var worst string = "success"
	for _, result := range results {
		if result.Status == "error" {
			return result.Status
		} else if result.Status == "failure" {
			worst = result.Status
		}
	}
	return worst
}

func (b *BaseExecutor) Exec(f func(*ExecutionContext, *PullRequestContext) ExecutionResult, ctx *ExecutionContext, github *GithubClient) {
	pullCtx := b.githubContext(ctx)
	result := f(ctx, pullCtx)
	comment := b.githubCommentRenderer.render(result, ctx.log.History.String(), ctx.command.verbose)
	github.CreateComment(pullCtx, comment)
}

func (b *BaseExecutor) githubContext(ctx *ExecutionContext) *PullRequestContext {
	return &PullRequestContext{
		repoFullName:          ctx.repoFullName,
		head:                  ctx.head,
		base:                  ctx.base,
		number:                ctx.pullNum,
		pullRequestLink:       ctx.pullLink,
		terraformApplier:      ctx.requesterUsername,
		terraformApplierEmail: ctx.requesterEmail,
	}
}

type Templater interface {
	Template() *CompiledTemplate
}

type GeneralError struct {
	Error error
}

func (g GeneralError) Template() *CompiledTemplate {
	return GeneralErrorTmpl
}
