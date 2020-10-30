package raw

// PolicySets is the raw schema for repo-level atlantis.yaml config.
type PolicySets struct {
	Version    *int        `yaml:"version,omitempty"`
	PolicySets []PolicySet `yaml:"policies,omitempty"`
}

type PolicySet struct {
	Path   string   `yaml:"path"`
	Source string   `yaml:"source"`
	Name   string   `yaml:"name"`
	Owners []string `yaml:"owners"`
}
