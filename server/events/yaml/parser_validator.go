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
// of error: os.IsNotExist(error) but it's instead preferred to check with
// HasConfigFile.
func (p *ParserValidator) ReadConfig(repoDir string, repoConfig raw.RepoConfig, repoName string, allowAllRepoConfig bool) (valid.Config, error) {
	configFile := p.configFilePath(repoDir)
	configData, err := ioutil.ReadFile(configFile) // nolint: gosec

	// NOTE: the error we return here must also be os.IsNotExist since that's
	// what our callers use to detect a missing config file.
	if err != nil && os.IsNotExist(err) {
		return valid.Config{}, err
	}

	// If it exists but we couldn't read it return an error.
	if err != nil {
		return valid.Config{}, errors.Wrapf(err, "unable to read %s file", AtlantisYAMLFilename)
	}

	// If the config file exists, parse it.
	config, err := p.parseAndValidate(configData, repoConfig, repoName, allowAllRepoConfig)
	if err != nil {
		return valid.Config{}, errors.Wrapf(err, "parsing %s", AtlantisYAMLFilename)
	}
	return config, err
}

func (p *ParserValidator) ReadServerConfig(configFile string) (raw.RepoConfig, error) {
	configData, err := ioutil.ReadFile(configFile) // nolint: gosec
	if err != nil {
		return raw.RepoConfig{}, errors.Wrapf(err, "unable to read %s file", configFile)
	}
	config, err := p.parseAndValidateServerConfig(configData)
	if err != nil {
		return raw.RepoConfig{}, errors.Wrapf(err, "parsing %s", configFile)
	}
	return config, err
}

func (p *ParserValidator) HasConfigFile(repoDir string) (bool, error) {
	_, err := os.Stat(p.configFilePath(repoDir))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err == nil {
		return true, nil
	}
	return false, err
}

func (p *ParserValidator) configFilePath(repoDir string) string {
	return filepath.Join(repoDir, AtlantisYAMLFilename)
}

func (p *ParserValidator) parseAndValidateServerConfig(configData []byte) (raw.RepoConfig, error) {
	var config raw.RepoConfig
	if err := yaml.UnmarshalStrict(configData, &config); err != nil {
		return raw.RepoConfig{}, err
	}

	validation.ErrorTag = "yaml"

	if err := config.Validate(); err != nil {
		return raw.RepoConfig{}, err
	}

	if err := p.validateRepoWorkflows(config); err != nil {
		return raw.RepoConfig{}, err
	}
	return config, nil
}

func (p *ParserValidator) parseAndValidate(configData []byte, repoConfig raw.RepoConfig, repoName string, allowAllRepoConfig bool) (valid.Config, error) {
	var rawConfig raw.Config
	if err := yaml.UnmarshalStrict(configData, &rawConfig); err != nil {
		return valid.Config{}, err
	}

	// Set ErrorTag to yaml so it uses the YAML field names in error messages.
	validation.ErrorTag = "yaml"

	var err error
	rawConfig, err = p.ValidateOverridesAndMergeConfig(rawConfig, repoConfig, repoName, allowAllRepoConfig)
	if err != nil {
		return valid.Config{}, err
	}

	if err := rawConfig.Validate(); err != nil {
		return valid.Config{}, err
	}

	// Top level validation.
	if err := p.validateWorkflows(rawConfig); err != nil {
		return valid.Config{}, err
	}

	validConfig := rawConfig.ToValid()
	if err := p.validateProjectNames(validConfig); err != nil {
		return valid.Config{}, err
	}

	return validConfig, nil
}

func (p *ParserValidator) getOverrideErrorMessage(key string) error {
	return fmt.Errorf("%q cannot be specified in %q by default.  To enable this, add %q to %q in the server side repo config", key, AtlantisYAMLFilename, key, raw.AllowedOverridesKey)
}

// Checks any sensitive fields present in atlantis.yaml against the list of allowed overrides and merge the configuration
// from the server side repo config with project settings found in atlantis.yaml
func (p *ParserValidator) ValidateOverridesAndMergeConfig(config raw.Config, repoConfig raw.RepoConfig, repoName string, allowAllRepoConfig bool) (raw.Config, error) {
	var finalProjects []raw.Project

	// Start with a repo regex that will match everything, but sets no allowed_overrides. This will
	// provide a default behavior of "deny all overrides" if no server side defined repos are matched
	lastMatchingRepo := raw.Repo{ID: "/.*/"}

	// Find the last repo to match.  If multiple are found, the last matched repo's settings will be used
	for _, repo := range repoConfig.Repos {
		matches, err := repo.Matches(repoName)
		if err != nil {
			return config, err
		} else if matches {
			lastMatchingRepo = repo
		}
	}

	for _, project := range config.Projects {
		// If atlantis.yaml has apply requirements, only honor them if this key is allowed in a server side
		// --repo-config or if --allow-repo-config is specified.
		if len(project.ApplyRequirements) > 0 && !(allowAllRepoConfig || lastMatchingRepo.IsOverrideAllowed(raw.ApplyRequirementsKey)) {
			return config, p.getOverrideErrorMessage(raw.ApplyRequirementsKey)
		}

		// Do not allow projects to specify a workflow unless it is explicitly allowed
		if project.Workflow != nil && !(allowAllRepoConfig || lastMatchingRepo.IsOverrideAllowed(raw.WorkflowKey)) {
			return config, p.getOverrideErrorMessage(raw.WorkflowKey)
		} else if project.Workflow == nil && lastMatchingRepo.Workflow != nil {
			project.Workflow = lastMatchingRepo.Workflow
		}

		finalProjects = append(finalProjects, project)
	}
	config.Projects = finalProjects

	if len(config.Workflows) > 0 && !(allowAllRepoConfig || lastMatchingRepo.AllowCustomWorkflows) {
		return config, fmt.Errorf("%q cannot be specified in %q by default.  To enable this, set %q to true in the server side repo config", raw.CustomWorkflowsKey, AtlantisYAMLFilename, raw.CustomWorkflowsKey)
	} else if len(config.Workflows) == 0 {
		if len(repoConfig.Workflows) > 0 {
			config.Workflows = repoConfig.Workflows
		}
	}
	return config, nil
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

func (p *ParserValidator) validateWorkflows(config raw.Config) error {
	for _, project := range config.Projects {
		if err := p.validateWorkflowExists(project, config.Workflows); err != nil {
			return err
		}
	}
	return nil
}

func (p *ParserValidator) validateRepoWorkflows(config raw.RepoConfig) error {
	for _, repo := range config.Repos {
		if err := p.validateRepoWorkflowExists(repo, config.Workflows); err != nil {
			return err
		}
	}
	return nil
}

func (p *ParserValidator) validateRepoWorkflowExists(repo raw.Repo, workflows map[string]raw.Workflow) error {
	if repo.Workflow == nil {
		return nil
	}
	workflow := *repo.Workflow
	for w := range workflows {
		if w == workflow {
			return nil
		}
	}
	return fmt.Errorf("workflow %q is not defined", workflow)
}

func (p *ParserValidator) validateWorkflowExists(project raw.Project, workflows map[string]raw.Workflow) error {
	if project.Workflow == nil {
		return nil
	}
	workflow := *project.Workflow
	for k := range workflows {
		if k == workflow {
			return nil
		}
	}
	return fmt.Errorf("workflow %q is not defined", workflow)
}
