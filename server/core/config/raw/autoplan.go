// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
)

// terraformAutoplanIndicators are intentionally broad to preserve the default
// autoplan behavior for changes inside existing projects.
var terraformAutoplanIndicators = []string{
	"*.tf*",
	"*.tofu",
	"*.tofu.json",
	"terragrunt.hcl",
	".terraform.lock.hcl",
}

// DefaultAutoPlanWhenModified is the default element in the when_modified
// list if none is defined.
func DefaultAutoPlanWhenModified() []string {
	var ret []string
	for _, indicator := range terraformAutoplanIndicators {
		ret = append(ret, fmt.Sprintf("**/%s", indicator))
	}
	return ret
}

type Autoplan struct {
	WhenModified []string `yaml:"when_modified,omitempty"`
	Enabled      *bool    `yaml:"enabled,omitempty"`
}

func (a Autoplan) ToValid() valid.Autoplan {
	var v valid.Autoplan
	if a.WhenModified == nil {
		v.WhenModified = DefaultAutoPlanWhenModified()
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

// DefaultAutoPlan returns the default autoplan config.
func DefaultAutoPlan() valid.Autoplan {
	return valid.Autoplan{
		WhenModified: DefaultAutoPlanWhenModified(),
		Enabled:      valid.DefaultAutoPlanEnabled,
	}
}
