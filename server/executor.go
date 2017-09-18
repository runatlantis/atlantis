package server

type Executor interface {
	Execute(ctx *CommandContext)
}
