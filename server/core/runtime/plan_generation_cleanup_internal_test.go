// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import (
	"os"
	"testing"

	"github.com/hashicorp/go-version"
	. "github.com/petergtz/pegomock/v4"
	tf "github.com/runatlantis/atlantis/server/core/terraform"
	tfclientmocks "github.com/runatlantis/atlantis/server/core/terraform/tfclient/mocks"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestTerminalRemediation_ReapsGenerationWithoutCanonicalCache(t *testing.T) {
	for _, name := range []string{"import", "state rm"} {
		t.Run(name, func(t *testing.T) {
			RegisterMockTestingT(t)
			dir := t.TempDir()
			ctx := command.ProjectContext{Log: logging.NewNoopLogger(t), Workspace: "default", PlanGeneration: "G1", SavedPlanHash: new(string)}
			store := &LocalPlanStore{}
			path := GetPlanFilePath(ctx, dir)
			Ok(t, os.WriteFile(path, []byte("accepted plan"), 0600))
			Ok(t, store.Save(ctx, path))
			ctx.AcceptedPlanGeneration, ctx.ExpectedPlanHash = "G1", *ctx.SavedPlanHash
			Ok(t, os.Remove(path))
			terraform := tfclientmocks.NewMockClient()
			distribution := tf.NewDistribution("terraform")
			tfVersion := version.Must(version.NewVersion("1.11.1"))
			runner := NewImportStepRunner(terraform, distribution, tfVersion, store)
			if name == "state rm" {
				runner = NewStateRmStepRunner(terraform, distribution, tfVersion, store)
			}
			When(terraform.RunCommandWithVersion(Any[command.ProjectContext](), Any[string](), Any[[]string](), Any[map[string]string](), Any[tf.Distribution](), Any[*version.Version](), Any[string]())).ThenReturn("completed", nil)
			output, err := runner.Run(ctx, nil, dir, nil)
			Ok(t, err)
			Equals(t, "completed", output)
			Assert(t, store.Load(ctx, path) != nil, "successful remediation must reap the accepted artifact even without a canonical cache")
		})
	}
}
