// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
)

func writeRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range files {
		path := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("creating dir for %s: %s", name, err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatalf("writing %s: %s", name, err)
		}
	}
	return dir
}

func projectCfg(dir, name string, dependsOn ...string) valid.MergedProjectCfg {
	return valid.MergedProjectCfg{
		RepoRelDir:      dir,
		Workspace:       DefaultWorkspace,
		Name:            name,
		AutoplanEnabled: true,
		DependsOn:       dependsOn,
	}
}

func dependentDirs(t *testing.T, index *OutputDependencyIndex, producerDir string, changed map[string]bool) []string {
	t.Helper()
	keys := index.Dependents([]OutputDependentTrigger{{
		RepoRelDir:     producerDir,
		Workspace:      DefaultWorkspace,
		ChangedOutputs: changed,
	}}, nil)
	dirs := make([]string, 0, len(keys))
	for _, key := range keys {
		dirs = append(dirs, key.RepoRelDir)
	}
	return dirs
}

func assertDirs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("expected dependents %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected dependents %v, got %v", want, got)
		}
	}
}

func TestBuildOutputDependencyIndex_S3RemoteState(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"envs/prod/network/main.tf": `
terraform {
  backend "s3" {
    bucket = "acme-tf"
    key    = "prod/network/terraform.tfstate"
    region = "us-east-1"
  }
}

output "vpc_id" {
  value = "vpc-123"
}

output "subnet_ids" {
  value = ["subnet-1"]
}
`,
		"envs/prod/app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {
    bucket = "acme-tf"
    key    = "prod/network/terraform.tfstate"
    region = "us-east-1"
  }
}

locals {
  vpc_id = data.terraform_remote_state.network.outputs.vpc_id
}
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("envs/prod/network", "network"),
		projectCfg("envs/prod/app", "app"),
	})

	assertDirs(t, dependentDirs(t, index, "envs/prod/network", map[string]bool{"vpc_id": true}), []string{"envs/prod/app"})
	// The app only reads vpc_id, so a change to another output must not
	// re-plan it.
	assertDirs(t, dependentDirs(t, index, "envs/prod/network", map[string]bool{"subnet_ids": true}), nil)
	// A different bucket is a different state.
	repoDirOtherBucket := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  backend "s3" {
    bucket = "acme-tf"
    key    = "network/terraform.tfstate"
  }
}
output "vpc_id" { value = "vpc-1" }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {
    bucket = "someone-else"
    key    = "network/terraform.tfstate"
  }
}
output "vpc" { value = data.terraform_remote_state.network.outputs.vpc_id }
`,
	})
	otherIndex := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDirOtherBucket, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})
	assertDirs(t, dependentDirs(t, otherIndex, "network", map[string]bool{"vpc_id": true}), nil)
}

func TestBuildOutputDependencyIndex_WholeOutputsObject(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  backend "s3" {
    bucket = "acme-tf"
    key    = "network/terraform.tfstate"
  }
}
output "vpc_id" { value = "vpc-1" }
output "subnet_ids" { value = ["subnet-1"] }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "s3"
  config = {
    bucket = "acme-tf"
    key    = "network/terraform.tfstate"
  }
}

output "remote" {
  value = data.terraform_remote_state.network.outputs
}
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})

	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"anything": true}), []string{"app"})
}

func TestBuildOutputDependencyIndex_LocalBackendPath(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"stacks/base/main.tf": `
output "id" { value = "base" }
`,
		"stacks/app/main.tf": `
data "terraform_remote_state" "base" {
  backend = "local"
  config = {
    path = "../base/terraform.tfstate"
  }
}

output "id" { value = data.terraform_remote_state.base.outputs.id }
`,
		"stacks/other/main.tf": `
data "terraform_remote_state" "base" {
  backend = "local"
  config = {
    path = "../../outside/terraform.tfstate"
  }
}

output "id" { value = data.terraform_remote_state.base.outputs.id }
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("stacks/base", "base"),
		projectCfg("stacks/app", "app"),
		projectCfg("stacks/other", "other"),
	})

	assertDirs(t, dependentDirs(t, index, "stacks/base", map[string]bool{"id": true}), []string{"stacks/app"})
}

func TestBuildOutputDependencyIndex_GCSDefaultPrefix(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  backend "gcs" {
    bucket = "acme-tf"
    prefix = "terraform/state"
  }
}
output "id" { value = "x" }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "gcs"
  config = {
    bucket = "acme-tf"
  }
}
output "id" { value = data.terraform_remote_state.network.outputs.id }
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})
	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"id": true}), []string{"app"})
}

func TestBuildOutputDependencyIndex_UnresolvableConfigIsIgnored(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  backend "s3" {
    bucket = "acme-tf"
    key    = "network/terraform.tfstate"
  }
}
output "vpc_id" { value = "vpc-1" }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "s3"
  config = local.remote_state_config
}
output "vpc" { value = data.terraform_remote_state.network.outputs.vpc_id }
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})
	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"vpc_id": true}), nil)
}

func TestBuildOutputDependencyIndex_TerraformCloudWorkspace(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  cloud {
    organization = "acme"

    workspaces {
      name = "prod-network"
    }
  }
}
output "vpc_id" { value = "vpc-1" }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "remote"
  config = {
    organization = "acme"
    workspaces = {
      name = "prod-network"
    }
  }
}
output "vpc" { value = data.terraform_remote_state.network.outputs.vpc_id }
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})
	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"vpc_id": true}), []string{"app"})
}

func TestBuildOutputDependencyIndex_UnresolvableProducerBackendIsNotTreatedAsLocal(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `
terraform {
  backend "s3" {
    bucket = "acme-tf"
    key    = "network/${terraform.workspace}/terraform.tfstate"
  }
}
output "vpc_id" { value = "vpc-1" }
`,
		"app/main.tf": `
data "terraform_remote_state" "network" {
  backend = "local"
  config = {
    path = "../network/terraform.tfstate"
  }
}
output "vpc" { value = data.terraform_remote_state.network.outputs.vpc_id }
`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app"),
	})
	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"vpc_id": true}), nil)
}

func TestBuildOutputDependencyIndex_DependsOn(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `output "vpc_id" { value = "vpc-1" }`,
		"app/main.tf":     `output "vpc" { value = "unrelated" }`,
	})

	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app", "network"),
	})
	// depends_on means the dependent may consume any output, so it is planned
	// for any change.
	assertDirs(t, dependentDirs(t, index, "network", map[string]bool{"anything": true}), []string{"app"})
}

func TestOutputDependencyIndex_DependentsExcludesPlannedProjects(t *testing.T) {
	repoDir := writeRepo(t, map[string]string{
		"network/main.tf": `output "vpc_id" { value = "vpc-1" }`,
		"app/main.tf":     `output "vpc" { value = "unrelated" }`,
		"web/main.tf":     `output "vpc" { value = "unrelated" }`,
	})
	index := BuildOutputDependencyIndex(logging.NewNoopLogger(t), repoDir, []valid.MergedProjectCfg{
		projectCfg("network", "network"),
		projectCfg("app", "app", "network"),
		projectCfg("web", "web", "network"),
	})

	keys := index.Dependents([]OutputDependentTrigger{{
		RepoRelDir:     "network",
		Workspace:      DefaultWorkspace,
		ChangedOutputs: map[string]bool{"vpc_id": true},
	}}, map[ProjectKey]bool{
		NewProjectKey("app", DefaultWorkspace, "app"): true,
	})
	if len(keys) != 1 || keys[0].RepoRelDir != "web" {
		t.Fatalf("expected only web, got %v", keys)
	}
}
