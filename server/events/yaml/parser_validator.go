package yaml

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	shlex "github.com/flynn-archive/go-shlex"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/yaml/raw"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	yaml "gopkg.in/yaml.v2"
)

// AtlantisYAMLFilename is the name of the config file for each repo.
const AtlantisYAMLFilename = "atlantis.yaml"

// ParserValidator parses and validates server-side repo config files and
// repo-level atlantis.yaml files.
type ParserValidator struct{}

// HasRepoCfg returns true if there is a repo config (atlantis.yaml) file
// for the repo at absRepoDir.
// Returns an error if for some reason it can't read that directory.
func (p *ParserValidator) HasRepoCfg(absRepoDir string) (bool, error) {
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

// ParseRepoCfg returns the parsed and validated atlantis.yaml config for the
// repo at absRepoDir.
// If there was no config file, it will return an os.IsNotExist(error).
func (p *ParserValidator) ParseRepoCfg(absRepoDir string, globalCfg valid.GlobalCfg, repoID string) (valid.RepoCfg, error) {
	configFile := p.repoCfgPath(absRepoDir, AtlantisYAMLFilename)
	configData, err := os.ReadFile(configFile) // nolint: gosec

	if err != nil {
		if !os.IsNotExist(err) {
			return valid.RepoCfg{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
		}
		// Don't wrap os.IsNotExist errors because we want our callers to be
		// able to detect if it's a NotExist err.
		return valid.RepoCfg{}, err
	}
	return p.ParseRepoCfgData(configData, globalCfg, repoID)
}

func (p *ParserValidator) ParseRepoCfgData(repoCfgData []byte, globalCfg valid.GlobalCfg, repoID string) (valid.RepoCfg, error) {
	var rawConfig raw.RepoCfg
	if err := yaml.UnmarshalStrict(repoCfgData, &rawConfig); err != nil {
		return valid.RepoCfg{}, err
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"
	if err := rawConfig.Validate(); err != nil {
		return valid.RepoCfg{}, err
	}

	validConfig := rawConfig.ToValid()

	// We do the project name validation after we get the valid config because
	// we need the defaults of dir and workspace to be populated.
	if err := p.validateProjectNames(validConfig); err != nil {
		return valid.RepoCfg{}, err
	}
	if validConfig.Version == 2 {
		// The only difference between v2 and v3 is how we parse custom run
		// commands.
		if err := p.applyLegacyShellParsing(&validConfig); err != nil {
			return validConfig, err
		}
	}

	err := globalCfg.ValidateRepoCfg(validConfig, repoID)
	return validConfig, err
}

// ParseGlobalCfg returns the parsed and validated global repo config file at
// configFile. defaultCfg will be merged into the parsed config.
// If there is no file at configFile it will return an error.
func (p *ParserValidator) ParseGlobalCfg(configFile string, defaultCfg valid.GlobalCfg) (valid.GlobalCfg, error) {
	configData, err := os.ReadFile(configFile) // nolint: gosec
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

	return p.validateRawGlobalCfg(rawCfg, defaultCfg, "yaml")
}

// ParseGlobalCfgJSON parses a json string cfgJSON into global config.
func (p *ParserValidator) ParseGlobalCfgJSON(cfgJSON string, defaultCfg valid.GlobalCfg) (valid.GlobalCfg, error) {
	var rawCfg raw.GlobalCfg
	err := json.Unmarshal([]byte(cfgJSON), &rawCfg)
	if err != nil {
		return valid.GlobalCfg{}, err
	}
	return p.validateRawGlobalCfg(rawCfg, defaultCfg, "json")
}

func (p *ParserValidator) validateRawGlobalCfg(rawCfg raw.GlobalCfg, defaultCfg valid.GlobalCfg, errTag string) (valid.GlobalCfg, error) {
	// Setting ErrorTag means our errors will use the field names defined in
	// the struct tags for yaml/json.
	validation.ErrorTag = errTag
	if err := rawCfg.Validate(); err != nil {
		return valid.GlobalCfg{}, err
	}

	validCfg := rawCfg.ToValid(defaultCfg)
	return validCfg, nil
}

func (p *ParserValidator) repoCfgPath(repoDir, cfgFilename string) string {
	return filepath.Join(repoDir, cfgFilename)
}

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

// applyLegacyShellParsing changes any custom run commands in cfg to use the old
// parsing method with shlex.Split().
func (p *ParserValidator) applyLegacyShellParsing(cfg *valid.RepoCfg) error {
	legacyParseF := func(s *valid.Step) error {
		if s.StepName == "run" {
			split, err := shlex.Split(s.RunCommand)
			if err != nil {
				return errors.Wrapf(err, "unable to parse %q", s.RunCommand)
			}
			s.RunCommand = strings.Join(split, " ")
		}
		return nil
	}

	for k := range cfg.Workflows {
		w := cfg.Workflows[k]
		for i := range w.Plan.Steps {
			s := &w.Plan.Steps[i]
			if err := legacyParseF(s); err != nil {
				return err
			}
		}
		for i := range w.Apply.Steps {
			s := &w.Apply.Steps[i]
			if err := legacyParseF(s); err != nil {
				return err
			}
		}
		cfg.Workflows[k] = w
	}
	return nil
}
