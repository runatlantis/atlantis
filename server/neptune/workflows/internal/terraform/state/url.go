package state

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

const (
	JobIDParamKey = "job-id"
)

// outputURLBuilder builds a URL for the route.
//
// It accepts a sequence of key/value pairs for the route variables.
type outputURLBuilder interface {
	URL(pairs ...string) (*url.URL, error)
}

type OutputURLGenerator struct {
	URLBuilder outputURLBuilder
}

func (g *OutputURLGenerator) Generate(jobID fmt.Stringer, BaseURL fmt.Stringer) (*url.URL, error) {
	jobIDParam := jobID.String()

	jobURL, err := g.URLBuilder.URL(
		JobIDParamKey, jobIDParam,
	)

	if err != nil {
		return nil, errors.Wrapf(err, "creating job url for %s", jobIDParam)
	}

	combined := BaseURL.String() + jobURL.String()
	return url.Parse(combined)
}
