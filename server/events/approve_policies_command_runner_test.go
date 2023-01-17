package events

import (
	"context"
	"fmt"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/db"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/lyft/feature"
	"github.com/stretchr/testify/assert"
	"testing"
)

const testUser = "user"

func buildCtx(t *testing.T) *command.Context {
	baseRepo := models.Repo{FullName: "testrepo"}
	return &command.Context{
		Pull:       models.PullRequest{BaseRepo: baseRepo},
		User:       models.User{},
		Log:        logging.NewNoopCtxLogger(t),
		RequestCtx: context.Background(),
	}
}

func TestApprovePoliciesCommandRunner_Success(t *testing.T) {
	testCtx := buildCtx(t)
	vcsStatusUpdater := &testVCSStatusUpdater{}
	policySet := valid.PolicySet{Name: testUser}
	policySets := valid.PolicySets{
		Owners: valid.PolicyOwners{
			Users: []string{testUser},
		},
		PolicySets: []valid.PolicySet{policySet},
	}
	builder := &testCmdBuilder{
		ctxs: []command.ProjectContext{
			{PolicySets: policySets},
		},
	}
	tmp := t.TempDir()
	defaultBoltDB, err := db.New(tmp)
	assert.NoError(t, err)
	dbUpdater := &DBUpdater{
		DB: defaultBoltDB,
	}
	commandRunner := ApprovePoliciesCommandRunner{
		allocator:                  &testAllocator{},
		vcsStatusUpdater:           vcsStatusUpdater,
		prjCmdBuilder:              builder,
		outputUpdater:              &testOutputUpdater{},
		prjCmdRunner:               &testCmdRunner{},
		policyCheckOutputGenerator: &testOutputGenerator{},
		dbUpdater:                  dbUpdater,
	}
	commandRunner.Run(testCtx, &command.Comment{})
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCalls, 1)
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCountCalls, 1)
	assert.Equal(t, vcsStatusUpdater.lastStatus, models.SuccessVCSStatus)
}

func TestApprovePoliciesCommandRunner_AllocationSkip(t *testing.T) {
	testCtx := buildCtx(t)
	allocator := &testAllocator{allocation: true}
	vcsStatusUpdater := &testVCSStatusUpdater{}
	commandRunner := ApprovePoliciesCommandRunner{
		allocator:        allocator,
		vcsStatusUpdater: vcsStatusUpdater,
	}
	commandRunner.Run(testCtx, &command.Comment{})
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCalls, 0)
}

func TestApprovePoliciesCommandRunner_BuilderError(t *testing.T) {
	testCtx := buildCtx(t)
	vcsStatusUpdater := &testVCSStatusUpdater{}
	prjCmdBuilder := &testCmdBuilder{
		error: assert.AnError,
	}
	commandRunner := ApprovePoliciesCommandRunner{
		allocator:        &testAllocator{},
		vcsStatusUpdater: vcsStatusUpdater,
		prjCmdBuilder:    prjCmdBuilder,
		outputUpdater:    &testOutputUpdater{},
	}
	commandRunner.Run(testCtx, &command.Comment{})
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCalls, 2)
	assert.Equal(t, vcsStatusUpdater.lastStatus, models.FailedVCSStatus)
}

func TestApprovePoliciesCommandRunner_NoProjects(t *testing.T) {
	testCtx := buildCtx(t)
	vcsStatusUpdater := &testVCSStatusUpdater{}
	commandRunner := ApprovePoliciesCommandRunner{
		allocator:        &testAllocator{},
		vcsStatusUpdater: vcsStatusUpdater,
		prjCmdBuilder:    &testCmdBuilder{},
		outputUpdater:    &testOutputUpdater{},
	}
	commandRunner.Run(testCtx, &command.Comment{})
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCalls, 1)
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCountCalls, 1)
	assert.Equal(t, vcsStatusUpdater.lastStatus, models.SuccessVCSStatus)
}

func TestApprovePoliciesCommandRunner_OutputStoreError(t *testing.T) {
	testCtx := buildCtx(t)
	vcsStatusUpdater := &testVCSStatusUpdater{}
	policySet := valid.PolicySet{Name: testUser}
	policySets := valid.PolicySets{
		Owners: valid.PolicyOwners{
			Users: []string{testUser},
		},
		PolicySets: []valid.PolicySet{policySet},
	}
	builder := &testCmdBuilder{
		ctxs: []command.ProjectContext{
			{PolicySets: policySets},
		},
	}
	commandRunner := ApprovePoliciesCommandRunner{
		allocator:                  &testAllocator{},
		vcsStatusUpdater:           vcsStatusUpdater,
		prjCmdBuilder:              builder,
		outputUpdater:              &testOutputUpdater{},
		prjCmdRunner:               &testCmdRunner{},
		policyCheckOutputGenerator: &testOutputGenerator{error: assert.AnError},
	}
	commandRunner.Run(testCtx, &command.Comment{})
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCalls, 2)
	assert.Equal(t, vcsStatusUpdater.numUpdateCombinedCountCalls, 0)
	assert.Equal(t, vcsStatusUpdater.lastStatus, models.FailedVCSStatus)
}

type testAllocator struct {
	allocation bool
	error      error
}

func (a *testAllocator) ShouldAllocate(_ feature.Name, _ feature.FeatureContext) (bool, error) {
	return a.allocation, a.error
}

type testVCSStatusUpdater struct {
	output                      string
	error                       error
	lastStatus                  models.VCSStatus
	numUpdateCombinedCalls      int
	numUpdateCombinedCountCalls int
}

func (v *testVCSStatusUpdater) UpdateCombined(_ context.Context, _ models.Repo, _ models.PullRequest, status models.VCSStatus, _ fmt.Stringer, _ string, _ string) (string, error) {
	v.lastStatus = status
	v.numUpdateCombinedCalls++
	return v.output, v.error
}

func (v *testVCSStatusUpdater) UpdateCombinedCount(_ context.Context, _ models.Repo, _ models.PullRequest, status models.VCSStatus, _ fmt.Stringer, _ int, _ int, _ string) (string, error) {
	v.lastStatus = status
	v.numUpdateCombinedCountCalls++
	return v.output, v.error
}

func (v *testVCSStatusUpdater) UpdateProject(context.Context, command.ProjectContext, fmt.Stringer, models.VCSStatus, string, string) (string, error) {
	return "", nil
}

type testCmdBuilder struct {
	ctxs  []command.ProjectContext
	error error
}

func (b *testCmdBuilder) BuildApprovePoliciesCommands(*command.Context, *command.Comment) ([]command.ProjectContext, error) {
	return b.ctxs, b.error
}

type testOutputUpdater struct {
}

func (u *testOutputUpdater) UpdateOutput(*command.Context, PullCommand, command.Result) {
}

type testCmdRunner struct {
	result command.ProjectResult
}

func (r *testCmdRunner) ApprovePolicies(command.ProjectContext) command.ProjectResult {
	return r.result
}

type testOutputGenerator struct {
	store command.PolicyCheckOutputStore
	error error
}

func (g *testOutputGenerator) GeneratePolicyCheckOutputStore(*command.Context, *command.Comment) (command.PolicyCheckOutputStore, error) {
	return g.store, g.error
}
