package events

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_executor.go Executor

type Executor interface {
	Execute(ctx *CommandContext) CommandResponse
}
