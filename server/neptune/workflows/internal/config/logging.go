package config

const (

	// keeping the same as what's defined here:
	// https://github.com/lyft/atlantis/blob/cca4b5e552da1f5acd7f06350b2d172dee14218b/server/logging/logger.go#L37-L38
	// to preserve history
	RevisionLogKey     = "sha"
	RepositoryLogKey   = "repository"
	ProjectLogKey      = "project"
	GHRequestIDLogKey  = "gh-request-id"
	DeploymentIDLogKey = "deployment-id"
)

type KVStore interface {
	Value(key interface{}) interface{}
}

func ExtractLogKeyFields(ctx KVStore) []interface{} {
	var args []interface{}

	for _, k := range []string{GHRequestIDLogKey, RepositoryLogKey, RevisionLogKey, ProjectLogKey, DeploymentIDLogKey} {
		if v, ok := ctx.Value(k).(string); ok {
			args = append(args, k, v)
		}
	}

	return args
}