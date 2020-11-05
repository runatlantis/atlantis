package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
)

const (
	LocalSourceType  string = "local"
	GithubSourceType string = "github"
)

// Policies is the raw schema for repo-level atlantis.yaml config.
type Policies struct {
	Version    string      `yaml:"conftest_version,omitempty" json:"conftest_version,omitempty"`
	PolicySets []PolicySet `yaml:"policy_sets" json:"policy_sets"`
}

func (p Policies) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Version),
		validation.Field(&p.PolicySets, validation.Required.Error("cannot be empty; Declare policies that you would like to enforce")),
	)
}

func (p Policies) ToValid() valid.Policies {
	v := valid.Policies{
		Version: p.Version,
	}

	validPolicySets := make([]valid.PolicySet, 0)
	for _, rawPolicySet := range p.PolicySets {
		validPolicySets = append(validPolicySets, rawPolicySet.ToValid())
	}
	v.PolicySets = validPolicySets

	return v
}

type PolicySet struct {
	Source PolicySetSource `yaml:"source" json:"source"`
	Name   string          `yaml:"name" json:"name"`
	Owners []string        `yaml:"owners,omitempty" json:"owners,omitempty"`
}

func (p PolicySet) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required.Error("is required")),
		validation.Field(&p.Owners),
		validation.Field(&p.Source, validation.Required.Error("is required")),
	)
}

func (p PolicySet) ToValid() valid.PolicySet {
	var policySet valid.PolicySet

	policySet.Name = p.Name
	policySet.Source = p.Source.ToValid()
	policySet.Owners = p.Owners

	return policySet
}

type PolicySetSource struct {
	Type string `yaml:"type" json:"type"`
	Path string `yaml:"path" json:"path"`
}

func (p PolicySetSource) ToValid() valid.PolicySetSource {
	var policySetSource valid.PolicySetSource

	policySetSource.Path = p.Path
	policySetSource.Type = p.Type

	return policySetSource
}

func (p PolicySetSource) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Path, validation.Required.Error("is required")),
		validation.Field(&p.Type, validation.In(LocalSourceType, GithubSourceType).Error("only 'local' and 'github' source types are supported")),
	)
}
