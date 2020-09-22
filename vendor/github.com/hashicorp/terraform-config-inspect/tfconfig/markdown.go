package tfconfig

import (
	"encoding/json"
	"io"
	"strings"
	"text/template"
)

func RenderMarkdown(w io.Writer, module *Module) error {
	tmpl := template.New("md")
	tmpl.Funcs(template.FuncMap{
		"tt": func(s string) string {
			return "`" + s + "`"
		},
		"commas": func(s []string) string {
			return strings.Join(s, ", ")
		},
		"json": func(v interface{}) (string, error) {
			j, err := json.Marshal(v)
			return string(j), err
		},
		"severity": func(s DiagSeverity) string {
			switch s {
			case DiagError:
				return "Error: "
			case DiagWarning:
				return "Warning: "
			default:
				return ""
			}
		},
	})
	template.Must(tmpl.Parse(markdownTemplate))
	return tmpl.Execute(w, module)
}

const markdownTemplate = `
# Module {{ tt .Path }}

{{- if .RequiredCore}}

Core Version Constraints:
{{- range .RequiredCore }}
* {{ tt . }}
{{- end}}{{end}}

{{- if .RequiredProviders}}

Provider Requirements:
{{- range $name, $req := .RequiredProviders }}
* **{{ $name }}{{ if $req.Source }} ({{ $req.Source | tt }}){{ end }}:** {{ if $req.VersionConstraints }}{{ commas $req.VersionConstraints | tt }}{{ else }}(any version){{ end }}
{{- end}}{{end}}

{{- if .Variables}}

## Input Variables
{{- range .Variables }}
* {{ tt .Name }}{{ if .Required }} (required){{else}} (default {{ json .Default | tt }}){{end}}
{{- if .Description}}: {{ .Description }}{{ end }}
{{- end}}{{end}}

{{- if .Outputs}}

## Output Values
{{- range .Outputs }}
* {{ tt .Name }}{{ if .Description}}: {{ .Description }}{{ end }}
{{- end}}{{end}}

{{- if .ManagedResources}}

## Managed Resources
{{- range .ManagedResources }}
* {{ printf "%s.%s" .Type .Name | tt }} from {{ tt .Provider.Name }}
{{- end}}{{end}}

{{- if .DataResources}}

## Data Resources
{{- range .DataResources }}
* {{ printf "data.%s.%s" .Type .Name | tt }} from {{ tt .Provider.Name }}
{{- end}}{{end}}

{{- if .ModuleCalls}}

## Child Modules
{{- range .ModuleCalls }}
* {{ tt .Name }} from {{ tt .Source }}{{ if .Version }} ({{ tt .Version }}){{ end }}
{{- end}}{{end}}

{{- if .Diagnostics}}

## Problems
{{- range .Diagnostics }}

## {{ severity .Severity }}{{ .Summary }}{{ if .Pos }}

(at {{ tt .Pos.Filename }} line {{ .Pos.Line }}{{ end }})
{{ if .Detail }}
{{ .Detail }}
{{- end }}

{{- end}}{{end}}

`
