package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/graymeta/stow"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type ContainerNotFoundError struct {
	Err error
}

func (c *ContainerNotFoundError) Error() string {
	return errors.Wrap(c.Err, "container not found").Error()
}

type ItemNotFoundError struct {
	Err error
}

func (i *ItemNotFoundError) Error() string {
	return errors.Wrap(i.Err, "item not found").Error()
}

func NewClient(storeConfig valid.StoreConfig) (*Client, error) {
	location, err := stow.Dial(string(storeConfig.BackendType), storeConfig.Config)
	if err != nil {
		return nil, errors.Wrap(err, "intializing stow client")
	}

	return &Client{
		Location:      location,
		ContainerName: storeConfig.ContainerName,
		Prefix:        storeConfig.Prefix,
	}, nil
}

type containerResolver interface {
	Container(id string) (stow.Container, error)
}

type Client struct {
	Location      containerResolver
	ContainerName string
	Prefix        string
}

// Return custom errors for the caller to be able to distinguish when container is not found vs item is not found
func (c *Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	container, err := c.Location.Container(c.ContainerName)
	if err != nil {
		return nil, &ContainerNotFoundError{
			Err: err,
		}
	}

	key = c.addPrefix(key)
	item, err := container.Item(key)
	if err != nil {
		if errors.Is(err, stow.ErrNotFound) {
			return nil, &ItemNotFoundError{
				Err: err,
			}
		}
		return nil, errors.Wrap(err, "getting item")
	}

	r, err := item.Open()
	if err != nil {
		return nil, errors.Wrap(err, "reading item")
	}

	return r, nil
}

func (c *Client) Set(ctx context.Context, key string, object []byte) error {
	container, err := c.Location.Container(c.ContainerName)
	if err != nil {
		return errors.Wrap(err, "resolving container")
	}

	key = c.addPrefix(key)
	_, err = container.Put(key, bytes.NewReader(object), int64(len(object)), nil)
	if err != nil {
		return errors.Wrap(err, "writing to container")
	}
	return nil
}

func (c *Client) addPrefix(key string) string {
	return fmt.Sprintf("%s/%s", c.Prefix, key)
}
