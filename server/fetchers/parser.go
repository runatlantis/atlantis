package fetchers

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

type ConfigSourceType int

const (
	Github ConfigSourceType = iota
	// Other source types should be added later here when their fetchers are implemented, like:
	// S3
	// Bitbucket
	// etc.
)

type Parser interface {
	ParseConnection(remoteReference string, token string) error
}

func GetType(remoteReference string) (ConfigSourceType, error) {
	u, err := url.Parse(remoteReference)

	if err != nil {
		return 0, errors.Wrap(err, "")
	}
	switch {
	case u.Hostname() == "github.com":
		return Github, nil
	default:
		return 0, errors.New(fmt.Sprintf("unknown source in remote reference: %s", remoteReference))
	}
}
