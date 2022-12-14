package events

import "github.com/google/uuid"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_uuid_generator.go UUIDGenerator

type UUIDGenerator interface {
	GenerateUUID() string
}

type DefaultPreWorkflowHookUUIDGenerator struct{}

func (g DefaultPreWorkflowHookUUIDGenerator) GenerateUUID() string {
	return uuid.NewString()
}
