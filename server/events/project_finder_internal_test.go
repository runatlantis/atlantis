// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"io/fs"
	"testing"
	"testing/fstest"

	. "github.com/runatlantis/atlantis/testing"
)

func TestGetProjectDirFromFs_ModuleParentWithMainTf(t *testing.T) {
	files := fstest.MapFS{
		"project/main.tf":             &fstest.MapFile{},
		"project/modules/foo/main.tf": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tf")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentWithMainTofu(t *testing.T) {
	files := fstest.MapFS{
		"project/main.tofu":             &fstest.MapFile{},
		"project/modules/foo/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentNoMainTfOrTofu(t *testing.T) {
	files := fstest.MapFS{
		"project/modules/foo/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu")
	Equals(t, "", result)
}

func TestGetProjectDirFromFs_NonModuleTofu(t *testing.T) {
	files := fstest.MapFS{
		"project/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/main.tofu")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentWithMainTofuJSON(t *testing.T) {
	files := fstest.MapFS{
		"project/main.tofu.json":             &fstest.MapFile{},
		"project/modules/foo/main.tofu.json": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu.json")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentTofuJSON_ChangedTofu(t *testing.T) {
	files := fstest.MapFS{
		"project/main.tofu.json":        &fstest.MapFile{},
		"project/modules/foo/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentVersionsTofu(t *testing.T) {
	files := fstest.MapFS{
		"project/versions.tofu":         &fstest.MapFile{},
		"project/modules/foo/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentVersionsTofuJSON(t *testing.T) {
	files := fstest.MapFS{
		"project/versions.tofu.json":         &fstest.MapFile{},
		"project/modules/foo/main.tofu.json": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu.json")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentNoIndicator(t *testing.T) {
	files := fstest.MapFS{
		"project/README.md":             &fstest.MapFile{},
		"project/modules/foo/main.tofu": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tofu")
	Equals(t, "", result)
}

func TestGetProjectDirFromFs_ModuleParentTfJSON(t *testing.T) {
	files := fstest.MapFS{
		"project/versions.tf.json":    &fstest.MapFile{},
		"project/modules/foo/main.tf": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tf")
	Equals(t, "project", result)
}

func TestGetProjectDirFromFs_ModuleParentTerragrunt(t *testing.T) {
	files := fstest.MapFS{
		"project/terragrunt.hcl":      &fstest.MapFile{},
		"project/modules/foo/main.tf": &fstest.MapFile{},
	}
	result := getProjectDirFromFs(fs.FS(files), "project/modules/foo/main.tf")
	Equals(t, "project", result)
}
