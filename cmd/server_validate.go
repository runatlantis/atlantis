package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/runatlantis/atlantis/server/logging"
)

// ServerValidateCmd validates an Atlantis server config file without starting
// the server. It checks for unknown keys that would be silently ignored.
type ServerValidateCmd struct {
	Viper  *viper.Viper
	Logger logging.SimpleLogging
}

// Init returns the runnable cobra command.
func (s *ServerValidateCmd) Init() *cobra.Command {
	c := &cobra.Command{
		Use:   "validate",
		Short: "Validate Atlantis server config file",
		Long: `Validate an Atlantis server config file for unknown keys.

Unknown keys are silently ignored by Atlantis at startup, which can lead to
misconfiguration (e.g. typos like "allow-draft-pr" instead of "allow-draft-prs").

This command reads the config file without starting the server and reports any
unrecognized keys.`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return s.run()
		},
	}

	c.Flags().String(ConfigFlag, "", "Path to the Atlantis server config file (required)\n")
	s.Viper.BindPFlag(ConfigFlag, c.Flags().Lookup(ConfigFlag)) // nolint: errcheck

	return c
}

func (s *ServerValidateCmd) run() error {
	configFile := s.Viper.GetString(ConfigFlag)
	if configFile == "" {
		return fmt.Errorf("--%s is required", ConfigFlag)
	}

	s.Viper.SetConfigFile(configFile)
	if err := s.Viper.ReadInConfig(); err != nil {
		return fmt.Errorf("reading config %s: %w", configFile, err)
	}

	unknowns := findUnknownKeys(s.Viper)
	if len(unknowns) > 0 {
		return fmt.Errorf("unknown keys in %s: %s", configFile, strings.Join(unknowns, ", "))
	}

	s.Logger.Info("config file %s is valid", configFile)
	return nil
}
