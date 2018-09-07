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

package events

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

const (
	planCommandTitle  = "Plan"
	applyCommandTitle = "Apply"
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
	Results []projectResultTmplData
	CommonData
}

type projectResultTmplData struct {
	Workspace  string
	RepoRelDir string
	Rendered   string
}

// Render formats the data into a markdown string.
// nolint: interfacer
func (m *MarkdownRenderer) Render(res CommandResult, cmdName CommandName, log string, verbose bool) string {
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

func (m *MarkdownRenderer) renderProjectResults(results []ProjectResult, common CommonData) string {
	var resultsTmplData []projectResultTmplData
	numPlanSuccesses := 0

	for _, result := range results {
		resultData := projectResultTmplData{
			Workspace:  result.Workspace,
			RepoRelDir: result.RepoRelDir,
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
			numPlanSuccesses++
		} else if result.ApplySuccess != "" {
			resultData.Rendered = m.renderTemplate(applySuccessTmpl, struct{ Output string }{result.ApplySuccess})
		} else {
			resultData.Rendered = "Found no template. This is a bug!"
		}
		resultsTmplData = append(resultsTmplData, resultData)
	}

	var tmpl *template.Template
	switch {
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses > 0:
		tmpl = singleProjectPlanSuccessTmpl
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses == 0:
		tmpl = singleProjectPlanUnsuccessfulTmpl
	case len(resultsTmplData) == 1 && common.Command == applyCommandTitle:
		tmpl = singleProjectApplyTmpl
	case common.Command == planCommandTitle:
		tmpl = multiProjectPlanTmpl
	case common.Command == applyCommandTitle:
		tmpl = multiProjectApplyTmpl
	default:
		return "no template matched–this is a bug"
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

// todo: refactor to remove duplication #refactor
var singleProjectApplyTmpl = template.Must(template.New("").Parse(
	"{{$result := index .Results 0}}Ran {{.Command}} in dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n{{$result.Rendered}}\n" + logTmpl))
var singleProjectPlanSuccessTmpl = template.Must(template.New("").Parse(
	"{{$result := index .Results 0}}Ran {{.Command}} in dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n{{$result.Rendered}}\n" +
		"\n" +
		"---\n" +
		"* :fast_forward: To **apply** all unapplied plans, comment:\n" +
		"  * `atlantis apply`" + logTmpl))
var singleProjectPlanUnsuccessfulTmpl = template.Must(template.New("").Parse(
	"{{$result := index .Results 0}}Ran {{.Command}} in dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n" +
		"{{$result.Rendered}}\n" + logTmpl))
var multiProjectPlanTmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(
	"Ran {{.Command}} for {{ len .Results }} projects:\n" +
		"{{ range $result := .Results }}" +
		"1. workspace: `{{$result.Workspace}}` dir: `{{$result.RepoRelDir}}`\n" +
		"{{end}}\n" +
		"{{ range $i, $result := .Results }}" +
		"### {{add $i 1}}. workspace: `{{$result.Workspace}}` dir: `{{$result.RepoRelDir}}`\n" +
		"{{$result.Rendered}}\n" +
		"---\n{{end}}{{ if gt (len .Results) 0 }}* :fast_forward: To **apply** all unapplied plans, comment:\n" +
		"  * `atlantis apply`{{end}}" +
		logTmpl))
var multiProjectApplyTmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(
	"Ran {{.Command}} for {{ len .Results }} projects:\n" +
		"{{ range $result := .Results }}" +
		"1. workspace: `{{$result.Workspace}}` dir: `{{$result.RepoRelDir}}`\n" +
		"{{end}}\n" +
		"{{ range $i, $result := .Results }}" +
		"### {{add $i 1}}. workspace: `{{$result.Workspace}}` dir: `{{$result.RepoRelDir}}`\n" +
		"{{$result.Rendered}}\n" +
		"---\n{{end}}" +
		logTmpl))
var planSuccessTmpl = template.Must(template.New("").Parse(
	"```diff\n" +
		"{{.TerraformOutput}}\n" +
		"```\n\n" +
		"* :arrow_forward: To **apply** this plan, comment:\n" +
		"  * `{{.ApplyCmd}}`\n" +
		"* :put_litter_in_its_place: To **delete** this plan click [here]({{.LockURL}})\n" +
		"* :repeat: To **plan** this project again, comment:\n" +
		"  * `{{.RePlanCmd}}`"))
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
