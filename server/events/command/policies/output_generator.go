package policies

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
)

type CommandOutputGenerator struct {
	PrjCommandRunner  events.ProjectPolicyCheckCommandRunner
	PrjCommandBuilder events.ProjectPlanCommandBuilder
}

func (f *CommandOutputGenerator) GeneratePolicyCheckOutputStore(ctx *command.Context, cmd *command.Comment) (command.PolicyCheckOutputStore, error) {
	prjCmds, err := f.PrjCommandBuilder.BuildPlanCommands(ctx, &command.Comment{
		RepoRelDir:  cmd.RepoRelDir,
		Name:        command.Plan,
		Workspace:   cmd.Workspace,
		ProjectName: cmd.ProjectName,
		LogLevel:    cmd.LogLevel,
	})
	if err != nil {
		ctx.Log.WarnContext(ctx.RequestCtx, fmt.Sprintf("unable to build plan command: %s", err))
		return command.PolicyCheckOutputStore{}, err
	}

	policyCheckCommands := f.getPolicyCheckCommands(ctx, prjCmds)
	policyCheckOutputStore := command.NewPolicyCheckOutputStore()
	for _, policyCheckCmd := range policyCheckCommands {
		policyCheckResult := f.PrjCommandRunner.PolicyCheck(policyCheckCmd)

		var output string
		if policyCheckResult.Failure != "" {
			output = policyCheckResult.Failure
		} else if policyCheckResult.PolicyCheckSuccess != nil {
			output = policyCheckResult.PolicyCheckSuccess.PolicyCheckOutput
		}

		policyCheckOutputStore.Set(
			policyCheckCmd.ProjectName,
			policyCheckCmd.Workspace,
			output,
		)
	}
	return *policyCheckOutputStore, nil
}

func (f *CommandOutputGenerator) getPolicyCheckCommands(
	ctx *command.Context,
	cmds []command.ProjectContext,
) []command.ProjectContext {
	policyCheckCmds := []command.ProjectContext{}
	for _, cmd := range cmds {
		if cmd.CommandName == command.PolicyCheck {
			policyCheckCmds = append(policyCheckCmds, cmd)
		}
	}
	return policyCheckCmds
}
