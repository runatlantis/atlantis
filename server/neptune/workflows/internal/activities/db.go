package activities

import (
	"context"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

type store interface {
	GetDeploymentInfo(ctx context.Context, repoName string, rootName string) (*root.DeploymentInfo, error)
	SetDeploymentInfo(ctx context.Context, deploymentInfo root.DeploymentInfo) error
}

type dbActivities struct {
	DeploymentInfoStore store
}

type FetchLatestDeploymentRequest struct {
	FullRepositoryName string
	RootName           string
}

type FetchLatestDeploymentResponse struct {
	DeploymentInfo *root.DeploymentInfo
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
	DeploymentInfo root.DeploymentInfo
}

func (a *dbActivities) StoreLatestDeployment(ctx context.Context, request StoreLatestDeploymentRequest) error {
	err := a.DeploymentInfoStore.SetDeploymentInfo(ctx, request.DeploymentInfo)
	if err != nil {
		return errors.Wrapf(err, "uploading deployment info for %s/%s [%s] ", request.DeploymentInfo.Repo.GetFullName(), request.DeploymentInfo.Root.Name, request.DeploymentInfo.ID)
	}

	return nil
}
