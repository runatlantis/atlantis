package raw

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

// Config is the representation for the whole config file at the top level.
type Config struct {
	Version   *int                `yaml:"version,omitempty"`
	Projects  []Project           `yaml:"projects,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows,omitempty"`
}

func (c Config) Validate() error {
	equals2 := func(value interface{}) error {
		if *value.(*int) != 2 {
			return errors.New("must equal 2")
		}
		return nil
	}
	return validation.ValidateStruct(&c,
		validation.Field(&c.Version, validation.NotNil, validation.By(equals2)),
		validation.Field(&c.Projects),
		validation.Field(&c.Workflows),
	)
}

func (c Config) ToValid() valid.Config {
	var validProjects []valid.Project
	for _, p := range c.Projects {
		validProjects = append(validProjects, p.ToValid())
	}

	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range c.Workflows {
		validWorkflows[k] = v.ToValid()
	}
	return valid.Config{
		Version:   *c.Version,
		Projects:  validProjects,
		Workflows: validWorkflows,
	}
}
