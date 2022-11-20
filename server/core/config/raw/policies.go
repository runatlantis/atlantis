package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// PolicySets is the raw schema for repo-level atlantis.yaml config.
type PolicySets struct {
	Version    *string      `yaml:"conftest_version,omitempty" json:"conftest_version,omitempty"`
	Owners     PolicyOwners `yaml:"owners,omitempty" json:"owners,omitempty"`
	PolicySets []PolicySet  `yaml:"policy_sets" json:"policy_sets"`
}

func (p PolicySets) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Version, validation.By(VersionValidator)),
		validation.Field(&p.PolicySets, validation.Required.Error("cannot be empty; Declare policies that you would like to enforce")),
	)
}

func (p PolicySets) ToValid() valid.PolicySets {
	policySets := valid.PolicySets{}

	if p.Version != nil {
		policySets.Version, _ = version.NewVersion(*p.Version)
	}

	policySets.Owners = p.Owners.ToValid()

	validPolicySets := make([]valid.PolicySet, 0)
	for _, rawPolicySet := range p.PolicySets {
		validPolicySets = append(validPolicySets, rawPolicySet.ToValid())
	}
	policySets.PolicySets = validPolicySets

	return policySets
}

type PolicyOwners struct {
	Users []string `yaml:"users,omitempty" json:"users,omitempty"`
}

func (o PolicyOwners) ToValid() valid.PolicyOwners {
	var policyOwners valid.PolicyOwners

	if len(o.Users) > 0 {
		policyOwners.Users = o.Users
	}
	return policyOwners
}

type PolicySet struct {
	Path   string       `yaml:"path" json:"path"`
	Source string       `yaml:"source" json:"source"`
	Name   string       `yaml:"name" json:"name"`
	Owners PolicyOwners `yaml:"owners,omitempty" json:"owners,omitempty"`
}

func (p PolicySet) Validate() error {
	return validation.ValidateStruct(&p,
		validation.Field(&p.Name, validation.Required.Error("is required")),
		validation.Field(&p.Owners),
		validation.Field(&p.Path, validation.Required.Error("is required")),
		validation.Field(&p.Source, validation.In(valid.LocalPolicySet, valid.GithubPolicySet).Error("only 'local' and 'github' source types are supported")),
	)
}

func (p PolicySet) ToValid() valid.PolicySet {
	var policySet valid.PolicySet

	policySet.Name = p.Name
	policySet.Path = p.Path
	policySet.Source = p.Source
	policySet.Owners = p.Owners.ToValid()

	return policySet
}
