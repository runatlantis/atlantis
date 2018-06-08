package yaml

type Project struct {
	Dir               string    `yaml:"dir"`
	Workspace         string    `yaml:"workspace"`
	Workflow          string    `yaml:"workflow"`
	TerraformVersion  string    `yaml:"terraform_version"`
	AutoPlan          *AutoPlan `yaml:"auto_plan,omitempty"`
	ApplyRequirements []string  `yaml:"apply_requirements"`
}

func (p *Project) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Use a type alias so unmarshal doesn't get into an infinite loop.
	type alias Project
	// Set up defaults.
	defaults := alias{
		Workspace: defaultWorkspace,
		AutoPlan: &AutoPlan{
			Enabled:      true,
			WhenModified: []string{"**/*.tf"},
		},
	}
	if err := unmarshal(&defaults); err != nil {
		return err
	}
	*p = Project(defaults)
	return nil
}
