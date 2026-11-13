// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package raw

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/runatlantis/atlantis/server/core/config/valid"
)

type Workflow struct {
	Apply       *Stage `yaml:"apply,omitempty" json:"apply,omitempty"`
	Plan        *Stage `yaml:"plan,omitempty" json:"plan,omitempty"`
	PolicyCheck *Stage `yaml:"policy_check,omitempty" json:"policy_check,omitempty"`
	Import      *Stage `yaml:"import,omitempty" json:"import,omitempty"`
	StateRm     *Stage `yaml:"state_rm,omitempty" json:"state_rm,omitempty"`
}

func (w Workflow) Validate() error {
	if err := validation.ValidateStruct(&w,
		validation.Field(&w.Apply),
		validation.Field(&w.Plan),
		validation.Field(&w.PolicyCheck),
		validation.Field(&w.Import),
		validation.Field(&w.StateRm),
	); err != nil {
		return err
	}

	stages := []struct {
		name  string
		stage *Stage
	}{
		{name: "plan", stage: w.Plan},
		{name: "apply", stage: w.Apply},
		{name: "policy_check", stage: w.PolicyCheck},
		{name: "import", stage: w.Import},
		{name: "state_rm", stage: w.StateRm},
	}
	for _, candidate := range stages {
		if candidate.stage == nil {
			continue
		}
		for _, step := range candidate.stage.Steps {
			mode := step.planStoreMode()
			if mode == string(valid.RunPlanStoreSaveMode) && candidate.name != "plan" {
				return fmt.Errorf("%s: run step plan_store mode %q is only valid in the plan stage", candidate.name, mode)
			}
			if mode == string(valid.RunPlanStoreConsumeMode) && candidate.name != "apply" {
				return fmt.Errorf("%s: run step plan_store mode %q is only valid in the apply stage", candidate.name, mode)
			}
		}
	}

	if w.Plan != nil && w.Plan.countPlanStoreMode(string(valid.RunPlanStoreSaveMode)) > 1 {
		return fmt.Errorf("plan: run step plan_store mode %q may only be configured once", valid.RunPlanStoreSaveMode)
	}
	if w.Apply != nil && w.Apply.countPlanStoreMode(string(valid.RunPlanStoreConsumeMode)) > 1 {
		return fmt.Errorf("apply: run step plan_store mode %q may only be configured once", valid.RunPlanStoreConsumeMode)
	}
	if w.Plan != nil && w.Plan.countPlanStoreMode(string(valid.RunPlanStoreSaveMode)) > 0 && w.Plan.hasStep(PlanStepName) {
		return fmt.Errorf("plan: run step plan_store mode %q cannot be combined with the built-in %q step", valid.RunPlanStoreSaveMode, PlanStepName)
	}
	if w.Apply != nil && w.Apply.countPlanStoreMode(string(valid.RunPlanStoreConsumeMode)) > 0 && w.Apply.hasStep(ApplyStepName) {
		return fmt.Errorf("apply: run step plan_store mode %q cannot be combined with the built-in %q step", valid.RunPlanStoreConsumeMode, ApplyStepName)
	}

	return nil
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
	v.StateRm = w.toValidStage(w.StateRm, valid.DefaultStateRmStage)

	return v
}
