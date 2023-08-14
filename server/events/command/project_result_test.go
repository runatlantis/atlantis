package command_test

import (
	"errors"
	"testing"

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestProjectResult_IsSuccessful(t *testing.T) {
	cases := map[string]struct {
		pr  command.ProjectResult
		exp bool
	}{
		"plan success": {
			command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{},
			},
			true,
		},
		"policy_check success": {
			command.ProjectResult{
				PolicyCheckResults: &models.PolicyCheckResults{},
			},
			true,
		},
		"apply success": {
			command.ProjectResult{
				ApplySuccess: "success",
			},
			true,
		},
		"failure": {
			command.ProjectResult{
				Failure: "failure",
			},
			false,
		},
		"error": {
			command.ProjectResult{
				Error: errors.New("error"),
			},
			false,
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			Equals(t, c.exp, c.pr.IsSuccessful())
		})
	}
}

func TestProjectResult_PlanStatus(t *testing.T) {
	cases := []struct {
		p         command.ProjectResult
		expStatus models.ProjectPlanStatus
	}{
		{
			p: command.ProjectResult{
				Command: command.Plan,
				Error:   errors.New("err"),
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Plan,
				Failure: "failure",
			},
			expStatus: models.ErroredPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command:     command.Plan,
				PlanSuccess: &models.PlanSuccess{},
			},
			expStatus: models.PlannedPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Plan,
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: "No changes. Infrastructure is up-to-date.",
				},
			},
			expStatus: models.PlannedNoChangesPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Apply,
				Error:   errors.New("err"),
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.Apply,
				Failure: "failure",
			},
			expStatus: models.ErroredApplyStatus,
		},
		{
			p: command.ProjectResult{
				Command:      command.Apply,
				ApplySuccess: "success",
			},
			expStatus: models.AppliedPlanStatus,
		},
		{
			p: command.ProjectResult{
				Command:            command.PolicyCheck,
				PolicyCheckResults: &models.PolicyCheckResults{},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.PolicyCheck,
				Failure: "failure",
			},
			expStatus: models.ErroredPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command:            command.ApprovePolicies,
				PolicyCheckResults: &models.PolicyCheckResults{},
			},
			expStatus: models.PassedPolicyCheckStatus,
		},
		{
			p: command.ProjectResult{
				Command: command.ApprovePolicies,
				Failure: "failure",
			},
			expStatus: models.ErroredPolicyCheckStatus,
		},
	}

	for _, c := range cases {
		t.Run(c.expStatus.String(), func(t *testing.T) {
			Equals(t, c.expStatus, c.p.PlanStatus())
		})
	}
}

func TestPlanSuccess_Summary(t *testing.T) {
	cases := []struct {
		p         command.ProjectResult
		expResult string
	}{
		{
			p: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
				},
			},
			expResult: "Plan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:

					No changes. Infrastructure is up-to-date.`,
				},
			},
			expResult: "No changes. Infrastructure is up-to-date.",
		},
		{
			p: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					No changes. Your infrastructure matches the configuration.`,
				},
			},
			expResult: "\n**Note: Objects have changed outside of Terraform**\nNo changes. Your infrastructure matches the configuration.",
		},
		{
			p: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
				},
			},
			expResult: "\n**Note: Objects have changed outside of Terraform**\nPlan: 0 to add, 0 to change, 1 to destroy.",
		},
		{
			p: command.ProjectResult{
				PlanSuccess: &models.PlanSuccess{
					TerraformOutput: `No match, expect empty`,
				},
			},
			expResult: "",
		},
	}

	for _, c := range cases {
		t.Run(c.expResult, func(t *testing.T) {
			Equals(t, c.expResult, c.p.PlanSuccess.Summary())
		})
	}
}

var Summary string

func BenchmarkPlanSuccess_Summary(b *testing.B) {
	var s string

	fixtures := map[string]string{
		"changes": `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
		"no changes": `
					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:

					No changes. Infrastructure is up-to-date.`,
		"changes outside Terraform": `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					No changes. Your infrastructure matches the configuration.`,
		"changes and changes outside": `
					Note: Objects have changed outside of Terraform

					Terraform detected the following changes made outside of Terraform since the
					last "terraform apply":

					An execution plan has been generated and is shown below.
					Resource actions are indicated with the following symbols:
					  - destroy

					Terraform will perform the following actions:

					  - null_resource.hi[1]


					Plan: 0 to add, 0 to change, 1 to destroy.`,
		"empty summary, no matches": `No match, expect empty`,
	}

	for name, output := range fixtures {
		p := &models.PlanSuccess{
			TerraformOutput: output,
		}

		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				s = p.Summary()
			}

			Summary = s
		})
	}
}
