package context

import "context"

type Key string

func (c Key) String() string {
	return string(c)
}

const (
	InstallationIDKey = Key("gh-installation-id")
	RequestIDKey      = Key("gh-request-id")
	RepositoryKey     = Key("repository")
	SHAKey            = Key("sha")
	PullNumKey        = Key("pull-num")
	ProjectKey        = Key("project")
	Err               = Key("err")
)

var Keys = []Key{RequestIDKey, RepositoryKey, PullNumKey, ProjectKey, SHAKey, InstallationIDKey, Err}

// Extracts relevant fields from context for structured logging.
func ExtractFields(ctx context.Context) map[string]interface{} {
	args := make(map[string]interface{})

	for _, k := range Keys {
		if ctx.Value(k) == nil {
			continue
		}
		args[k.String()] = ctx.Value(k)
	}

	return args
}

// Copies fields from a context to a new context created from a given base.
func CopyFields(base context.Context, from context.Context) context.Context { //nolint:golint // avoiding refactor while adding linter action
	for _, k := range Keys {
		base = context.WithValue(base, k, from.Value(k))
	}
	return base
}
