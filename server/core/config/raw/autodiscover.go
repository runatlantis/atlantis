package raw

import (
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Autodiscover struct {
	Enabled *bool `yaml:"enabled,omitempty"`
}

func (a Autodiscover) ToValid() valid.Autodiscover {
	var v valid.Autodiscover

	if a.Enabled == nil {
		v.Enabled = true
	} else {
		v.Enabled = *a.Enabled
	}

	return v
}

func (a Autodiscover) Validate() error {
	return nil
}

// DefaultAutoDiscover returns the default autodiscover config.
func DefaultAutoDiscover() valid.Autodiscover {
	return valid.Autodiscover{
		Enabled: valid.DefaultAutoDiscoverEnabled,
	}
}
