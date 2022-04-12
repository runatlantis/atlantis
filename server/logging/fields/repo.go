// Consolidating fields helper functions for the new logger interface
// Once we move to context.Context we can remove this helpers package
// TODO: Remove this package once we fully move to context.Context

package fields

import (
	"strconv"

	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

func Repo(repo models.Repo) map[string]interface{} {
	return map[string]interface{}{
		logging.RepositoryKey.String(): repo.FullName,
	}
}

func PullRequest(pull models.PullRequest) map[string]interface{} {
	return map[string]interface{}{
		logging.RepositoryKey.String(): pull.BaseRepo.FullName,
		logging.PullNumKey.String():    strconv.Itoa(pull.Num),
		logging.SHAKey.String():        pull.HeadCommit,
	}
}

func PullRequestWithErr(pull models.PullRequest, err error) map[string]interface{} {
	kv := PullRequest(pull)
	kv[logging.Err.String()] = err
	return kv
}
