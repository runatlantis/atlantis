package activities

import (
	"context"
	"net/url"
)

type workerInfoActivity struct {
	ServerURL *url.URL
}

type GetWorkerInfoResponse struct {
	ServerURL *url.URL
}

// GetWorkerInfo exists because this is the only way to pass host level info from worker construction
// to our workflows. This should be invoked as part of a session.
func (a *workerInfoActivity) GetWorkerInfo(ctx context.Context) (*GetWorkerInfoResponse, error) {
	return &GetWorkerInfoResponse{
		ServerURL: a.ServerURL,
	}, nil

}
