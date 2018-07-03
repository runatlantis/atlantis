package raw

import "github.com/runatlantis/atlantis/server/events/yaml/valid"

const DefaultAutoPlanWhenModified = "**/*.tf"
const DefaultAutoPlanEnabled = true

type Autoplan struct {
	WhenModified []string `yaml:"when_modified,omitempty"`
	Enabled      *bool    `yaml:"enabled,omitempty"`
}

func (a Autoplan) ToValid() valid.Autoplan {
	var v valid.Autoplan
	if a.WhenModified == nil {
		v.WhenModified = []string{DefaultAutoPlanWhenModified}
	} else {
		v.WhenModified = a.WhenModified
	}

	if a.Enabled == nil {
		v.Enabled = true
	} else {
		v.Enabled = *a.Enabled
	}

	return v
}

func (a Autoplan) Validate() error {
	return nil
}

func DefaultAutoPlan() valid.Autoplan {
	return valid.Autoplan{
		WhenModified: []string{DefaultAutoPlanWhenModified},
		Enabled:      DefaultAutoPlanEnabled,
	}
}
