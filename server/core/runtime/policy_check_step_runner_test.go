package runtime_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	expectedOutput = "output"
	executablePath = "some/path/conftest"
)

func buildTestPrjCtx(t *testing.T) command.ProjectContext {
	v, err := version.NewVersion("1.0")
	assert.NoError(t, err)
	return command.ProjectContext{
		Log: logging.NewNoopCtxLogger(t),
		BaseRepo: models.Repo{
			FullName: "owner/repo",
		},
		PolicySets: valid.PolicySets{
			Version:    v,
			PolicySets: []valid.PolicySet{},
		},
	}
}

func TestRun_Successful(t *testing.T) {
	prjCtx := buildTestPrjCtx(t)
	ensurer := &mockEnsurer{}
	executor := &mockExecutor{
		output: expectedOutput,
	}
	legacyExecutor := &mockLegacyExecutor{}
	allocator := &mockFeatureAllocator{
		enabled: true,
	}
	runner := &runtime.PolicyCheckStepRunner{
		VersionEnsurer: ensurer,
		Executor:       executor,
		LegacyExecutor: legacyExecutor,
		Allocator:      allocator,
	}
	output, err := runner.Run(context.Background(), prjCtx, []string{}, executablePath, map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, output, expectedOutput)
	assert.True(t, ensurer.isCalled)
	assert.True(t, executor.isCalled)
	assert.True(t, allocator.isCalled)
	assert.False(t, legacyExecutor.isCalled)
}

func TestRun_LegacySuccess(t *testing.T) {
	prjCtx := buildTestPrjCtx(t)
	ensurer := &mockEnsurer{}
	executor := &mockExecutor{}
	legacyExecutor := &mockLegacyExecutor{
		output: expectedOutput,
	}
	allocator := &mockFeatureAllocator{}
	runner := &runtime.PolicyCheckStepRunner{
		VersionEnsurer: ensurer,
		Executor:       executor,
		LegacyExecutor: legacyExecutor,
		Allocator:      allocator,
	}
	output, err := runner.Run(context.Background(), prjCtx, []string{}, executablePath, map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, output, expectedOutput)
	assert.True(t, ensurer.isCalled)
	assert.False(t, executor.isCalled)
	assert.True(t, allocator.isCalled)
	assert.True(t, legacyExecutor.isCalled)
}

func TestRun_LegacySuccess_AllocationError(t *testing.T) {
	prjCtx := buildTestPrjCtx(t)
	ensurer := &mockEnsurer{}
	executor := &mockExecutor{}
	legacyExecutor := &mockLegacyExecutor{
		output: expectedOutput,
	}
	allocator := &mockFeatureAllocator{
		err: assert.AnError,
	}
	runner := &runtime.PolicyCheckStepRunner{
		VersionEnsurer: ensurer,
		Executor:       executor,
		LegacyExecutor: legacyExecutor,
		Allocator:      allocator,
	}
	output, err := runner.Run(context.Background(), prjCtx, []string{}, executablePath, map[string]string{})
	assert.NoError(t, err)
	assert.Equal(t, output, expectedOutput)
	assert.True(t, ensurer.isCalled)
	assert.False(t, executor.isCalled)
	assert.True(t, allocator.isCalled)
	assert.True(t, legacyExecutor.isCalled)
}

func TestRun_LegacyFailure(t *testing.T) {
	prjCtx := buildTestPrjCtx(t)
	ensurer := &mockEnsurer{}
	executor := &mockExecutor{}
	legacyExecutor := &mockLegacyExecutor{
		output: expectedOutput,
		err:    assert.AnError,
	}
	allocator := &mockFeatureAllocator{}
	runner := &runtime.PolicyCheckStepRunner{
		VersionEnsurer: ensurer,
		Executor:       executor,
		LegacyExecutor: legacyExecutor,
		Allocator:      allocator,
	}
	output, err := runner.Run(context.Background(), prjCtx, []string{}, executablePath, map[string]string{})
	assert.Error(t, err)
	assert.Equal(t, output, expectedOutput)
	assert.True(t, ensurer.isCalled)
	assert.False(t, executor.isCalled)
	assert.True(t, allocator.isCalled)
	assert.True(t, legacyExecutor.isCalled)
}

func TestRun_EnsurerFailure(t *testing.T) {
	prjCtx := buildTestPrjCtx(t)
	ensurer := &mockEnsurer{
		err: assert.AnError,
	}
	executor := &mockExecutor{}
	legacyExecutor := &mockLegacyExecutor{}
	allocator := &mockFeatureAllocator{}
	runner := &runtime.PolicyCheckStepRunner{
		VersionEnsurer: ensurer,
		Executor:       executor,
		LegacyExecutor: legacyExecutor,
		Allocator:      allocator,
	}
	output, err := runner.Run(context.Background(), prjCtx, []string{}, executablePath, map[string]string{})
	assert.Error(t, err)
	assert.Empty(t, output)
	assert.True(t, ensurer.isCalled)
	assert.False(t, executor.isCalled)
	assert.False(t, allocator.isCalled)
	assert.False(t, legacyExecutor.isCalled)
}

type mockFeatureAllocator struct {
	enabled  bool
	err      error
	isCalled bool
}

func (t *mockFeatureAllocator) ShouldAllocate(_ feature.Name, _ feature.FeatureContext) (bool, error) {
	t.isCalled = true
	return t.enabled, t.err
}

type mockLegacyExecutor struct {
	output   string
	err      error
	isCalled bool
}

func (t *mockLegacyExecutor) Run(_ context.Context, _ command.ProjectContext, _ string, _ map[string]string, _ string, _ []string) (string, error) {
	t.isCalled = true
	return t.output, t.err
}

type mockExecutor struct {
	output   string
	err      error
	isCalled bool
}

func (t *mockExecutor) Run(_ context.Context, _ command.ProjectContext, _ string, _ map[string]string, _ string, _ []string) (string, error) {
	t.isCalled = true
	return t.output, t.err
}

type mockEnsurer struct {
	output   string
	err      error
	isCalled bool
}

func (t *mockEnsurer) EnsureExecutorVersion(_ logging.Logger, _ *version.Version) (string, error) {
	t.isCalled = true
	return t.output, t.err
}
