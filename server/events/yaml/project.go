package yaml

type Project struct {
	Dir               *string   `yaml:"dir,omitempty"`
	Workspace         *string   `yaml:"workspace,omitempty"`
	Workflow          *string   `yaml:"workflow,omitempty"`
	TerraformVersion  *string   `yaml:"terraform_version,omitempty"`
	AutoPlan          *AutoPlan `yaml:"auto_plan,omitempty"`
	ApplyRequirements []string  `yaml:"apply_requirements,omitempty"`
}
