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

// ReadConfig returns the parsed and validated atlantis.yaml config for repoDir.
// If there was no config file, then this can be detected by checking the type
// of error: os.IsNotExist(error).
func (r *ParserValidator) ReadConfig(repoDir string) (valid.Spec, error) {
	configFile := filepath.Join(repoDir, AtlantisYAMLFilename)
	configData, err := ioutil.ReadFile(configFile)

	// NOTE: the error we return here must also be os.IsNotExist since that's
	// what our callers use to detect a missing config file.
	if err != nil && os.IsNotExist(err) {
		return valid.Spec{}, err
	}

	// If it exists but we couldn't read it return an error.
	if err != nil {
		return valid.Spec{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
	}

	// If the config file exists, parse it.
	config, err := r.parseAndValidate(configData)
	if err != nil {
		return valid.Spec{}, errors.Wrapf(err, "parsing %s", AtlantisYAMLFilename)
	}
	return config, err
}

func (r *ParserValidator) parseAndValidate(configData []byte) (valid.Spec, error) {
	var rawSpec raw.Spec
	if err := yaml.UnmarshalStrict(configData, &rawSpec); err != nil {
		return valid.Spec{}, err
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"

	if err := rawSpec.Validate(); err != nil {
		return valid.Spec{}, err
	}

	// Top level validation.
	for _, p := range rawSpec.Projects {
		if p.Workflow != nil {
			workflow := *p.Workflow
			found := false
			for k := range rawSpec.Workflows {
				if k == workflow {
					found = true
				}
			}
			if !found {
				return valid.Spec{}, fmt.Errorf("workflow %q is not defined", workflow)
			}
		}
	}

	return rawSpec.ToValid(), nil
}
