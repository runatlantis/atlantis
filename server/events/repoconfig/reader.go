package repoconfig

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
)

const AtlantisYAMLFilename = "atlantis.yaml"
const PlanStageName = "plan"
const ApplyStageName = "apply"

type Reader struct{}

// ReadConfig returns the parsed and validated config for repoDir.
// If there was no config, it returns a nil pointer.
func (r *Reader) ReadConfig(repoDir string) (*RepoConfig, error) {
	configFile := filepath.Join(repoDir, AtlantisYAMLFilename)
	configData, err := ioutil.ReadFile(configFile)

	// If the file doesn't exist return nil.
	if err != nil && os.IsNotExist(err) {
		return nil, nil
	}

	// If it exists but we couldn't read it return an error.
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
	}

	// If the config file exists, parse it.
	config, err := r.parseAndValidate(configData)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing %s", AtlantisYAMLFilename)
	}
	return &config, err
}

func (r *Reader) parseAndValidate(configData []byte) (RepoConfig, error) {
	var repoConfig RepoConfig
	if err := yaml.UnmarshalStrict(configData, &repoConfig); err != nil {
		// Unmarshal error messages aren't fit for user output. We need to
		// massage them.
		// todo: fix "field autoplan not found in struct repoconfig.alias" errors
		return repoConfig, errors.New(strings.Replace(err.Error(), " into repoconfig.RepoConfig", "", -1))
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
