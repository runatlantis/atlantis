package raw

type DriftDetection struct {
	Enabled *bool   `yaml:"enabled,omitempty"`
	Cron    *string `yaml:"cron,omitempty"`
}
