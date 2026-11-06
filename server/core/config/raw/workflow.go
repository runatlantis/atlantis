// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Workflow struct {
	Apply       *Stage `yaml:"apply,omitempty" json:"apply,omitempty"`
	Plan        *Stage `yaml:"plan,omitempty" json:"plan,omitempty"`
	PolicyCheck *Stage `yaml:"policy_check,omitempty" json:"policy_check,omitempty"`
	Import      *Stage `yaml:"import,omitempty" json:"import,omitempty"`
	Output      *Stage `yaml:"output,omitempty" json:"output,omitempty"`
	StateList   *Stage `yaml:"state_list,omitempty" json:"state_list,omitempty"`
	StateShow   *Stage `yaml:"state_show,omitempty" json:"state_show,omitempty"`
	StateRm     *Stage `yaml:"state_rm,omitempty" json:"state_rm,omitempty"`
}

func (w Workflow) Validate() error {
	return validation.ValidateStruct(&w,
		validation.Field(&w.Apply),
		validation.Field(&w.Plan),
		validation.Field(&w.PolicyCheck),
		validation.Field(&w.Import),
		validation.Field(&w.Output),
		validation.Field(&w.StateList),
		validation.Field(&w.StateShow),
		validation.Field(&w.StateRm),
	)
}

func (w Workflow) toValidStage(stage *Stage, defaultStage valid.Stage) valid.Stage {
	if stage == nil || stage.Steps == nil {
		return defaultStage
	}

	return stage.ToValid()
}

func (w Workflow) ToValid(name string) valid.Workflow {
	v := valid.Workflow{
		Name: name,
	}

	v.Apply = w.toValidStage(w.Apply, valid.DefaultApplyStage)
	v.Plan = w.toValidStage(w.Plan, valid.DefaultPlanStage)
	v.PolicyCheck = w.toValidStage(w.PolicyCheck, valid.DefaultPolicyCheckStage)
	v.Import = w.toValidStage(w.Import, valid.DefaultImportStage)
	v.Output = w.toValidStage(w.Output, valid.DefaultOutputStage)
	v.StateList = w.toValidStage(w.StateList, valid.DefaultStateListStage)
	v.StateShow = w.toValidStage(w.StateShow, valid.DefaultStateShowStage)
	v.StateRm = w.toValidStage(w.StateRm, valid.DefaultStateRmStage)

	return v
}
