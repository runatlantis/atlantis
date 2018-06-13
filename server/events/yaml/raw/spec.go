package raw

import (
	"errors"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

type Spec struct {
	Version   *int                `yaml:"version,omitempty"`
	Projects  []Project           `yaml:"projects,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows,omitempty"`
}

func (s Spec) Validate() error {
	equals2 := func(value interface{}) error {
		if *value.(*int) != 2 {
			return errors.New("must equal 2")
		}
		return nil
	}
	return validation.ValidateStruct(&s,
		validation.Field(&s.Version, validation.NotNil, validation.By(equals2)),
		validation.Field(&s.Projects),
		validation.Field(&s.Workflows),
	)
}

func (s Spec) ToValid() valid.Spec {
	var validProjects []valid.Project
	for _, p := range s.Projects {
		validProjects = append(validProjects, p.ToValid())
	}

	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range s.Workflows {
		validWorkflows[k] = v.ToValid()
	}
	return valid.Spec{
		Version:   *s.Version,
		Projects:  validProjects,
		Workflows: validWorkflows,
	}
}
