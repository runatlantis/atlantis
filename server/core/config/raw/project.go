package raw

import (
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation"
	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

const (
	DefaultWorkspace           = "default"
	ApprovedApplyRequirement   = "approved"
	MergeableApplyRequirement  = "mergeable"
	UnDivergedApplyRequirement = "undiverged"
	UnlockedApplyRequirement   = "unlocked"
)

var applyRequirements = map[string]bool{
	ApprovedApplyRequirement:   true,
	MergeableApplyRequirement:  true,
	UnDivergedApplyRequirement: true,
	UnlockedApplyRequirement:   true,
}

var supportedApplyReqs = buildSupportedApplyReqs()

type Project struct {
	Name                    *string           `yaml:"name,omitempty"`
	Dir                     *string           `yaml:"dir,omitempty"`
	Workspace               *string           `yaml:"workspace,omitempty"`
	Workflow                *string           `yaml:"workflow,omitempty"`
	PullRequestWorkflowName *string           `yaml:"pull_request_workflow,omitempty"`
	DeploymentWorkflowName  *string           `yaml:"deployment_workflow,omitempty"`
	TerraformVersion        *string           `yaml:"terraform_version,omitempty"`
	Autoplan                *Autoplan         `yaml:"autoplan,omitempty"`
	ApplyRequirements       []string          `yaml:"apply_requirements,omitempty"`
	Tags                    map[string]string `yaml:"tags,omitempty"`
	WorkflowModeType        *string           `yaml:"workflow_mode_type,omitempty"`
}

func (p Project) Validate() error {
	hasDotDot := func(value interface{}) error {
		if strings.Contains(*value.(*string), "..") {
			return errors.New("cannot contain '..'")
		}
		return nil
	}

	validName := func(value interface{}) error {
		strPtr := value.(*string)
		if strPtr == nil {
			return nil
		}
		if *strPtr == "" {
			return errors.New("if set cannot be empty")
		}
		if !validProjectName(*strPtr) {
			return fmt.Errorf("%q is not allowed: must contain only URL safe characters", *strPtr)
		}
		return nil
	}
	return validation.ValidateStruct(&p,
		validation.Field(&p.Dir, validation.Required, validation.By(hasDotDot)),
		validation.Field(&p.ApplyRequirements, validation.By(validApplyReq)),
		validation.Field(&p.TerraformVersion, validation.By(VersionValidator)),
		validation.Field(&p.Name, validation.By(validName)),
	)
}

func (p Project) ToValid() valid.Project {
	var v valid.Project
	// Prepend ./ and then run .Clean() so we're guaranteed to have a relative
	// directory. This is necessary because we use this dir without sanitation
	// in DefaultProjectFinder.
	cleanedDir := filepath.Clean("./" + *p.Dir)
	v.Dir = cleanedDir

	if p.Workspace == nil || *p.Workspace == "" {
		v.Workspace = DefaultWorkspace
	} else {
		v.Workspace = *p.Workspace
	}

	v.WorkflowName = p.Workflow
	v.PullRequestWorkflowName = p.PullRequestWorkflowName
	v.DeploymentWorkflowName = p.DeploymentWorkflowName
	if p.WorkflowModeType != nil {
		v.WorkflowModeType = p.WorkflowModeType
	}
	if p.TerraformVersion != nil {
		v.TerraformVersion, _ = version.NewVersion(*p.TerraformVersion)
	}
	if p.Autoplan == nil {
		v.Autoplan = DefaultAutoPlan()
	} else {
		v.Autoplan = p.Autoplan.ToValid()
	}

	// There are no default apply requirements.
	v.ApplyRequirements = p.ApplyRequirements

	v.Tags = p.Tags
	v.Name = p.Name

	return v
}

// validProjectName returns true if the project name is valid.
// Since the name might be used in URLs and definitely in files we don't
// support any characters that must be url escaped *except* for '/' because
// users like to name their projects to match the directory it's in.
func validProjectName(name string) bool {
	nameWithoutSlashes := strings.Replace(name, "/", "-", -1)
	return nameWithoutSlashes == url.QueryEscape(nameWithoutSlashes)
}

func validApplyReq(value interface{}) error {
	applyRequirementsList := make([]string, 0, len(applyRequirements))
	for applyReq := range applyRequirements {
		applyRequirementsList = append(applyRequirementsList, applyReq)
	}
	sort.Strings(applyRequirementsList)

	reqs := value.([]string)
	for _, req := range reqs {
		if _, ok := applyRequirements[req]; !ok {
			return fmt.Errorf("%q is not a valid apply_requirement, supported apply requirements are: %s", req, supportedApplyReqs)
		}
	}
	return nil
}

func buildSupportedApplyReqs() string {
	// Sort keys.
	applyRequirementsList := make([]string, 0, len(applyRequirements))
	for applyReq := range applyRequirements {
		applyRequirementsList = append(applyRequirementsList, applyReq)
	}
	sort.Strings(applyRequirementsList)

	return "\"" + strings.Join(applyRequirementsList, "\", \"") + "\""
}
