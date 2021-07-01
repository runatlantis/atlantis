package runtime

import (
	version "github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_versionedexecutorworkflow.go VersionedExecutorWorkflow

// VersionedExecutorWorkflow defines a versioned execution for a given project context
type VersionedExecutorWorkflow interface {
	ExecutorVersionEnsurer
	Executor
}

// Executor runs an executable with provided environment variables and arguments and returns stdout
type Executor interface {
	Run(ctx models.ProjectCommandContext, executablePath string, envs map[string]string, workdir string, extraArgs []string) (string, error)
}

// ExecutorVersionEnsurer ensures a given version exists and outputs a path to the executable
type ExecutorVersionEnsurer interface {
	EnsureExecutorVersion(log logging.SimpleLogging, v *version.Version) (string, error)
}
