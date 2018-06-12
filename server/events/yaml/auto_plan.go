package yaml

type AutoPlan struct {
	WhenModified []string `yaml:"when_modified,omitempty"`
	Enabled      *bool    `yaml:"enabled,omitempty"`
}
