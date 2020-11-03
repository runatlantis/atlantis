package models

import (
	"os/exec"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_exec.go Exec

type Exec interface {
	LookPath(file string) (string, error)
}

type LocalExec struct{}

func (e LocalExec) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}
