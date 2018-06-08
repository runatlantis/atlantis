package yaml

type Config struct {
	Version   int                 `yaml:"version"`
	Projects  []Project           `yaml:"projects"`
	Workflows map[string]Workflow `yaml:"workflows"`
}
