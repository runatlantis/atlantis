package raw

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

// DefaultAutomerge is the default setting for automerge.
const DefaultAutomerge = false

// RepoCfg is the raw schema for repo-level atlantis.yaml config.
type RepoCfg struct {
	Version   *int                `yaml:"version,omitempty"`
	Projects  []Project           `yaml:"projects,omitempty"`
	Workflows map[string]Workflow `yaml:"workflows,omitempty"`
	Automerge *bool               `yaml:"automerge,omitempty"`
}

func (r RepoCfg) Validate() error {
	equals2 := func(value interface{}) error {
		asIntPtr := value.(*int)
		if asIntPtr == nil {
			return errors.New("is required. If you've just upgraded Atlantis you need to rewrite your atlantis.yaml for version 3. See www.runatlantis.io/docs/upgrading-atlantis-yaml.html")
		}
		if *asIntPtr != 2 && *asIntPtr != 3 {
			return errors.New("only versions 2 and 3 are supported")
		}
		return nil
	}
	return validation.ValidateStruct(&r,
		validation.Field(&r.Version, validation.By(equals2)),
		validation.Field(&r.Projects),
		validation.Field(&r.Workflows),
	)
}

func (r RepoCfg) ToValid() valid.RepoCfg {
	validWorkflows := make(map[string]valid.Workflow)
	for k, v := range r.Workflows {
		validWorkflows[k] = v.ToValid(k)
	}

	var validProjects []valid.Project
	for _, p := range r.Projects {
		validProjects = append(validProjects, p.ToValid())
	}

	automerge := DefaultAutomerge
	if r.Automerge != nil {
		automerge = *r.Automerge
	}

	return valid.RepoCfg{
		Version:   *r.Version,
		Projects:  validProjects,
		Workflows: validWorkflows,
		Automerge: automerge,
	}
}
