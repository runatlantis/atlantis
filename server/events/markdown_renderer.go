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
	"github.com/runatlantis/atlantis/server/events/models"
)

const (
	planCommandTitle  = "Plan"
	applyCommandTitle = "Apply"
	// maxUnwrappedLines is the maximum number of lines the Terraform output
	// can be before we wrap it in an expandable template.
	maxUnwrappedLines = 12
)

// MarkdownRenderer renders responses as markdown.
type MarkdownRenderer struct {
	// GitlabSupportsCommonMark is true if the version of GitLab we're
	// using supports the CommonMark markdown format.
	// If we're not configured with a GitLab client, this will be false.
	GitlabSupportsCommonMark bool
	DisableApplyAll          bool
	DisableMarkdownFolding   bool
}

// commonData is data that all responses have.
type commonData struct {
	Command         string
	Verbose         bool
	Log             string
	PlansDeleted    bool
	DisableApplyAll bool
}

// errData is data about an error response.
type errData struct {
	Error string
	commonData
}

// failureData is data about a failure response.
type failureData struct {
	Failure string
	commonData
}

// resultData is data about a successful response.
type resultData struct {
	Results []projectResultTmplData
	commonData
}

type planSuccessData struct {
	models.PlanSuccess
	PlanWasDeleted bool
}

type projectResultTmplData struct {
	Workspace   string
	RepoRelDir  string
	ProjectName string
	Rendered    string
}

// Render formats the data into a markdown string.
// nolint: interfacer
func (m *MarkdownRenderer) Render(res CommandResult, cmdName models.CommandName, log string, verbose bool, vcsHost models.VCSHostType) string {
	commandStr := strings.Title(cmdName.String())
	common := commonData{
		Command:         commandStr,
		Verbose:         verbose,
		Log:             log,
		PlansDeleted:    res.PlansDeleted,
		DisableApplyAll: m.DisableApplyAll,
	}
	if res.Error != nil {
		return m.renderTemplate(unwrappedErrWithLogTmpl, errData{res.Error.Error(), common})
	}
	if res.Failure != "" {
		return m.renderTemplate(failureWithLogTmpl, failureData{res.Failure, common})
	}
	return m.renderProjectResults(res.ProjectResults, common, vcsHost)
}

func (m *MarkdownRenderer) renderProjectResults(results []models.ProjectResult, common commonData, vcsHost models.VCSHostType) string {
	var resultsTmplData []projectResultTmplData
	numPlanSuccesses := 0

	for _, result := range results {
		resultData := projectResultTmplData{
			Workspace:   result.Workspace,
			RepoRelDir:  result.RepoRelDir,
			ProjectName: result.ProjectName,
		}
		if result.Error != nil {
			tmpl := unwrappedErrTmpl
			if m.shouldUseWrappedTmpl(vcsHost, result.Error.Error()) {
				tmpl = wrappedErrTmpl
			}
			resultData.Rendered = m.renderTemplate(tmpl, struct {
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
			if m.shouldUseWrappedTmpl(vcsHost, result.PlanSuccess.TerraformOutput) {
				resultData.Rendered = m.renderTemplate(planSuccessWrappedTmpl, planSuccessData{PlanSuccess: *result.PlanSuccess, PlanWasDeleted: common.PlansDeleted})
			} else {
				resultData.Rendered = m.renderTemplate(planSuccessUnwrappedTmpl, planSuccessData{PlanSuccess: *result.PlanSuccess, PlanWasDeleted: common.PlansDeleted})
			}
			numPlanSuccesses++
		} else if result.ApplySuccess != "" {
			if m.shouldUseWrappedTmpl(vcsHost, result.ApplySuccess) {
				resultData.Rendered = m.renderTemplate(applyWrappedSuccessTmpl, struct{ Output string }{result.ApplySuccess})
			} else {
				resultData.Rendered = m.renderTemplate(applyUnwrappedSuccessTmpl, struct{ Output string }{result.ApplySuccess})
			}
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
	return m.renderTemplate(tmpl, resultData{resultsTmplData, common})
}

// shouldUseWrappedTmpl returns true if we should use the wrapped markdown
// templates that collapse the output to make the comment smaller on initial
// load. Some VCS providers or versions of VCS providers don't support this
// syntax.
func (m *MarkdownRenderer) shouldUseWrappedTmpl(vcsHost models.VCSHostType, output string) bool {
	if m.DisableMarkdownFolding {
		return false
	}

	// Bitbucket Cloud and Server don't support the folding markdown syntax.
	if vcsHost == models.BitbucketServer || vcsHost == models.BitbucketCloud {
		return false
	}

	if vcsHost == models.Gitlab && !m.GitlabSupportsCommonMark {
		return false
	}

	return strings.Count(output, "\n") > maxUnwrappedLines
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
	"{{$result := index .Results 0}}Ran {{.Command}} for {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n{{$result.Rendered}}\n" + logTmpl))
var singleProjectPlanSuccessTmpl = template.Must(template.New("").Parse(
	"{{$result := index .Results 0}}Ran {{.Command}} for {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n{{$result.Rendered}}\n" +
		"\n" +
		"{{ if ne .DisableApplyAll true  }}---\n" +
		"* :fast_forward: To **apply** all unapplied plans from this pull request, comment:\n" +
		"    * `atlantis apply`\n" +
		"* :put_litter_in_its_place: To delete all plans and locks for the PR, comment:\n" +
		"    * `atlantis unlock`{{ end }}" + logTmpl))
var singleProjectPlanUnsuccessfulTmpl = template.Must(template.New("").Parse(
	"{{$result := index .Results 0}}Ran {{.Command}} for dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n\n" +
		"{{$result.Rendered}}\n" + logTmpl))
var multiProjectPlanTmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(
	"Ran {{.Command}} for {{ len .Results }} projects:\n\n" +
		"{{ range $result := .Results }}" +
		"1. {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n" +
		"{{end}}\n" +
		"{{ $disableApplyAll := .DisableApplyAll }}{{ range $i, $result := .Results }}" +
		"### {{add $i 1}}. {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n" +
		"{{$result.Rendered}}\n\n" +
		"{{ if ne $disableApplyAll true }}---\n{{end}}{{end}}{{ if ne .DisableApplyAll true }}{{ if and (gt (len .Results) 0) (not .PlansDeleted) }}* :fast_forward: To **apply** all unapplied plans from this pull request, comment:\n" +
		"    * `atlantis apply`\n" +
		"* :put_litter_in_its_place: To delete all plans and locks for the PR, comment:\n" +
		"    * `atlantis unlock`" +
		"{{end}}{{end}}" +
		logTmpl))
var multiProjectApplyTmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(
	"Ran {{.Command}} for {{ len .Results }} projects:\n\n" +
		"{{ range $result := .Results }}" +
		"1. {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n" +
		"{{end}}\n" +
		"{{ range $i, $result := .Results }}" +
		"### {{add $i 1}}. {{ if $result.ProjectName }}project: `{{$result.ProjectName}}` {{ end }}dir: `{{$result.RepoRelDir}}` workspace: `{{$result.Workspace}}`\n" +
		"{{$result.Rendered}}\n\n" +
		"---\n{{end}}" +
		logTmpl))
var planSuccessUnwrappedTmpl = template.Must(template.New("").Parse(
	"```diff\n" +
		"{{.TerraformOutput}}\n" +
		"```\n\n" + planNextSteps +
		"{{ if .HasDiverged }}\n\n:warning: The branch we're merging into is ahead, it is recommended to pull new commits first.{{end}}"))

var planSuccessWrappedTmpl = template.Must(template.New("").Parse(
	"<details><summary>Show Output</summary>\n\n" +
		"```diff\n" +
		"{{.TerraformOutput}}\n" +
		"```\n\n" +
		planNextSteps + "\n" +
		"</details>" +
		"{{ if .HasDiverged }}\n\n:warning: The branch we're merging into is ahead, it is recommended to pull new commits first.{{end}}"))

// planNextSteps are instructions appended after successful plans as to what
// to do next.
var planNextSteps = "{{ if .PlanWasDeleted }}This plan was not saved because one or more projects failed and automerge requires all plans pass.{{ else }}* :arrow_forward: To **apply** this plan, comment:\n" +
	"    * `{{.ApplyCmd}}`\n" +
	"* :put_litter_in_its_place: To **delete** this plan click [here]({{.LockURL}})\n" +
	"* :repeat: To **plan** this project again, comment:\n" +
	"    * `{{.RePlanCmd}}`{{end}}"
var applyUnwrappedSuccessTmpl = template.Must(template.New("").Parse(
	"```diff\n" +
		"{{.Output}}\n" +
		"```"))
var applyWrappedSuccessTmpl = template.Must(template.New("").Parse(
	"<details><summary>Show Output</summary>\n\n" +
		"```diff\n" +
		"{{.Output}}\n" +
		"```\n" +
		"</details>"))
var unwrappedErrTmplText = "**{{.Command}} Error**\n" +
	"```\n" +
	"{{.Error}}\n" +
	"```"
var wrappedErrTmplText = "**{{.Command}} Error**\n" +
	"<details><summary>Show Output</summary>\n\n" +
	"```\n" +
	"{{.Error}}\n" +
	"```\n</details>"
var unwrappedErrTmpl = template.Must(template.New("").Parse(unwrappedErrTmplText))
var unwrappedErrWithLogTmpl = template.Must(template.New("").Parse(unwrappedErrTmplText + logTmpl))
var wrappedErrTmpl = template.Must(template.New("").Parse(wrappedErrTmplText))
var failureTmplText = "**{{.Command}} Failed**: {{.Failure}}"
var failureTmpl = template.Must(template.New("").Parse(failureTmplText))
var failureWithLogTmpl = template.Must(template.New("").Parse(failureTmplText + logTmpl))
var logTmpl = "{{if .Verbose}}\n<details><summary>Log</summary>\n  <p>\n\n```\n{{.Log}}```\n</p></details>{{end}}\n"
