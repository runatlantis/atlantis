package event

import (
	"fmt"
	"os"
	"path/filepath"

	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/raw"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"gopkg.in/yaml.v2"
)

// AtlantisYAMLFilename is the name of the config file for each repo.
const AtlantisYAMLFilename = "atlantis.yaml"

// ParserValidator parses and validates server-side repo config files and
// repo-level atlantis.yaml files.
type ParserValidator struct {
	GlobalCfg valid.GlobalCfg
}

// ParseRepoCfg returns the parsed and validated atlantis.yaml config for the
// repo at absRepoDir. If there was no config file, it will return an os.IsNotExist(error).
func (p *ParserValidator) ParseRepoCfg(absRepoDir string, repoID string) (valid.RepoCfg, error) {
	hasRepoCfg, err := p.hasRepoCfg(absRepoDir)
	if err != nil {
		return valid.RepoCfg{}, errors.Wrap(err, "determining if repo cfg exists")
	}
	if !hasRepoCfg {
		return valid.RepoCfg{}, os.ErrNotExist
	}
	configFile := p.repoCfgPath(absRepoDir, AtlantisYAMLFilename)
	configData, err := os.ReadFile(configFile) // nolint: gosec
	if err != nil {
		return valid.RepoCfg{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)

	}
	return p.parseRepoCfgData(configData, repoID)
}

// hasRepoCfg returns true if there is a repo config (atlantis.yaml) file for the repo at absRepoDir.
// Returns an error if for some reason it can't read that directory.
func (p *ParserValidator) hasRepoCfg(absRepoDir string) (bool, error) {
	// Checks for a config file with an invalid extension (atlantis.yml)
	const invalidExtensionFilename = "atlantis.yml"
	_, err := os.Stat(p.repoCfgPath(absRepoDir, invalidExtensionFilename))
	if err == nil {
		return false, errors.Errorf("found %q as config file; rename using the .yaml extension - %q", invalidExtensionFilename, AtlantisYAMLFilename)
	}
	_, err = os.Stat(p.repoCfgPath(absRepoDir, AtlantisYAMLFilename))
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func (p *ParserValidator) parseRepoCfgData(repoCfgData []byte, repoID string) (valid.RepoCfg, error) {
	var rawConfig raw.RepoCfg
	if err := yaml.UnmarshalStrict(repoCfgData, &rawConfig); err != nil {
		return valid.RepoCfg{}, errors.Wrap(err, "unmarshalling repo cfg yaml")
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"
	if err := rawConfig.Validate(); err != nil {
		return valid.RepoCfg{}, errors.Wrap(err, "validating raw config")
	}

	validConfig := rawConfig.ToValid()
	// We do the project name validation after we get the valid config because
	// we need the defaults of dir and workspace to be populated.
	if err := p.validateProjectNames(validConfig); err != nil {
		return valid.RepoCfg{}, errors.Wrap(err, "validating project names")
	}
	err := p.GlobalCfg.ValidateRepoCfg(validConfig, repoID)
	return validConfig, errors.Wrap(err, "validating repo cfg")
}

func (p *ParserValidator) repoCfgPath(repoDir, cfgFilename string) string {
	return filepath.Join(repoDir, cfgFilename)
}

// TODO: rename to root
func (p *ParserValidator) validateProjectNames(config valid.RepoCfg) error {
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
