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

package markdown

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	_ "embed" // embedding files

	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Renderer renders responses as markdown.
type Renderer struct {
	TemplateResolver TemplateResolver

	DisableApplyAll          bool
	DisableApply             bool
	EnableDiffMarkdownFormat bool
}

// commonData is data that all responses have.
type commonData struct {
	Command                  string
	DisableApplyAll          bool
	DisableApply             bool
	EnableDiffMarkdownFormat bool
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

type projectResultTmplData struct {
	Workspace   string
	RepoRelDir  string
	ProjectName string
	Rendered    string
}

// Render formats the data into a markdown string for a command.
// nolint: interfacer
func (m *Renderer) Render(res command.Result, cmdName command.Name, baseRepo models.Repo) string {
	commandStr := cases.Title(language.English).String(strings.ReplaceAll(cmdName.String(), "_", " "))
	common := commonData{
		Command:                  commandStr,
		DisableApplyAll:          m.DisableApplyAll || m.DisableApply,
		DisableApply:             m.DisableApply,
		EnableDiffMarkdownFormat: m.EnableDiffMarkdownFormat,
	}
	if res.Error != nil {
		return m.renderTemplate(template.Must(template.New("").Parse(unwrappedErrWithLogTmpl)), errData{res.Error.Error(), common})
	}
	if res.Failure != "" {
		return m.renderTemplate(template.Must(template.New("").Parse(failureWithLogTmpl)), failureData{res.Failure, common})
	}

	return m.renderProjectResults(res.ProjectResults, common, cmdName, baseRepo)
}

// RenderProject formats the data into a markdown string for a project
func (m *Renderer) RenderProject(prjRes command.ProjectResult, cmdName fmt.Stringer, baseRepo models.Repo) string {
	commandStr := cases.Title(language.English).String(strings.ReplaceAll(cmdName.String(), "_", " "))
	common := commonData{
		Command:                  commandStr,
		DisableApply:             m.DisableApply,
		EnableDiffMarkdownFormat: m.EnableDiffMarkdownFormat,
	}
	template, templateData := m.TemplateResolver.ResolveProject(prjRes, baseRepo, common)
	return m.renderTemplate(template, templateData)
}

func (m *Renderer) renderProjectResults(results []command.ProjectResult, common commonData, cmdName command.Name, baseRepo models.Repo) string {
	// render project results
	var prjResultTmplData []projectResultTmplData
	for _, result := range results {
		prjResultTmplData = append(prjResultTmplData, projectResultTmplData{
			Workspace:   result.Workspace,
			RepoRelDir:  result.RepoRelDir,
			ProjectName: result.ProjectName,
			Rendered:    m.RenderProject(result, cmdName, baseRepo),
		})
	}

	// render aggregate operation result
	numPlanSuccesses, numPolicyCheckSuccesses, numVersionSuccesses := m.countSuccesses(results)
	tmpl := m.TemplateResolver.Resolve(common, baseRepo, len(results), numPlanSuccesses, numPolicyCheckSuccesses, numVersionSuccesses)
	if tmpl == nil {
		return "no template matched–this is a bug"
	}
	return m.renderTemplate(tmpl, resultData{prjResultTmplData, common})
}

func (m *Renderer) countSuccesses(results []command.ProjectResult) (numPlanSuccesses, numPolicyCheckSuccesses, numVersionSuccesses int) {
	for _, result := range results {
		switch {
		case result.PlanSuccess != nil:
			numPlanSuccesses++
		case result.PolicyCheckSuccess != nil:
			numPolicyCheckSuccesses++
		case result.VersionSuccess != "":
			numVersionSuccesses++
		}
	}
	return
}

func (m *Renderer) renderTemplate(tmpl *template.Template, data interface{}) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return buf.String()
}
