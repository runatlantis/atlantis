package server

type Executor interface {
	execute(ctx *ExecutionContext, pullCtx *PullRequestContext) ExecutionResult
}
