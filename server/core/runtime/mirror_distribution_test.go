// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/terraform"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestVersionStepRunner_ProjectOverrideReusesMirrorDistribution(t *testing.T) {
	for _, defaultName := range []string{"terraform", "opentofu"} {
		t.Run(defaultName, func(t *testing.T) {
			RegisterMockTestingT(t)
			configured := terraform.NewDistribution(defaultName, "https://mirror.example", terraform.APIAuth{BearerToken: "token"})
			projectName := "terraform"
			expected := terraform.SelectDistribution(configured, &projectName)
			v, err := version.NewVersion("1.9.3")
			Ok(t, err)
			executor := tfclientmocks.NewMockClient()
			ctx := command.ProjectContext{
				Log: logging.NewNoopLogger(t), Workspace: "default", TerraformDistribution: &projectName,
			}
			When(executor.RunCommandWithVersion(ctx, "/project", []string{"version"}, nil, expected, v, "default")).ThenReturn("version output", nil)
			runner := VersionStepRunner{TerraformExecutor: executor, DefaultTFDistribution: configured, DefaultTFVersion: v}
			output, err := runner.Run(ctx, nil, "/project", nil)
			Ok(t, err)
			Equals(t, "version output", output)
			executor.VerifyWasCalledOnce().RunCommandWithVersion(ctx, "/project", []string{"version"}, nil, expected, v, "default")
		})
	}
}
