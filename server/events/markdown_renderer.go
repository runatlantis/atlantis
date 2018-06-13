// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
package events

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

// MarkdownRenderer renders responses as markdown.
type MarkdownRenderer struct{}

// CommonData is data that all responses have.
type CommonData struct {
	Command string
	Verbose bool
	Log     string
}

// ErrData is data about an error response.
type ErrData struct {
	Error string
	CommonData
}

// FailureData is data about a failure response.
type FailureData struct {
	Failure string
	CommonData
}

// ResultData is data about a successful response.
type ResultData struct {
	Results []ProjectResultTmplData
	CommonData
}

type ProjectResultTmplData struct {
	Workspace string
	Dir       string
	Rendered  string
}

// Render formats the data into a markdown string.
// nolint: interfacer
func (m *MarkdownRenderer) Render(res CommandResponse, cmdName CommandName, log string, verbose bool) string {
	commandStr := strings.Title(cmdName.String())
	common := CommonData{commandStr, verbose, log}
	if res.Error != nil {
		return m.renderTemplate(errWithLogTmpl, ErrData{res.Error.Error(), common})
	}
	if res.Failure != "" {
		return m.renderTemplate(failureWithLogTmpl, FailureData{res.Failure, common})
	}
	return m.renderProjectResults(res.ProjectResults, common)
}

func (m *MarkdownRenderer) renderProjectResults(pathResults []ProjectResult, common CommonData) string {
	var resultsTmplData []ProjectResultTmplData

	for _, result := range pathResults {
		resultData := ProjectResultTmplData{
			Workspace: result.Workspace,
			Dir:       result.Path,
		}
		if result.Error != nil {
			resultData.Rendered = m.renderTemplate(errTmpl, struct {
				Command string
				Error   string
			}{
				Command: common.Command,
				Error:   result.Error.Error(),
			})
		} else if result.Failure != "" {
			resultData.Rendered = m.renderTemplate(failureTmpl, struct {
				Command string
				Failure string
			}{
				Command: common.Command,
				Failure: result.Failure,
			})
		} else if result.PlanSuccess != nil {
			resultData.Rendered = m.renderTemplate(planSuccessTmpl, *result.PlanSuccess)
		} else if result.ApplySuccess != "" {
			resultData.Rendered = m.renderTemplate(applySuccessTmpl, struct{ Output string }{result.ApplySuccess})
		} else {
			resultData.Rendered = "Found no template. This is a bug!"
		}
		resultsTmplData = append(resultsTmplData, resultData)
	}

	var tmpl *template.Template
	if len(resultsTmplData) == 1 {
		tmpl = singleProjectTmpl
	} else {
		tmpl = multiProjectTmpl
	}
	return m.renderTemplate(tmpl, ResultData{resultsTmplData, common})
}

func (m *MarkdownRenderer) renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}

var singleProjectTmpl = template.Must(template.New("").Parse("{{$result := index .Results 0}}Ran {{.Command}} in dir: `{{$result.Dir}}` workspace: `{{$result.Workspace}}`\n{{$result.Rendered}}\n" + logTmpl))
var multiProjectTmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(
	"Ran {{.Command}} for {{ len .Results }} projects:\n" +
		"{{ range $result := .Results }}" +
		"1. workspace: `{{$result.Workspace}}` path: `{{$result.Dir}}`\n" +
		"{{end}}\n" +
		"{{ range $i, $result := .Results }}" +
		"### {{add $i 1}}. workspace: `{{$result.Workspace}}` path: `{{$result.Dir}}`\n" +
		"{{$result.Rendered}}\n" +
		"---\n{{end}}" +
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
