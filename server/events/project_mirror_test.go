// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	. "github.com/petergtz/pegomock/v4"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/tfclient"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/mocks"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"github.com/uber-go/tally/v4"
)

func TestProjectTerraformOverridePreservesMirror(t *testing.T) {
	for _, defaultName := range []string{"terraform", "opentofu"} {
		for _, auth := range []struct {
			name   string
			config terraform.APIAuth
			header string
		}{
			{"bearer", terraform.APIAuth{BearerToken: "mirror-token"}, "Bearer mirror-token"},
			{"basic", terraform.APIAuth{Username: "user", Password: "pass"}, "Basic dXNlcjpwYXNz"},
		} {
			for _, policyEnabled := range []bool{false, true} {
				name := "plan"
				if policyEnabled {
					name = "policy"
				}
				t.Run(defaultName+"/"+auth.name+"/"+name, func(t *testing.T) {
					RegisterMockTestingT(t)
					var lookups atomic.Int32
					mirror := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.Header.Get("Authorization") != auth.header {
							http.Error(w, "unauthorized", http.StatusUnauthorized)
							return
						}
						if r.URL.Path != "/terraform/index.json" {
							http.NotFound(w, r)
							return
						}
						lookups.Add(1)
						w.Header().Set("Content-Type", "application/json")
						_, _ = w.Write([]byte(`{"name":"terraform","versions":{"1.9.3":{"name":"terraform","version":"1.9.3","builds":[]}}}`))
					}))
					t.Cleanup(mirror.Close)
					dir := t.TempDir()
					for _, binary := range []string{"terraform", "tofu"} {
						Ok(t, os.WriteFile(filepath.Join(dir, binary), []byte("#!/bin/sh\necho 'Terraform v0.0.1'\n"), 0700)) // #nosec G306 -- executable version stub in a temporary test directory
					}
					t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
					Ok(t, os.WriteFile(filepath.Join(dir, "main.tf"), []byte("terraform { required_version = \"~> 1.9.0\" }\n"), 0600))
					logger := logging.NewNoopLogger(t)
					client, err := tfclient.NewTestClient(logger,
						terraform.NewDistribution(defaultName, mirror.URL, auth.config),
						t.TempDir(), t.TempDir(), "", "", "", "default-tf-version", mirror.URL, true, false, nil)
					Ok(t, err)
					projectDistribution := "terraform"
					project := valid.MergedProjectCfg{
						RepoRelDir: ".", Workspace: "default", TerraformDistribution: &projectDistribution,
						Workflow: valid.Workflow{Plan: valid.DefaultPlanStage},
					}
					builder := events.NewProjectCommandContextBuilder(policyEnabled, mocks.NewMockCommentBuilder(), tally.NoopScope)
					contexts := builder.BuildProjectContext(&command.Context{Log: logger, Scope: tally.NoopScope},
						command.Plan, "", project, nil, dir, false, false, false, false, false, client)
					Equals(t, 1, len(contexts))
					Assert(t, contexts[0].TerraformVersion != nil, "mirror version must be resolved")
					Equals(t, "1.9.3", contexts[0].TerraformVersion.String())
					Equals(t, int32(1), lookups.Load())
				})
			}
		}
	}
}
