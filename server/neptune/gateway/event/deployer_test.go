package event_test

import (
	"context"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	"github.com/stretchr/testify/assert"
	"go.temporal.io/sdk/client"
	"testing"
)

func TestDeploy(t *testing.T) {
	logger := logging.NewNoopCtxLogger(t)
	version, err := version.NewVersion("1.0.3")
	assert.NoError(t, err)
	// use default values for testing
	deployOptions := event.RootDeployOptions{}

	t.Run("root config builder error", func(t *testing.T) {
		signaler := &mockDeploySignaler{}
		ctx := context.Background()
		deployer := event.RootDeployer{
			DeploySignaler:    signaler,
			Logger:            logger,
			RootConfigBuilder: &mockRootConfigBuilder{error: assert.AnError},
		}

		err := deployer.Deploy(ctx, deployOptions)
		assert.Error(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("not platform mode", func(t *testing.T) {
		ctx := context.Background()
		signaler := &mockDeploySignaler{}
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
			WorkflowMode:     valid.DefaultWorkflowMode,
		}
		rootCfgs := []*valid.MergedProjectCfg{
			&rootCfg,
		}
		rootConfigBuilder := &mockRootConfigBuilder{
			rootConfigs: rootCfgs,
		}
		deployer := event.RootDeployer{
			DeploySignaler:    signaler,
			Logger:            logger,
			RootConfigBuilder: rootConfigBuilder,
		}

		err := deployer.Deploy(ctx, deployOptions)
		assert.NoError(t, err)
		assert.False(t, signaler.called)
	})

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		signaler := &mockDeploySignaler{run: testRun{}}
		rootCfg := valid.MergedProjectCfg{
			Name: testRoot,
			DeploymentWorkflow: valid.Workflow{
				Plan:  valid.DefaultPlanStage,
				Apply: valid.DefaultApplyStage,
			},
			TerraformVersion: version,
			WorkflowMode:     valid.PlatformWorkflowMode,
		}
		rootCfgs := []*valid.MergedProjectCfg{
			&rootCfg,
		}
		rootConfigBuilder := &mockRootConfigBuilder{
			rootConfigs: rootCfgs,
		}
		deployer := event.RootDeployer{
			DeploySignaler:    signaler,
			Logger:            logger,
			RootConfigBuilder: rootConfigBuilder,
		}

		err := deployer.Deploy(ctx, deployOptions)
		assert.NoError(t, err)
		assert.True(t, signaler.called)
	})
}

type mockRootConfigBuilder struct {
	rootConfigs []*valid.MergedProjectCfg
	error       error
}

func (r *mockRootConfigBuilder) Build(_ context.Context, _ models.Repo, _ string, _ string, _ int64, _ event.BuilderOptions) ([]*valid.MergedProjectCfg, error) {
	return r.rootConfigs, r.error
}

type mockDeploySignaler struct {
	run    client.WorkflowRun
	error  error
	called bool
}

func (d *mockDeploySignaler) SignalWorkflow(_ context.Context, _ string, _ string, _ string, _ interface{}) error {
	d.called = true
	return d.error
}

func (d *mockDeploySignaler) SignalWithStartWorkflow(_ context.Context, _ *valid.MergedProjectCfg, _ event.RootDeployOptions) (client.WorkflowRun, error) {
	d.called = true
	return d.run, d.error
}
