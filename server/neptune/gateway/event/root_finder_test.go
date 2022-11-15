// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.

package event_test

import (
	"context"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/neptune/gateway/event"
	. "github.com/runatlantis/atlantis/testing"
)

func TestDefaultProjectFinder_FindRoots(t *testing.T) {
	// Create dir structure:
	// main.tf
	// project1/
	//   main.tf
	//   terraform.tfvars.json
	// project2/
	//   main.tf
	//   terraform.tfvars
	// modules/
	//   module/
	//	  main.tf
	_, cleanup := DirStructure(t, map[string]interface{}{
		"main.tf": nil,
		"project1": map[string]interface{}{
			"main.tf":               nil,
			"terraform.tfvars.json": nil,
		},
		"project2": map[string]interface{}{
			"main.tf":          nil,
			"terraform.tfvars": nil,
		},
		"modules": map[string]interface{}{
			"module": map[string]interface{}{
				"main.tf": nil,
			},
		},
	})
	defer cleanup()

	cases := []struct {
		description  string
		config       valid.RepoCfg
		modified     []string
		expProjPaths []string
		skipMkDir    bool
	}{
		{
			// When autoplan is disabled, we still return the modified project.
			// If our caller is interested in autoplan enabled projects, they'll
			// need to filter the results.
			description: "autoplan disabled",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: ".",
						Autoplan: valid.Autoplan{
							Enabled:      false,
							WhenModified: []string{"**/*.tf"},
						},
					},
				},
			},
			modified:     []string{"main.tf"},
			expProjPaths: []string{"."},
		},
		{
			description: "autoplan default",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: ".",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf"},
						},
					},
				},
			},
			modified:     []string{"main.tf"},
			expProjPaths: []string{"."},
		},
		{
			description: "parent dir modified",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf"},
						},
					},
				},
			},
			modified:     []string{"main.tf"},
			expProjPaths: nil,
		},
		{
			description: "parent dir modified matches",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project1",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"../**/*.tf"},
						},
					},
				},
			},
			modified:     []string{"main.tf"},
			expProjPaths: []string{"project1"},
		},
		{
			description: "multiple projects",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: ".",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"*.tf"},
						},
					},
					{
						Dir: "project1",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"../modules/module/*.tf", "**/*.tf"},
						},
					},
					{
						Dir: "project2",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf"},
						},
					},
				},
			},
			modified:     []string{"main.tf", "modules/module/another.tf", "project2/nontf.txt"},
			expProjPaths: []string{".", "project1"},
		},
		{
			description: ".tfvars file modified",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project2",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"*.tf*"},
						},
					},
				},
			},
			modified:     []string{"project2/terraform.tfvars"},
			expProjPaths: []string{"project2"},
		},
		{
			description: "file excluded",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project1",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"*.tf", "!exclude-me.tf"},
						},
					},
				},
			},
			modified:     []string{"project1/exclude-me.tf"},
			expProjPaths: nil,
		},
		{
			description: "some files excluded and others included",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project1",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"*.tf", "!exclude-me.tf"},
						},
					},
				},
			},
			modified:     []string{"project1/exclude-me.tf", "project1/include-me.tf"},
			expProjPaths: []string{"project1"},
		},
		{
			description: "multiple dirs excluded",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project1",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"**/*.tf", "!subdir1/*", "!subdir2/*"},
						},
					},
				},
			},
			modified:     []string{"project1/subdir1/main.tf", "project1/subdir2/main.tf"},
			expProjPaths: nil,
		},
		{
			description: "skip temp dir",
			config: valid.RepoCfg{
				Projects: []valid.Project{
					{
						Dir: "project2",
						Autoplan: valid.Autoplan{
							Enabled:      true,
							WhenModified: []string{"*.tf*"},
						},
					},
				},
			},
			modified:     []string{"project2/terraform.tfvars"},
			expProjPaths: nil,
			skipMkDir:    true,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			rf := event.RepoRootFinder{
				Logger: logging.NewNoopCtxLogger(t),
			}
			tempDir := t.TempDir()
			if !c.skipMkDir {
				for _, proj := range c.config.Projects {
					assert.NoError(t, os.MkdirAll(filepath.Join(tempDir, proj.Dir), 0700))
				}
			}
			projects, err := rf.FindRoots(context.Background(), c.config, tempDir, c.modified)
			assert.NoError(t, err)
			assert.Equal(t, len(c.expProjPaths), len(projects))
			for i, proj := range projects {
				assert.Equal(t, c.expProjPaths[i], proj.Dir)
			}
		})
	}
}
