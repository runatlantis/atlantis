package wrappers

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/uber-go/tally/v4"
)

type projectContext struct {
	events.ProjectCommandContextBuilder
}

func WrapProjectContext(
	projectCtxBuilder events.ProjectCommandContextBuilder,
) *projectContext { //nolint:revive // avoiding refactor while adding linter action
	return &projectContext{
		projectCtxBuilder,
	}
}

func (p *projectContext) EnablePolicyChecks(
	commentBuilder events.CommentBuilder,
) *projectContext {
	p.ProjectCommandContextBuilder = &events.PolicyCheckProjectContextBuilder{
		ProjectCommandContextBuilder: p.ProjectCommandContextBuilder,
		CommentBuilder:               commentBuilder,
	}
	return p
}

func (p *projectContext) WithInstrumentation(scope tally.Scope) *projectContext {
	p.ProjectCommandContextBuilder = &events.InstrumentedProjectCommandContextBuilder{
		ProjectCommandContextBuilder: p.ProjectCommandContextBuilder,
		ProjectCounter:               scope.Counter("projects"),
	}
	return p
}
