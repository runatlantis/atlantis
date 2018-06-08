package yaml

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

// AtlantisYAMLFilename is the name of the config file for each repo.
const AtlantisYAMLFilename = "atlantis.yaml"

type ParserValidator struct{}

// ReadConfig returns the parsed and validated atlantis.yaml config for repoDir.
// If there was no config file, then this can be detected by checking the type
// of error: os.IsNotExist(error).
func (r *ParserValidator) ReadConfig(repoDir string) (Config, error) {
	configFile := filepath.Join(repoDir, AtlantisYAMLFilename)
	configData, err := ioutil.ReadFile(configFile)

	// NOTE: the error we return here must also be os.IsNotExist since that's
	// what our callers use to detect a missing config file.
	if err != nil && os.IsNotExist(err) {
		return Config{}, err
	}

	// If it exists but we couldn't read it return an error.
	if err != nil {
		return Config{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
	}

	// If the config file exists, parse it.
	config, err := r.parseAndValidate(configData)
	if err != nil {
		return Config{}, errors.Wrapf(err, "parsing %s", AtlantisYAMLFilename)
	}
	return config, err
}

func (r *ParserValidator) parseAndValidate(configData []byte) (Config, error) {
	var repoConfig Config
	if err := yaml.UnmarshalStrict(configData, &repoConfig); err != nil {
		return repoConfig, err
	}

	// Validate version.
	if repoConfig.Version != 2 {
		// todo: this will fail old atlantis.yaml files, we should deal with them in a better way.
		return repoConfig, errors.New("unknown version: must have \"version: 2\" set")
	}

	// Validate projects.
	if len(repoConfig.Projects) == 0 {
		return repoConfig, errors.New("'projects' key must exist and contain at least one element")
	}

	for i, project := range repoConfig.Projects {
		if project.Dir == "" {
			return repoConfig, fmt.Errorf("project at index %d invalid: dir key must be set and non-empty", i)
		}
	}
	return repoConfig, nil
}
