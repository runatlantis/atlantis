package server

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

var singleProjectTmpl = template.Must(template.New("").Parse("{{ range $result := .Results }}{{$result}}{{end}}\n" + logTmpl))
var multiProjectTmpl = template.Must(template.New("").Parse(
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

type CommonData struct {
	Command string
	Verbose bool
	Log     string
}

type ErrData struct {
	Error string
	CommonData
}

type FailureData struct {
	Failure string
	CommonData
}

type ResultData struct {
	Results map[string]string
	CommonData
}

func (g *GithubCommentRenderer) render(res CommandResponse, log string, verbose bool) string {
	commandStr := strings.Title(res.Command.String())
	common := CommonData{commandStr, verbose, log}
	if res.Error != nil {
		return g.renderTemplate(errWithLogTmpl, ErrData{res.Error.Error(), common})
	}
	if res.Failure != "" {
		return g.renderTemplate(failureWithLogTmpl, FailureData{res.Failure, common})
	}
	return g.renderProjectResults(res.ProjectResults, common)
}

func (g *GithubCommentRenderer) renderProjectResults(pathResults []ProjectResult, common CommonData) string {
	results := make(map[string]string)
	for _, result := range pathResults {
		if result.Error != nil {
			results[result.Path] = g.renderTemplate(errTmpl, struct {
				Command string
				Error   string
			}{
				Command: common.Command,
				Error:   result.Error.Error(),
			})
		} else if result.Failure != "" {
			results[result.Path] = g.renderTemplate(failureTmpl, struct {
				Command string
				Failure string
			}{
				Command: common.Command,
				Failure: result.Failure,
			})
		} else if result.PlanSuccess != nil {
			results[result.Path] = g.renderTemplate(planSuccessTmpl, *result.PlanSuccess)
		} else if result.ApplySuccess != "" {
			results[result.Path] = g.renderTemplate(applySuccessTmpl, struct{ Output string }{result.ApplySuccess})
		} else {
			results[result.Path] = "Found no template. This is a bug!"
		}
	}

	var tmpl *template.Template
	if len(results) == 1 {
		tmpl = singleProjectTmpl
	} else {
		tmpl = multiProjectTmpl
	}
	return g.renderTemplate(tmpl, ResultData{results, common})
}

func (g *GithubCommentRenderer) renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}
