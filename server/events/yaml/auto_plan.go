package yaml

type AutoPlan struct {
	WhenModified []string `yaml:"when_modified"`
	Enabled      bool     `yaml:"enabled"`
}

// NOTE: AutoPlan does not implement UnmarshalYAML because we are unable to set
// defaults for bool and []string fields and so we just use the normal yaml
// unmarshalling.
