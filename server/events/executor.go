package events

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_executor.go Executor

// Executor is the generic interface implemented by each command type:
// help, plan, and apply.
type Executor interface {
	Execute(ctx *CommandContext) CommandResponse
}
