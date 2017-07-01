package server

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"

	yaml "gopkg.in/yaml.v2"
)

const ProjectConfigFile = "atlantis.yaml"

type PrePlan struct {
	Commands []string `yaml:"commands"`
}

type PreApply struct {
	Commands []string `yaml:"commands"`
}

type ConfigReader struct{}

type ProjectConfigYaml struct {
	PrePlan          PrePlan                 `yaml:"pre_plan"`
	PreApply         PreApply                `yaml:"pre_apply"`
	TerraformVersion string                  `yaml:"terraform_version"`
	ExtraArguments   []CommandExtraArguments `yaml:"extra_arguments"`
}

type ProjectConfig struct {
	PrePlan  PrePlan
	PreApply PreApply
	// TerraformVersion is the version specified in the config file or nil if version wasn't specified
	TerraformVersion *version.Version
	ExtraArguments   []CommandExtraArguments
}

type CommandExtraArguments struct {
	Name      string   `yaml:"command_name"`
	Arguments []string `yaml:"arguments"`
}

func (c *ConfigReader) Exists(execPath string) bool {
	// Check if config file exists
	_, err := os.Stat(filepath.Join(execPath, ProjectConfigFile))
	return err == nil
}

func (c *ConfigReader) Read(execPath string) (ProjectConfig, error) {
	var pc ProjectConfig
	raw, err := ioutil.ReadFile(filepath.Join(execPath, ProjectConfigFile))
	if err != nil {
		return pc, fmt.Errorf("Couldn't read config file %q: %v", execPath, err)
	}
	var pcYaml ProjectConfigYaml
	if err := yaml.Unmarshal(raw, &pcYaml); err != nil {
		return pc, fmt.Errorf("Couldn't decode yaml in config file %q: %v", execPath, err)
	}

	var v *version.Version
	if pcYaml.TerraformVersion != "" {
		v, err = version.NewVersion(pcYaml.TerraformVersion)
		if err != nil {
			return pc, errors.Wrap(err, "parsing terraform_version")
		}
	}
	pc = ProjectConfig{
		TerraformVersion: v,
		ExtraArguments:   pcYaml.ExtraArguments,
		PreApply:         pcYaml.PreApply,
		PrePlan:          pcYaml.PrePlan,
	}

	return pc, nil
}

func (c *ProjectConfig) GetExtraArguments(command string) []string {
	for _, value := range c.ExtraArguments {
		if value.Name == command {
			return value.Arguments
		}
	}
	return nil
}
