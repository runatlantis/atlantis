package storage_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/graymeta/stow"
	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/stretchr/testify/assert"
)

type testContainer struct {
	t    *testing.T
	item struct {
		id   string
		resp stow.Item
		err  error
	}
}

func (t *testContainer) Item(id string) (stow.Item, error) {
	assert.Equal(t.t, t.item.id, id)
	return t.item.resp, t.item.err
}

// Unused methods
func (t *testContainer) Put(name string, r io.Reader, size int64, metadata map[string]interface{}) (stow.Item, error) {
	return nil, nil
}

func (t *testContainer) ID() string {
	return ""
}

func (t *testContainer) Name() string {
	return ""
}

func (t *testContainer) Items(prefix, cursor string, count int) ([]stow.Item, string, error) {
	return []stow.Item{}, "", nil
}

func (t *testContainer) RemoveItem(id string) error {
	return nil
}

type testContainerResolver struct {
	t         *testing.T
	name      string
	container stow.Container
	err       error
}

func (t *testContainerResolver) Container(name string) (stow.Container, error) {
	assert.Equal(t.t, t.name, name)
	return t.container, t.err
}

func TestClient_Get(t *testing.T) {
	id := "1234"
	containerName := "container"
	prefix := "prefix"
	expErr := errors.New("error")

	t.Run("should throw container not found error when Container not found", func(t *testing.T) {
		location := &testContainerResolver{
			t:    t,
			name: containerName,
			err:  expErr,
		}

		client := storage.Client{
			Location:      location,
			ContainerName: containerName,
			Prefix:        prefix,
		}

		readCloser, err := client.Get(context.Background(), id)
		assert.Nil(t, readCloser)
		assert.Equal(t, &storage.ContainerNotFoundError{
			Err: expErr,
		}, err)

	})

	t.Run("should throw item not found error when Item not found", func(t *testing.T) {
		container := &testContainer{
			t: t,
			item: struct {
				id   string
				resp stow.Item
				err  error
			}{
				id:  fmt.Sprintf("%s/%s", prefix, id),
				err: stow.ErrNotFound,
			},
		}

		location := &testContainerResolver{
			t:         t,
			name:      containerName,
			container: container,
		}

		client := storage.Client{
			Location:      location,
			ContainerName: containerName,
			Prefix:        prefix,
		}

		readCloser, err := client.Get(context.Background(), id)
		assert.Nil(t, readCloser)
		assert.Equal(t, &storage.ItemNotFoundError{
			Err: stow.ErrNotFound,
		}, err)
	})
}
