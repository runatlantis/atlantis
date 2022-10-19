package policies_test

import (
	"context"
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/command/policies"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestPolicyCheckOutputGenerator(t *testing.T) {
	t.Run("error build plan command fails", func(t *testing.T) {
		ctx := command.Context{
			RequestCtx: context.Background(),
			Log:        logging.NewNoopCtxLogger(t),
		}

		comment := command.Comment{
			Name: command.ApprovePolicies,
		}

		prjCmdBuilder := testPrjCmdBuilder{
			Resp: struct {
				prjCtxs []command.ProjectContext
				err     error
			}{
				prjCtxs: []command.ProjectContext{},
				err:     errors.New("error"),
			},
		}

		outputGenerator := policies.CommandOutputGenerator{
			PrjCommandBuilder: &prjCmdBuilder,
		}

		_, err := outputGenerator.GeneratePolicyCheckOutputStore(&ctx, &comment)
		assert.EqualError(t, err, "error")
	})

	t.Run("only runs policy check commands", func(t *testing.T) {
		ctx := command.Context{
			RequestCtx: context.Background(),
			Log:        logging.NewNoopCtxLogger(t),
		}

		comment := command.Comment{
			Name: command.ApprovePolicies,
		}

		result := command.ProjectResult{
			Failure: "Policies Failed",
		}

		planPrjCtx := command.ProjectContext{
			CommandName: command.Plan,
		}

		policyCheckPrjCtx := command.ProjectContext{
			CommandName: command.PolicyCheck,
			ProjectName: "project",
			Workspace:   "workspace",
		}

		prjCmdBuilder := testPrjCmdBuilder{
			Resp: struct {
				prjCtxs []command.ProjectContext
				err     error
			}{
				prjCtxs: []command.ProjectContext{
					planPrjCtx, policyCheckPrjCtx,
				},
				err: nil,
			},
		}

		prjCmdRunner := strictTestPolicyCheckCommandRunner{
			runners: []*testPolicyCheckCommandRunner{
				{
					expectedPrjCtx: policyCheckPrjCtx,
					result:         result,
				},
			},
		}

		outputGenerator := policies.CommandOutputGenerator{
			PrjCommandBuilder: &prjCmdBuilder,
			PrjCommandRunner:  &prjCmdRunner,
		}

		store, err := outputGenerator.GeneratePolicyCheckOutputStore(&ctx, &comment)
		assert.Nil(t, err)
		assert.Equal(t, store.Get("project", "workspace").PolicyCheckOutput, "Policies Failed")
	})

}

type strictTestPolicyCheckCommandRunner struct {
	t *testing.T

	runners []*testPolicyCheckCommandRunner
	count   int
}

func (t *strictTestPolicyCheckCommandRunner) PolicyCheck(prjCtx command.ProjectContext) command.ProjectResult {
	if t.count > len(t.runners)-1 {
		t.t.FailNow()
	}
	res := t.runners[t.count].PolicyCheck(prjCtx)
	t.count++
	return res
}

type testPolicyCheckCommandRunner struct {
	expectedPrjCtx command.ProjectContext

	called bool
	result command.ProjectResult
}

func (t *testPolicyCheckCommandRunner) PolicyCheck(prjCtx command.ProjectContext) command.ProjectResult {
	t.expectedPrjCtx = prjCtx
	t.called = true
	return t.result
}

type testPrjCmdBuilder struct {
	Resp struct {
		prjCtxs []command.ProjectContext
		err     error
	}
}

func (t *testPrjCmdBuilder) BuildPlanCommands(ctx *command.Context, comment *command.Comment) ([]command.ProjectContext, error) {
	return t.Resp.prjCtxs, t.Resp.err
}

func (t *testPrjCmdBuilder) BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error) {
	return []command.ProjectContext{}, nil
}
