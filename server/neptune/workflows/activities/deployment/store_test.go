package deployment_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	"github.com/stretchr/testify/assert"
)

type testStowClient struct {
	t   *testing.T
	get struct {
		key        string
		readCloser io.ReadCloser
		err        error
	}
}

func (t *testStowClient) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	assert.Equal(t.t, t.get.key, key)
	return t.get.readCloser, t.get.err
}

// Unused
func (t *testStowClient) Set(ctx context.Context, key string, object []byte) error {
	return nil
}

func TestStore_GetDeploymentInfo(t *testing.T) {
	repoName := "repo"
	rootName := "root"
	key := deployment.BuildKey(repoName, rootName)
	clientErr := errors.New("error")

	t.Run("return error when container not found", func(t *testing.T) {
		stowClient := &testStowClient{
			t: t,
			get: struct {
				key        string
				readCloser io.ReadCloser
				err        error
			}{
				key: key,
				err: &storage.ContainerNotFoundError{Err: clientErr},
			},
		}
		store, err := deployment.NewStore(stowClient)
		assert.Nil(t, err)

		deploymentInfo, err := store.GetDeploymentInfo(context.TODO(), repoName, rootName)
		assert.Nil(t, deploymentInfo)
		assert.Equal(t, err, &storage.ContainerNotFoundError{Err: clientErr})
	})

	t.Run("empty deployment object when item not found", func(t *testing.T) {
		stowClient := &testStowClient{
			t: t,
			get: struct {
				key        string
				readCloser io.ReadCloser
				err        error
			}{
				key: key,
				err: &storage.ItemNotFoundError{Err: clientErr},
			},
		}
		store, err := deployment.NewStore(stowClient)
		assert.Nil(t, err)

		deploymentInfo, err := store.GetDeploymentInfo(context.TODO(), repoName, rootName)
		assert.Nil(t, err)
		assert.Nil(t, deploymentInfo)

	})
}
