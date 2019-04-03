package cmd

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"text/template"
)

// usageTmpl returns a cobra-compatible usage template that will be printed
// during the help output.
// This template prints help like:
//   --name=<value>
//    <description>
// We use it over the default template so that the output it easier to read.
func usageTmpl(stringFlags map[string]stringFlag, intFlags map[string]intFlag, boolFlags map[string]boolFlag) string {
	var flagNames []string
	for name, f := range stringFlags {
		if f.hidden {
			continue
		}
		flagNames = append(flagNames, name)
	}
	for name, f := range boolFlags {
		if f.hidden {
			continue
		}
		flagNames = append(flagNames, name)
	}
	for name, f := range intFlags {
		if f.hidden {
			continue
		}
		flagNames = append(flagNames, name)
	}
	sort.Strings(flagNames)

	type flag struct {
		Name        string
		Description string
		IsBoolFlag  bool
	}

	var flags []flag
	for _, name := range flagNames {
		var descrip string
		var isBool bool
		if f, ok := stringFlags[name]; ok {
			descripWithDefault := f.description
			if f.defaultValue != "" {
				descripWithDefault += fmt.Sprintf(" (default %q)", f.defaultValue)
			}
			descrip = to80CharCols(descripWithDefault)
			isBool = false
		} else if f, ok := boolFlags[name]; ok {
			descrip = to80CharCols(f.description)
			isBool = true
		} else if f, ok := intFlags[name]; ok {
			descripWithDefault := f.description
			if f.defaultValue != 0 {
				descripWithDefault += fmt.Sprintf(" (default %d)", f.defaultValue)
			}
			descrip = to80CharCols(descripWithDefault)
			isBool = false
		} else {
			panic("this is a bug")
		}

		flags = append(flags, flag{
			Name:        name,
			Description: descrip,
			IsBoolFlag:  isBool,
		})
	}

	tmpl := template.Must(template.New("").Parse(
		"  --{{.Name}}{{if not .IsBoolFlag}}=<value>{{end}}\n{{.Description}}\n"))
	var flagHelpOutput string
	for _, f := range flags {
		buf := &bytes.Buffer{}
		if err := tmpl.Execute(buf, f); err != nil {
			panic(err)
		}
		flagHelpOutput += buf.String()
	}

	// Most of this template is taken from cobra.Command.UsageTemplate()
	// but we're subbing out the "Flags:" section with our custom output.
	return fmt.Sprintf(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:

%s{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`, flagHelpOutput)
}

func to80CharCols(s string) string {
	var splitAt80 string
	splitSpaces := strings.Split(s, " ")
	var nextLine string
	for i, spaceSplit := range splitSpaces {
		if len(nextLine)+len(spaceSplit)+1 > 80 {
			splitAt80 += fmt.Sprintf("      %s\n", strings.TrimSuffix(nextLine, " "))
			nextLine = ""
		}
		if i == len(splitSpaces)-1 {
			nextLine += spaceSplit + " "
			splitAt80 += fmt.Sprintf("      %s\n", strings.TrimSuffix(nextLine, " "))
			break
		}
		nextLine += spaceSplit + " "
	}

	return splitAt80
}
