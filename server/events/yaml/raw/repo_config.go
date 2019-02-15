package raw

import (
	"fmt"
	"github.com/go-ozzo/ozzo-validation"
	"regexp"
	"strings"
)

const ApplyRequirementsKey string = "apply_requirements"
const WorkflowKey string = "workflow"
const CustomWorkflowsKey string = "workflows"
const AllowedOverridesKey string = "allowed_overrides"

type RepoConfig struct {
	Repos     []Repo              `yaml:"repos"`
	Workflows map[string]Workflow `yaml:"workflows"`
}

type Repo struct {
	ID                   string   `yaml:"id"`
	ApplyRequirements    []string `yaml:"apply_requirements"`
	Workflow             *string  `yaml:"workflow,omitempty"`
	AllowedOverrides     []string `yaml:"allowed_overrides"`
	AllowCustomWorkflows bool     `yaml:"allow_custom_workflows"`
}

func (r RepoConfig) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.Repos),
		validation.Field(&r.Workflows))
}

func (r Repo) Validate() error {
	return validation.ValidateStruct(&r,
		validation.Field(&r.ID, validation.Required),
		validation.Field(&r.AllowedOverrides, validation.By(checkOverrideValues)),
	)
}

func checkOverrideValues(values interface{}) error {
	if allValues, _ := values.([]string); allValues != nil {
		validOverrides := []string{"apply_requirements", "workflow"}
		for _, value := range allValues {
			for _, validStr := range validOverrides {
				if value == validStr {
					return nil
				}
			}
		}
		return fmt.Errorf("value must be one of %v", validOverrides)
	}
	return nil
}

func (r *Repo) IsOverrideAllowed(override string) bool {
	for _, allowed := range r.AllowedOverrides {
		if allowed == override {
			return true
		}
	}
	return false
}

func (r *Repo) Matches(repoName string) (bool, error) {
	if strings.Index(r.ID, "/") == 0 && strings.LastIndex(r.ID, "/") == len(r.ID)-1 {
		matchString := strings.Trim(r.ID, "/")
		compiled, err := regexp.Compile(matchString)
		if err != nil {
			return false, fmt.Errorf("regex compile of repo.ID `%s`: %s", r.ID, err)
		}
		return compiled.MatchString(repoName), nil
	}
	return repoName == r.ID, nil
}
