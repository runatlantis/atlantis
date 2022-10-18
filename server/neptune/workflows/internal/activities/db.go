package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/activities/deployment"
)

type store interface {
	GetDeploymentInfo(ctx context.Context, repoName string, rootName string) (*deployment.Info, error)
	SetDeploymentInfo(ctx context.Context, deploymentInfo *deployment.Info) error
}

type dbActivities struct {
	DeploymentInfoStore store
}

type FetchLatestDeploymentRequest struct {
	FullRepositoryName string
	RootName           string
}

type FetchLatestDeploymentResponse struct {
	DeploymentInfo *deployment.Info
}

func (a *dbActivities) FetchLatestDeployment(ctx context.Context, request FetchLatestDeploymentRequest) (FetchLatestDeploymentResponse, error) {
	deploymentInfo, err := a.DeploymentInfoStore.GetDeploymentInfo(ctx, request.FullRepositoryName, request.RootName)
	if err != nil {
		return FetchLatestDeploymentResponse{}, errors.Wrapf(err, "fetching deployment info for %s/%s", request.FullRepositoryName, request.RootName)
	}

	return FetchLatestDeploymentResponse{
		DeploymentInfo: deploymentInfo,
	}, nil
}

type StoreLatestDeploymentRequest struct {
	DeploymentInfo *deployment.Info
}

func (a *dbActivities) StoreLatestDeployment(ctx context.Context, request StoreLatestDeploymentRequest) error {
	err := a.DeploymentInfoStore.SetDeploymentInfo(ctx, request.DeploymentInfo)
	if err != nil {
		return errors.Wrapf(err, "uploading deployment info for %s/%s [%s] ", request.DeploymentInfo.Repo.GetFullName(), request.DeploymentInfo.Root.Name, request.DeploymentInfo.ID)
	}

	return nil
}
