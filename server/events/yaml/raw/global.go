package raw

import (
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"regexp"
	"strings"
)

const ApplyRequirementsKey string = "apply_requirements"
const WorkflowKey string = "workflow"
const CustomWorkflowsKey string = "workflows"
const AllowedOverridesKey string = "allowed_overrides"

type GlobalCfg struct {
	Repos     []Repo              `yaml:"repos"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Repo struct {
	ID                   string   `yaml:"id"`
	ApplyRequirements    []string `yaml:"apply_requirements"`
	Workflow             *string  `yaml:"workflow,omitempty"`
	AllowedOverrides     []string `yaml:"allowed_overrides"`
	AllowCustomWorkflows *bool    `yaml:"allow_custom_workflows,omitempty"`
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
	return nil
}

func (g GlobalCfg) ToValid() valid.GlobalCfg {
	workflows := make(map[string]valid.Workflow)
	for k, v := range g.Workflows {
		workflows[k] = v.ToValid(k)
	}

	var repos []valid.Repo
	for _, r := range g.Repos {
		repos = append(repos, r.ToValid(workflows))
	}
	return valid.GlobalCfg{
		Repos:     repos,
		Workflows: workflows,
	}
}

func (r Repo) HasRegexID() bool {
	return strings.HasPrefix(r.ID, "/") && strings.HasSuffix(r.ID, "/")
}

func (r Repo) Validate() error {
	idValid := func(value interface{}) error {
		id := value.(string)
		if !r.HasRegexID() {
			return nil
		}
		_, err := regexp.Compile(strings.Trim(id, "/"))
		return err
	}

	overridesValid := func(value interface{}) error {
		overrides := value.([]string)
		for _, o := range overrides {
			if o != ApplyRequirementsKey && o != WorkflowKey {
				return fmt.Errorf("%q is invalid, only %s and %s are supported", o, ApplyRequirementsKey, WorkflowKey)
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
		// Safe to use MustCompile because we test it in Validate().
		idRegex = regexp.MustCompile(strings.Trim(id, "/"))
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
		AllowedOverrides:     r.AllowedOverrides,
		AllowCustomWorkflows: r.AllowCustomWorkflows,
	}
}
