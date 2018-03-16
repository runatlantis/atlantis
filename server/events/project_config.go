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
//
package events

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// ProjectConfigFile is the filename of Atlantis project config.
const ProjectConfigFile = "atlantis.yaml"

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_project_config_reader.go ProjectConfigReader

// ProjectConfigReader implements reading project config.
type ProjectConfigReader interface {
	// Exists returns true if a project config file exists at projectPath.
	Exists(projectPath string) bool
	// Read attempts to read the project config file for the project at projectPath.
	// NOTE: projectPath is not the path to the actual config file, just to the
	// project root.
	// Returns the parsed ProjectConfig or error if unable to read.
	Read(projectPath string) (ProjectConfig, error)
}

// Hook represents the commands that can be run at a certain stage.
type Hook struct {
	Commands []string `yaml:"commands"`
}

// projectConfigYAML is used to parse the YAML.
type projectConfigYAML struct {
	PreInit          Hook                    `yaml:"pre_init"`
	PreGet           Hook                    `yaml:"pre_get"`
	PrePlan          Hook                    `yaml:"pre_plan"`
	PostPlan         Hook                    `yaml:"post_plan"`
	PreApply         Hook                    `yaml:"pre_apply"`
	PostApply        Hook                    `yaml:"post_apply"`
	TerraformVersion string                  `yaml:"terraform_version"`
	ExtraArguments   []commandExtraArguments `yaml:"extra_arguments"`
}

// ProjectConfig is a more usable version of projectConfigYAML that we can
// return to our callers. It holds the config for a project.
type ProjectConfig struct {
	// PreInit is a slice of command strings to run prior to terraform init.
	PreInit []string
	// PreGet is a slice of command strings to run prior to terraform get.
	PreGet []string
	// PrePlan is a slice of command strings to run prior to terraform plan.
	PrePlan []string
	// PostPlan is a slice of command strings to run after terraform plan.
	PostPlan []string
	// PreApply is a slice of command strings to run prior to terraform apply.
	PreApply []string
	// PostApply is a slice of command strings to run after terraform apply.
	PostApply []string
	// TerraformVersion is the version specified in the config file or nil
	// if version wasn't specified.
	TerraformVersion *version.Version
	// extraArguments is the extra args that we should tack on to certain
	// terraform commands. It shouldn't be used directly and instead callers
	// should use the GetExtraArguments method on ProjectConfig.
	extraArguments []commandExtraArguments
}

// commandExtraArguments is used to parse the config file. These are the args
// that should be tacked on to certain terraform commands.
type commandExtraArguments struct {
	// Name is the name of the command we should add the args to.
	Name string `yaml:"command_name"`
	// Arguments is the list of args we should append.
	Arguments []string `yaml:"arguments"`
}

// ProjectConfigManager deals with project config files that users can
// use to specify additional behaviour around how Atlantis executes for a project.
type ProjectConfigManager struct{}

// Exists returns true if an atlantis config file exists for the project at
// projectPath. projectPath is an absolute path to the project.
func (c *ProjectConfigManager) Exists(projectPath string) bool {
	_, err := os.Stat(filepath.Join(projectPath, ProjectConfigFile))
	return err == nil
}

// Read attempts to read the project config file for the project at projectPath.
// NOTE: projectPath is not the path to the actual config file.
// Returns the parsed ProjectConfig or error if unable to read.
func (c *ProjectConfigManager) Read(execPath string) (ProjectConfig, error) {
	var pc ProjectConfig
	filename := filepath.Join(execPath, ProjectConfigFile)
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		return pc, errors.Wrapf(err, "reading %s", ProjectConfigFile)
	}
	var pcYaml projectConfigYAML
	if err = yaml.Unmarshal(raw, &pcYaml); err != nil {
		return pc, errors.Wrapf(err, "parsing %s", ProjectConfigFile)
	}

	var v *version.Version
	if pcYaml.TerraformVersion != "" {
		v, err = version.NewVersion(pcYaml.TerraformVersion)
		if err != nil {
			return pc, errors.Wrap(err, "parsing terraform_version")
		}
	}
	return ProjectConfig{
		TerraformVersion: v,
		extraArguments:   pcYaml.ExtraArguments,
		PreInit:          pcYaml.PreInit.Commands,
		PreGet:           pcYaml.PreGet.Commands,
		PostApply:        pcYaml.PostApply.Commands,
		PreApply:         pcYaml.PreApply.Commands,
		PrePlan:          pcYaml.PrePlan.Commands,
		PostPlan:         pcYaml.PostPlan.Commands,
	}, nil
}

// GetExtraArguments returns the arguments that were specified to be appended
// to command in the project config file.
func (c *ProjectConfig) GetExtraArguments(command string) []string {
	for _, value := range c.extraArguments {
		if value.Name == command {
			return value.Arguments
		}
	}
	return nil
}
