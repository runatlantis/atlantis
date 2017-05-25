package main

type Executor interface {
	execute(ctx *ExecutionContext, prCtx *PullRequestContext) ExecutionResult
}
