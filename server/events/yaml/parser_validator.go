package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"gopkg.in/yaml.v2"
)

// AtlantisYAMLFilename is the name of the config file for each repo.
const AtlantisYAMLFilename = "atlantis.yaml"

type ParserValidator struct{}

func (p *ParserValidator) HasRepoCfg(repoDir string) (bool, error) {
	_, err := os.Stat(p.repoCfgPath(repoDir))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

// ParseRepoCfg returns the parsed and validated atlantis.yaml config for repoDir.
// If there was no config file, then this can be detected by checking the type
// of error: os.IsNotExist(error) but it's instead preferred to check with
// HasRepoCfg.
func (p *ParserValidator) ParseRepoCfg(repoDir string, globalCfg valid.GlobalCfg, repoID string) (valid.Config, error) {
	configFile := p.repoCfgPath(repoDir)
	configData, err := ioutil.ReadFile(configFile) // nolint: gosec

	if err != nil {
		if !os.IsNotExist(err) {
			return valid.Config{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
		}
		// Don't wrap os.IsNotExist errors because we want our callers to be
		// able to detect if it's a NotExist err.
		return valid.Config{}, err
	}

	var rawConfig raw.Config
	if err := yaml.UnmarshalStrict(configData, &rawConfig); err != nil {
		return valid.Config{}, err
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"
	if err := rawConfig.Validate(); err != nil {
		return valid.Config{}, err
	}

	validConfig := rawConfig.ToValid()

	// We do the project name validation after we get the valid config because
	// we need the defaults of dir and workspace to be populated.
	if err := p.validateProjectNames(validConfig); err != nil {
		return valid.Config{}, err
	}
	err = globalCfg.ValidateRepoCfg(validConfig, repoID)
	return validConfig, err
}

func (p *ParserValidator) ParseGlobalCfg(configFile string, defaultCfg valid.GlobalCfg) (valid.GlobalCfg, error) {
	configData, err := ioutil.ReadFile(configFile) // nolint: gosec
	if err != nil {
		return valid.GlobalCfg{}, errors.Wrapf(err, "unable to read %s file", configFile)
	}
	if len(configData) == 0 {
		return valid.GlobalCfg{}, fmt.Errorf("file %s was empty", configFile)
	}

	var rawCfg raw.GlobalCfg
	if err := yaml.UnmarshalStrict(configData, &rawCfg); err != nil {
		return valid.GlobalCfg{}, err
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"
	if err := rawCfg.Validate(); err != nil {
		return valid.GlobalCfg{}, err
	}

	validCfg := rawCfg.ToValid()

	// Add defaults to the parsed config.
	validCfg.Repos = append(defaultCfg.Repos, validCfg.Repos...)
	for k, v := range defaultCfg.Workflows {
		// We won't override existing workflows.
		if _, ok := validCfg.Workflows[k]; !ok {
			validCfg.Workflows[k] = v
		}
	}
	return validCfg, nil
}

func (p *ParserValidator) repoCfgPath(repoDir string) string {
	return filepath.Join(repoDir, AtlantisYAMLFilename)
}

func (p *ParserValidator) validateProjectNames(config valid.Config) error {
	// First, validate that all names are unique.
	seen := make(map[string]bool)
	for _, project := range config.Projects {
		if project.Name != nil {
			name := *project.Name
			exists := seen[name]
			if exists {
				return fmt.Errorf("found two or more projects with name %q; project names must be unique", name)
			}
			seen[name] = true
		}
	}

	// Next, validate that all dir/workspace combos are named.
	// This map's keys will be 'dir/workspace' and the values are the names for
	// that project.
	dirWorkspaceToNames := make(map[string][]string)
	for _, project := range config.Projects {
		key := fmt.Sprintf("%s/%s", project.Dir, project.Workspace)
		names := dirWorkspaceToNames[key]

		// If there is already a project with this dir/workspace then this
		// project must have a name.
		if len(names) > 0 && project.Name == nil {
			return fmt.Errorf("there are two or more projects with dir: %q workspace: %q that are not all named; they must have a 'name' key so they can be targeted for apply's separately", project.Dir, project.Workspace)
		}
		var name string
		if project.Name != nil {
			name = *project.Name
		}
		dirWorkspaceToNames[key] = append(dirWorkspaceToNames[key], name)
	}

	return nil
}
