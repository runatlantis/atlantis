package runtime

import (
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestCleanRemoteOpOutput(t *testing.T) {
	cases := []struct {
		out string
		exp string
	}{
		{
			`
Running apply in the remote backend. Output will stream here. Pressing Ctrl-C
will cancel the remote apply if its still pending. If the apply started it
will stop streaming the logs, but will not stop the apply running remotely.

Preparing the remote apply...

To view this run in a browser, visit:
https://app.terraform.io/app/lkysow-enterprises/atlantis-tfe-test-dir2/runs/run-BCzC79gMDNmGU76T

Waiting for the plan to start...

Terraform v0.11.11

Configuring remote state backend...
Initializing Terraform configuration...
2019/02/27 21:47:23 [DEBUG] Using modified User-Agent: Terraform/0.11.11 TFE/d161c1b
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.

null_resource.dir2[1]: Refreshing state... (ID: 8554368366766418126)
null_resource.dir2: Refreshing state... (ID: 8492616078576984857)

------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  - null_resource.dir2[1]


Plan: 0 to add, 0 to change, 1 to destroy.

Do you want to perform these actions in workspace "atlantis-tfe-test-dir2"?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: 
2019/02/27 21:47:36 [DEBUG] Using modified User-Agent: Terraform/0.11.11 TFE/d161c1b
null_resource.dir2[1]: Destroying... (ID: 8554368366766418126)
null_resource.dir2[1]: Destruction complete after 0s

Apply complete! Resources: 0 added, 0 changed, 1 destroyed.
`,
			`2019/02/27 21:47:36 [DEBUG] Using modified User-Agent: Terraform/0.11.11 TFE/d161c1b
null_resource.dir2[1]: Destroying... (ID: 8554368366766418126)
null_resource.dir2[1]: Destruction complete after 0s

Apply complete! Resources: 0 added, 0 changed, 1 destroyed.
`,
		},
		{
			"nodelim",
			"nodelim",
		},
	}

	for _, c := range cases {
		t.Run(c.exp, func(t *testing.T) {
			a := ApplyStepRunner{}
			Equals(t, c.exp, a.cleanRemoteApplyOutput(c.out))
		})
	}
}

// Test: works normally, sends yes, updates run urls
// Test: if plans don't match, sends no
