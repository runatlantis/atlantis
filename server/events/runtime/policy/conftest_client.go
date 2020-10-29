package policy

import (
	"fmt"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

// SourceResolver resolves the policy set to a local fs path
type SourceResolver interface {
	Resolve(policySet models.PolicySet) (string, error)
}

// LocalSourceResolver resolves a local policy set to a local fs path
type LocalSourceResolver struct {
}

func (p *LocalSourceResolver) Resolve(policySet models.PolicySet) (string, error) {
	return "some/path", nil

}

// SourceResolverProxy proxies to underlying source resolvers dynamically
type SourceResolverProxy struct {
	localSourceResolver SourceResolver
}

func (p *SourceResolverProxy) Resolve(policySet models.PolicySet) (string, error) {
	switch source := policySet.Source; source {
	case models.LocalPolicySet:
		return p.localSourceResolver.Resolve(policySet)
	default:
		return "", errors.New(fmt.Sprintf("unable to resolve policy set source %s", source))
	}
}

// ConfTestExecutorWorkflow runs a versioned conftest binary with the args built from the project context.
// Project context defines whether conftest runs a local policy set or runs a test on a remote policy set.
type ConfTestExecutorWorkflow struct {
	SourceResolver SourceResolver
}

func NewConfTestExecutorWorkflow() *ConfTestExecutorWorkflow {
	return &ConfTestExecutorWorkflow{
		SourceResolver: &SourceResolverProxy{
			localSourceResolver: &LocalSourceResolver{},
		},
	}
}

func (c *ConfTestExecutorWorkflow) Run(log *logging.SimpleLogger, executablePath string, envs map[string]string, args []string) (string, error) {
	return "success", nil

}

func (c *ConfTestExecutorWorkflow) EnsureExecutorVersion(log *logging.SimpleLogger, v *version.Version) (string, error) {
	return "some/path", nil

}

func (c *ConfTestExecutorWorkflow) ResolveArgs(ctx models.ProjectCommandContext) ([]string, error) {
	return []string{""}, nil
}
