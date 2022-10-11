package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
)

type client interface {
	Get(ctx context.Context, key string) (io.ReadCloser, error)
	Set(ctx context.Context, key string, object []byte) error
}

func NewStore(stowClient client) (*Store, error) {
	return &Store{
		stowClient: stowClient,
	}, nil

}

type Store struct {
	stowClient client
}

func (s *Store) GetDeploymentInfo(ctx context.Context, repoName string, rootName string) (*root.DeploymentInfo, error) {
	key := BuildKey(repoName, rootName)

	reader, err := s.stowClient.Get(ctx, key)
	if err != nil {
		switch err.(type) {

		// Fail if container is not found
		case *storage.ContainerNotFoundError:
			return nil, err

		// First deploy for this root
		case *storage.ItemNotFoundError:
			return nil, nil

		default:
			return nil, errors.Wrap(err, "getting item")
		}
	}
	defer reader.Close()

	decoder := json.NewDecoder(reader)

	var deploymentInfo root.DeploymentInfo
	err = decoder.Decode(&deploymentInfo)
	if err != nil {
		return nil, errors.Wrap(err, "decoding item")
	}

	return &deploymentInfo, nil
}

func (s *Store) SetDeploymentInfo(ctx context.Context, deploymentInfo root.DeploymentInfo) error {
	key := BuildKey(deploymentInfo.Repo.GetFullName(), deploymentInfo.Root.Name)
	object, err := json.Marshal(deploymentInfo)
	if err != nil {
		return errors.Wrap(err, "marshalling deployment info")
	}

	err = s.stowClient.Set(ctx, key, object)
	if err != nil {
		return errors.Wrap(err, "writing to store")
	}
	return nil
}

func BuildKey(repo string, root string) string {
	return fmt.Sprintf("%s/%s/deployment.json", repo, root)
}
