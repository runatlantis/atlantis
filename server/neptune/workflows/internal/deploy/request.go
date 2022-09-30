package deploy

import "github.com/runatlantis/atlantis/server/neptune/workflows/internal/deploy/request"

type Request struct {
	GHRequestID string
	Repository  request.Repo
}