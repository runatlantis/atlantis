package repoconfig

// InitStep runs `terraform init`.
type InitStep struct {
	ExtraArgs []string
	Meta      StepMeta
}

func (i *InitStep) Run() (string, error) {
	// If we're running < 0.9 we have to use `terraform get` instead of `init`.
	if MustConstraint("< 0.9.0").Check(i.Meta.TerraformVersion) {
		i.Meta.Log.Info("running terraform version %s so will use `get` instead of `init`", i.Meta.TerraformVersion)
		terraformGetCmd := append([]string{"get", "-no-color"}, i.ExtraArgs...)
		_, err := i.Meta.TerraformExecutor.RunCommandWithVersion(i.Meta.Log, i.Meta.AbsolutePath, terraformGetCmd, i.Meta.TerraformVersion, i.Meta.Workspace)
		return "", err
	} else {
		_, err := i.Meta.TerraformExecutor.RunCommandWithVersion(i.Meta.Log, i.Meta.AbsolutePath, append([]string{"init", "-no-color"}, i.ExtraArgs...), i.Meta.TerraformVersion, i.Meta.Workspace)
		return "", err
	}
}
