package policy

import (
	"context"
	"fmt"
	"github.com/palantir/go-githubapp/githubapp"
	runtime_models "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/metrics"
	"github.com/runatlantis/atlantis/server/vcs/provider/github"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

type policyFilter interface {
	Filter(ctx context.Context, installationToken int64, repo models.Repo, prNum int, failedPolicies []valid.PolicySet) ([]valid.PolicySet, error)
}

type exec interface {
	CombinedOutput(args []string, envs map[string]string, workdir string) (string, error)
}

const (
	conftestScope = "conftest.policies"
	// use internal server error message for user to understand error is from atlantis
	internalError = "internal server error"
)

// ConfTestExecutor runs a versioned conftest binary with the args built from the project context.
// Project context defines whether conftest runs a local policy set or runs a test on a remote policy set.
type ConfTestExecutor struct {
	Exec         exec
	PolicyFilter policyFilter
}

func NewConfTestExecutor(creator githubapp.ClientCreator, org string) *ConfTestExecutor {
	reviewsFetcher := &github.PRReviewerFetcher{
		ClientCreator: creator,
	}
	teamMemberFetcher := &github.TeamMemberFetcher{
		ClientCreator: creator,
		Org:           org,
	}
	return &ConfTestExecutor{
		Exec:         runtime_models.LocalExec{},
		PolicyFilter: events.NewApprovedPolicyFilter(reviewsFetcher, teamMemberFetcher),
	}
}

// Run performs conftest policy tests against changes and fails if any policy does not pass. It also runs an all-or-nothing
// filter that will filter out all policy failures based on the filter criteria.
func (c *ConfTestExecutor) Run(_ context.Context, prjCtx command.ProjectContext, executablePath string, envs map[string]string, workdir string, extraArgs []string) (string, error) {
	var policyNames []string
	var failedPolicies []valid.PolicySet
	var totalCmdOutput []string
	var policyErr error

	inputFile := filepath.Join(workdir, prjCtx.GetShowResultFileName())
	scope := prjCtx.Scope.SubScope(conftestScope)

	for _, policySet := range prjCtx.PolicySets.PolicySets {
		var policyArgs []Arg
		for _, path := range policySet.Paths {
			policyArgs = append(policyArgs, NewPolicyArg(path))
		}
		policyNames = append(policyNames, policySet.Name)
		args := ConftestTestCommandArgs{
			PolicyArgs: policyArgs,
			ExtraArgs:  extraArgs,
			InputFile:  inputFile,
			Command:    executablePath,
		}
		serializedArgs, err := args.build()
		if err != nil {
			prjCtx.Log.WarnContext(prjCtx.RequestCtx, "No policies have been configured")
			scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
			return "", errors.Wrap(err, "building args")
		}

		cmdOutput, cmdErr := c.Exec.CombinedOutput(serializedArgs, envs, workdir)
		// Continue running other policies if one fails since it might not be the only failing one
		if cmdErr != nil {
			policyErr = cmdErr
			failedPolicies = append(failedPolicies, policySet)
		}
		totalCmdOutput = append(totalCmdOutput, cmdOutput)
	}

	title := c.buildTitle(policyNames)
	output := c.sanitizeOutput(inputFile, title+strings.Join(totalCmdOutput, "\n"))
	if prjCtx.InstallationToken == 0 {
		prjCtx.Log.ErrorContext(prjCtx.RequestCtx, "missing installation token")
		scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return output, errors.New(internalError)
	}

	failedPolicies, err := c.PolicyFilter.Filter(prjCtx.RequestCtx, prjCtx.InstallationToken, prjCtx.HeadRepo, prjCtx.Pull.Num, failedPolicies)
	if err != nil {
		prjCtx.Log.ErrorContext(prjCtx.RequestCtx, fmt.Sprintf("error filtering out approved policies: %s", err.Error()))
		scope.Counter(metrics.ExecutionErrorMetric).Inc(1)
		return output, errors.New(internalError)
	}
	if len(failedPolicies) == 0 {
		scope.Counter(metrics.ExecutionSuccessMetric).Inc(1)
		return output, nil
	}
	// use policyErr here as policy error output is what the user should see
	scope.Counter(metrics.ExecutionFailureMetric).Inc(1)
	return output, policyErr
}

func (c *ConfTestExecutor) buildTitle(policySetNames []string) string {
	return fmt.Sprintf("Checking plan against the following policies: \n  %s\n", strings.Join(policySetNames, "\n  "))
}

func (c *ConfTestExecutor) sanitizeOutput(inputFile string, output string) string {
	return strings.Replace(output, inputFile, "<redacted plan file>", -1)
}
