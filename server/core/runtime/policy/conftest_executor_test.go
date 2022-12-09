package policy_test

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/policy"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	contextInternal "github.com/runatlantis/atlantis/server/neptune/context"
	"github.com/stretchr/testify/assert"
	"github.com/uber-go/tally/v4"
	"strings"
	"testing"
)

const (
	path           = "/path/to/some/place"
	output         = "test output"
	workDir        = "workDir"
	executablePath = "executablepath"
	policyA        = "A"
	policyB        = "B"
)

func buildTestProjectCtx(t *testing.T, policySets []valid.PolicySet) command.ProjectContext {
	ctx := context.WithValue(context.Background(), contextInternal.InstallationIDKey, int64(1))
	return command.ProjectContext{
		PolicySets: valid.PolicySets{
			Version:    nil,
			Owners:     valid.PolicyOwners{},
			PolicySets: policySets,
		},
		Log:               logging.NewNoopCtxLogger(t),
		Scope:             tally.NewTestScope("test", map[string]string{}),
		RequestCtx:        ctx,
		InstallationToken: 1,
	}
}

func buildTestTitle(policySets []valid.PolicySet) string {
	var names []string
	for _, policy := range policySets {
		names = append(names, policy.Name)
	}
	return fmt.Sprintf("Checking plan against the following policies: \n  %s\n", strings.Join(names, "\n  "))
}

func TestConfTestExecutor_PolicySuccess(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
	}
	policyFilter := &mockPolicyFilter{}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	policySets := []valid.PolicySet{
		{Name: policyA},
		{Name: policyB},
	}
	prjCtx := buildTestProjectCtx(t, policySets)
	expectedTitle := buildTestTitle(policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.NoError(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.True(t, exec.isCalled)
	assert.True(t, policyFilter.isCalled)
	assert.Contains(t, cmdOutput, expectedTitle)
	assert.Contains(t, cmdOutput, output)
}

func TestConfTestExecutor_PolicySuccess_FilteredFailures(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
		error:  assert.AnError,
	}
	policyFilter := &mockPolicyFilter{}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	policySets := []valid.PolicySet{
		{Name: policyA},
		{Name: policyB},
	}
	prjCtx := buildTestProjectCtx(t, policySets)
	expectedTitle := buildTestTitle(policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.NoError(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.True(t, exec.isCalled)
	assert.True(t, policyFilter.isCalled)
	assert.Contains(t, cmdOutput, expectedTitle)
	assert.Contains(t, cmdOutput, output)
}

func TestConfTestExecutor_PolicyFailure_NotFiltered(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
		error:  assert.AnError,
	}
	policySets := []valid.PolicySet{
		{Name: policyA},
		{Name: policyB},
	}
	policyFilter := &mockPolicyFilter{
		policies: policySets,
	}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	prjCtx := buildTestProjectCtx(t, policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	expectedTitle := buildTestTitle(policySets)
	assert.Error(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.True(t, exec.isCalled)
	assert.True(t, policyFilter.isCalled)
	assert.Contains(t, cmdOutput, expectedTitle)
	assert.Contains(t, cmdOutput, output)
}

func TestConfTestExecutor_BuildArgError(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
	}
	policyFilter := &mockPolicyFilter{}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	var policySets []valid.PolicySet
	prjCtx := buildTestProjectCtx(t, policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.Error(t, err)
	assert.False(t, sourceResolver.isCalled)
	assert.False(t, exec.isCalled)
	assert.False(t, policyFilter.isCalled)
	assert.Empty(t, cmdOutput)
}

func TestConfTestExecutor_SourceResolverError(t *testing.T) {
	sourceResolver := &mockSourceResolver{error: assert.AnError}
	exec := &mockExec{
		output: output,
	}
	policyFilter := &mockPolicyFilter{}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	policySets := []valid.PolicySet{
		{Name: policyA},
	}
	prjCtx := buildTestProjectCtx(t, policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.Error(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.False(t, exec.isCalled)
	assert.False(t, policyFilter.isCalled)
	assert.Empty(t, cmdOutput)
}

func TestConfTestExecutor_FilterFailure(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
	}
	policySets := []valid.PolicySet{
		{Name: policyA},
		{Name: policyB},
	}
	policyFilter := &mockPolicyFilter{error: assert.AnError}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	prjCtx := buildTestProjectCtx(t, policySets)
	expectedTitle := buildTestTitle(policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.Error(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.True(t, exec.isCalled)
	assert.True(t, policyFilter.isCalled)
	assert.Contains(t, cmdOutput, expectedTitle)
	assert.Contains(t, cmdOutput, output)
}

func TestConfTestExecutor_MissingInstallationToken(t *testing.T) {
	sourceResolver := &mockSourceResolver{path: path}
	exec := &mockExec{
		output: output,
	}
	policyFilter := &mockPolicyFilter{}
	executor := policy.ConfTestExecutor{
		SourceResolver: sourceResolver,
		Exec:           exec,
		PolicyFilter:   policyFilter,
	}
	var args []string
	policySets := []valid.PolicySet{
		{Name: policyA},
		{Name: policyB},
	}
	prjCtx := command.ProjectContext{
		PolicySets: valid.PolicySets{
			Version:    nil,
			Owners:     valid.PolicyOwners{},
			PolicySets: policySets,
		},
		Log:        logging.NewNoopCtxLogger(t),
		Scope:      tally.NewTestScope("test", map[string]string{}),
		RequestCtx: context.Background(),
	}
	expectedTitle := buildTestTitle(policySets)
	cmdOutput, err := executor.Run(context.Background(), prjCtx, executablePath, map[string]string{}, workDir, args)
	assert.Error(t, err)
	assert.True(t, sourceResolver.isCalled)
	assert.True(t, exec.isCalled)
	assert.False(t, policyFilter.isCalled)
	assert.Contains(t, cmdOutput, expectedTitle)
	assert.Contains(t, cmdOutput, output)
}

type mockSourceResolver struct {
	isCalled bool
	path     string
	error    error
}

func (r *mockSourceResolver) Resolve(_ valid.PolicySet) (string, error) {
	r.isCalled = true
	return r.path, r.error
}

type mockPolicyFilter struct {
	isCalled bool
	policies []valid.PolicySet
	error    error
}

func (r *mockPolicyFilter) Filter(_ context.Context, _ int64, _ models.Repo, _ int, _ []valid.PolicySet) ([]valid.PolicySet, error) {
	r.isCalled = true
	return r.policies, r.error
}

type mockExec struct {
	isCalled bool
	output   string
	error    error
}

func (r *mockExec) CombinedOutput(_ []string, _ map[string]string, _ string) (string, error) {
	r.isCalled = true
	return r.output, r.error
}
