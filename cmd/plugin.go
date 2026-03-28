// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/runatlantis/atlantis/plugin"
	"github.com/spf13/cobra"
)

// PluginCmd manages VCS provider plugins for Atlantis.
type PluginCmd struct {
	// Registry is the plugin registry to operate on.
	// Defaults to plugin.DefaultRegistry when constructed via main.
	Registry *plugin.Registry
}

// Init returns the runnable cobra command for plugin management.
func (p *PluginCmd) Init() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Manage Atlantis VCS provider plugins",
	}
	cmd.AddCommand(p.listCmd())
	cmd.AddCommand(p.addCmd())
	return cmd
}

// listCmd returns the "atlantis plugin list" subcommand.
func (p *PluginCmd) listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available VCS provider plugins",
		RunE: func(cmd *cobra.Command, _ []string) error {
			plugins := p.Registry.List()
			if len(plugins) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No plugins registered.")
				return nil
			}
			sort.Slice(plugins, func(i, j int) bool {
				return plugins[i].Name() < plugins[j].Name()
			})
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tDESCRIPTION\tVERSION")
			for _, pl := range plugins {
				fmt.Fprintf(w, "%s\t%s\t%s\n", pl.Name(), pl.Description(), pl.Version())
			}
			return w.Flush()
		},
	}
}

// addCmd returns the "atlantis plugin add <name>" subcommand.
// If the plugin is installed it shows its configuration requirements along with
// its source URL. If the plugin is not installed but is a known plugin in the
// built-in catalog, it shows download instructions. Unknown plugin names produce
// an error.
func (p *PluginCmd) addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Show configuration for an installed plugin or download instructions for a known one",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			out := cmd.OutOrStdout()

			pl, ok := p.Registry.Get(name)
			if !ok {
				// Not installed locally; check the built-in catalog.
				sourceURL, known := plugin.LookupSource(name)
				if !known {
					return fmt.Errorf("plugin %q is not registered and is not a known plugin; run \"atlantis plugin list\" to see installed plugins", name)
				}
				fmt.Fprintf(out, "Plugin %q is not installed locally.\n\n", name)
				fmt.Fprintf(out, "Source:  %s\n", sourceURL)
				fmt.Fprintf(out, "To install, download the binary for your platform from:\n")
				fmt.Fprintf(out, "  %s/releases\n\n", sourceURL)
				fmt.Fprintln(out, "After installing, restart Atlantis for the plugin to take effect.")
				return nil
			}

			fmt.Fprintf(out, "Plugin %q (v%s)\n", pl.Name(), pl.Version())
			fmt.Fprintf(out, "%s\n", pl.Description())
			if url := pl.SourceURL(); url != "" {
				fmt.Fprintf(out, "Source:  %s\n", url)
			}
			fmt.Fprintln(out)

			var required, optional []plugin.ConfigKey
			for _, k := range pl.ConfigKeys() {
				if k.Required {
					required = append(required, k)
				} else {
					optional = append(optional, k)
				}
			}

			if len(required) > 0 {
				fmt.Fprintln(out, "Required configuration:")
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				for _, k := range required {
					fmt.Fprintf(w, "  --%s\t(%s)\t%s\n", k.Flag, k.EnvVar, k.Desc)
				}
				if err := w.Flush(); err != nil {
					return err
				}
			}

			if len(optional) > 0 {
				fmt.Fprintln(out, "\nOptional configuration:")
				w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
				for _, k := range optional {
					fmt.Fprintf(w, "  --%s\t(%s)\t%s\n", k.Flag, k.EnvVar, k.Desc)
				}
				return w.Flush()
			}

			return nil
		},
	}
}
