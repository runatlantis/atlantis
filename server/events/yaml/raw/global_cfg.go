package raw

import (
	"fmt"
	"regexp"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

// GlobalCfg is the raw schema for server-side repo config.
type GlobalCfg struct {
	Repos     []Repo              `yaml:"repos" json:"repos"`
	Workflows map[string]Workflow `yaml:"workflows" json:"workflows"`
}

// Repo is the raw schema for repos in the server-side repo config.
type Repo struct {
	ID                   string   `yaml:"id" json:"id"`
	ApplyRequirements    []string `yaml:"apply_requirements" json:"apply_requirements"`
	Workflow             *string  `yaml:"workflow,omitempty" json:"workflow,omitempty"`
	AllowedWorkflows     []string `yaml:"allowed_workflows,omitempty" json:"allowed_workflows,omitempty"`
	AllowedOverrides     []string `yaml:"allowed_overrides" json:"allowed_overrides"`
	AllowCustomWorkflows *bool    `yaml:"allow_custom_workflows,omitempty" json:"allow_custom_workflows,omitempty"`
}

func (g GlobalCfg) Validate() error {
	err := validation.ValidateStruct(&g,
		validation.Field(&g.Repos),
		validation.Field(&g.Workflows))
	if err != nil {
		return err
	}

	// Check that all workflows referenced by repos are actually defined.
	for _, repo := range g.Repos {
		if repo.Workflow == nil {
			continue
		}
		name := *repo.Workflow
		if name == valid.DefaultWorkflowName {
			// The 'default' workflow will always be defined.
			continue
		}
		found := false
		for w := range g.Workflows {
			if w == name {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("workflow %q is not defined", name)
		}
	}

	// Check that all allowed workflows are defined
	for _, repo := range g.Repos {
		if repo.AllowedWorkflows == nil {
			continue
		}
		for _, name := range repo.AllowedWorkflows {
			if name == valid.DefaultWorkflowName {
				// The 'default' workflow will always be defined.
				continue
			}
			found := false
			for w := range g.Workflows {
				if w == name {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("workflow %q is not defined", name)
			}
		}
	}
	return nil
}

func (g GlobalCfg) ToValid(defaultCfg valid.GlobalCfg) valid.GlobalCfg {
	workflows := make(map[string]valid.Workflow)
	for k, v := range g.Workflows {
		validatedWorkflow := v.ToValid(k)
		workflows[k] = validatedWorkflow
		if k == valid.DefaultWorkflowName {
			// Handle the special case where they're redefining the default
			// workflow. In this case, our default repo config references
			// the "old" default workflow and so needs to be redefined.
			defaultCfg.Repos[0].Workflow = &validatedWorkflow
		}
	}
	// Merge in defaults without overriding.
	for k, v := range defaultCfg.Workflows {
		if _, ok := workflows[k]; !ok {
			workflows[k] = v
		}
	}

	var repos []valid.Repo
	for _, r := range g.Repos {
		repos = append(repos, r.ToValid(workflows))
	}
	repos = append(defaultCfg.Repos, repos...)
	return valid.GlobalCfg{
		Repos:     repos,
		Workflows: workflows,
	}
}

// HasRegexID returns true if r is configured with a regex id instead of an
// exact match id.
func (r Repo) HasRegexID() bool {
	return strings.HasPrefix(r.ID, "/") && strings.HasSuffix(r.ID, "/")
}

func (r Repo) Validate() error {
	idValid := func(value interface{}) error {
		id := value.(string)
		if !r.HasRegexID() {
			return nil
		}
		_, err := regexp.Compile(id[1 : len(id)-1])
		return errors.Wrapf(err, "parsing: %s", id)
	}

	overridesValid := func(value interface{}) error {
		overrides := value.([]string)
		for _, o := range overrides {
			if o != valid.ApplyRequirementsKey && o != valid.WorkflowKey {
				return fmt.Errorf("%q is not a valid override, only %q and %q are supported", o, valid.ApplyRequirementsKey, valid.WorkflowKey)
			}
		}
		return nil
	}

	workflowExists := func(value interface{}) error {
		// We validate workflows in ParserValidator.validateRepoWorkflows
		// because we need the list of workflows to validate.
		return nil
	}

	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required, validation.By(idValid)),
		validation.Field(&r.AllowedOverrides, validation.By(overridesValid)),
		validation.Field(&r.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&r.Workflow, validation.By(workflowExists)),
	)
}

func (r Repo) ToValid(workflows map[string]valid.Workflow) valid.Repo {
	var id string
	var idRegex *regexp.Regexp
	if r.HasRegexID() {
		withoutSlashes := r.ID[1 : len(r.ID)-1]
		// Safe to use MustCompile because we test it in Validate().
		idRegex = regexp.MustCompile(withoutSlashes)
	} else {
		id = r.ID
	}

	var workflow *valid.Workflow
	if r.Workflow != nil {
		// This key is guaranteed to exist because we test for it in
		// ParserValidator.validateRepoWorkflows.
		ptr := workflows[*r.Workflow]
		workflow = &ptr
	}

	return valid.Repo{
		ID:                   id,
		IDRegex:              idRegex,
		ApplyRequirements:    r.ApplyRequirements,
		Workflow:             workflow,
		AllowedWorkflows:     r.AllowedWorkflows,
		AllowedOverrides:     r.AllowedOverrides,
		AllowCustomWorkflows: r.AllowCustomWorkflows,
	}
}
