package wrappers

import (
	"github.com/runatlantis/atlantis/server/events"
	"github.com/uber-go/tally"
)

type projectContext struct {
	events.ProjectCommandContextBuilder
}

func WrapProjectContext(
	projectCtxBuilder events.ProjectCommandContextBuilder,
) *projectContext {
	return &projectContext{
		projectCtxBuilder,
	}
}

func (p *projectContext) WithPolicyChecks(
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
