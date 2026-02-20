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
func (p *PluginCmd) addCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name>",
		Short: "Show configuration instructions for a VCS provider plugin",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			pl, ok := p.Registry.Get(name)
			if !ok {
				return fmt.Errorf("plugin %q is not registered; run \"atlantis plugin list\" to see available plugins", name)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Plugin %q (v%s)\n", pl.Name(), pl.Version())
			fmt.Fprintf(out, "%s\n\n", pl.Description())

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
