package yaml

type Stage struct {
	Steps []StepConfig `yaml:"steps"` // can either be a built in step like 'plan' or a custom step like 'run: echo hi'
}
