package server

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var singlePathTmpl = template.Must(template.New("").Parse("{{ range $result := .Results }}{{$result}}{{end}}\n" + logTmpl))
var multiPathTmpl = template.Must(template.New("").Parse(
	"Ran {{.Command}} in {{ len .Results }} directories:\n" +
		"{{ range $path, $result := .Results }}" +
		" * `{{$path}}`\n" +
		"{{end}}\n" +
		"{{ range $path, $result := .Results }}" +
		"##{{$path}}/\n" +
		"{{$result}}\n" +
		"---{{end}}" +
		logTmpl))
var planSuccessTmpl = template.Must(template.New("").Parse(
	"```diff\n" +
		"{{.TerraformOutput}}\n" +
		"```\n\n" +
		"* To **discard** this plan click [here]({{.LockURL}})."))
var applySuccessTmpl = template.Must(template.New("").Parse(
	"```diff\n" +
		"{{.Output}}\n" +
		"```"))
var errTmplText = "**{{.Command}} Error**\n" +
	"```\n" +
	"{{.Error}}\n" +
	"```\n"
var errTmpl = template.Must(template.New("").Parse(errTmplText))
var errWithLogTmpl = template.Must(template.New("").Parse(errTmplText + logTmpl))
var failureTmplText = "**{{.Command}} Failed**: {{.Failure}}\n"
var failureTmpl = template.Must(template.New("").Parse(failureTmplText))
var failureWithLogTmpl = template.Must(template.New("").Parse(failureTmplText + logTmpl))
var logTmpl = "{{if .Verbose}}\n<details><summary>Log</summary>\n  <p>\n\n```\n{{.Log}}```\n</p></details>{{end}}\n"

// GithubCommentRenderer renders responses as GitHub comments
type GithubCommentRenderer struct{}


func (g *GithubCommentRenderer) render(res CommandResponse, log string, verbose bool) string {
	commandStr := strings.Title(res.Command.String())
	if res.Error != nil {
		return g.renderTemplate(errWithLogTmpl, struct{
			Command string
			Error string
			Verbose bool
			Log string
		}{commandStr, res.Error.Error(), verbose, log})
	}
	if res.Failure != "" {
		return g.renderTemplate(failureWithLogTmpl, struct{
			Command string
			Failure string
			Verbose bool
			Log string
		}{commandStr, res.Failure, verbose, log})
	}
	return g.renderProjectResults(res.ProjectResults, commandStr, log, verbose)
}

func (g *GithubCommentRenderer) renderProjectResults(pathResults []ProjectResult, command string, log string, verbose bool) string {
	renderedOutputs := make(map[string]string)
	for _, result := range pathResults {
		if result.Error != nil {
			renderedOutputs[result.Path] = g.renderTemplate(errTmpl, struct{
				Command string
				Output string
			}{
				Command: command,
				Output: result.Error.Error(),
			})
		} else if result.Failure != "" {
			renderedOutputs[result.Path] = g.renderTemplate(failureTmpl, struct{
				Command string
				Failure string
			}{
				Command: command,
				Failure: result.Failure,
			})
		} else if result.PlanSuccess != nil {
			renderedOutputs[result.Path] = g.renderTemplate(planSuccessTmpl, *result.PlanSuccess)
		} else if result.ApplySuccess != "" {
			renderedOutputs[result.Path] = g.renderTemplate(applySuccessTmpl, struct{Output string}{result.ApplySuccess})
		} else {
			renderedOutputs[result.Path] = "Found no template. This is a bug!"
		}
	}

	var tmpl *template.Template
	if len(renderedOutputs) == 1 {
		tmpl = singlePathTmpl
	} else {
		tmpl = multiPathTmpl
	}
	return g.renderTemplate(tmpl, struct {
		Results map[string]string
		Log     string
		Verbose bool
		Command string
	}{renderedOutputs, log, verbose, command})
}

func (g *GithubCommentRenderer) renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}
