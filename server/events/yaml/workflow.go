package yaml

type Workflow struct {
	Apply *Stage `yaml:"apply,omitempty"`
	Plan  *Stage `yaml:"plan,omitempty"`
}
