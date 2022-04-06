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
	"io/ioutil"
	"strings"
	"text/template"

	_ "embed"

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
)

var (
	planCommandTitle            = command.Plan.TitleString()
	applyCommandTitle           = command.Apply.TitleString()
	policyCheckCommandTitle     = command.PolicyCheck.TitleString()
	approvePoliciesCommandTitle = command.ApprovePolicies.TitleString()
	versionCommandTitle         = command.Version.TitleString()
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
	DisableApply             bool
	DisableMarkdownFolding   bool
	DisableRepoLocking       bool
	EnableDiffMarkdownFormat bool
}

// commonData is data that all responses have.
type commonData struct {
	Command                  string
	Verbose                  bool
	Log                      string
	DisableApplyAll          bool
	DisableApply             bool
	DisableRepoLocking       bool
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

type planSuccessData struct {
	models.PlanSuccess
	PlanSummary              string
	PlanWasDeleted           bool
	DisableApply             bool
	DisableRepoLocking       bool
	EnableDiffMarkdownFormat bool
}

type policyCheckSuccessData struct {
	models.PolicyCheckSuccess
}

type projectResultTmplData struct {
	Workspace   string
	RepoRelDir  string
	ProjectName string
	Rendered    string
}

// Render formats the data into a markdown string.
// nolint: interfacer
// TODO: remove support for verbose as this is deprecated, since history is not actually saved.
func (m *MarkdownRenderer) Render(res command.Result, cmdName command.Name, log string, verbose bool, vcsHost models.VCSHostType, templateOverrides map[string]string) string {
	commandStr := strings.Title(strings.Replace(cmdName.String(), "_", " ", -1))
	common := commonData{
		Command:                  commandStr,
		Verbose:                  verbose,
		Log:                      log,
		DisableApplyAll:          m.DisableApplyAll || m.DisableApply,
		DisableApply:             m.DisableApply,
		DisableRepoLocking:       m.DisableRepoLocking,
		EnableDiffMarkdownFormat: m.EnableDiffMarkdownFormat,
	}
	if res.Error != nil {
		return m.renderTemplate(template.Must(template.New("").Parse(unwrappedErrWithLogTmpl)), errData{res.Error.Error(), common})
	}
	if res.Failure != "" {
		return m.renderTemplate(template.Must(template.New("").Parse(failureWithLogTmpl)), failureData{res.Failure, common})
	}
	return m.renderProjectResults(res.ProjectResults, common, vcsHost, templateOverrides)
}

func (m *MarkdownRenderer) renderProjectResults(results []command.ProjectResult, common commonData, vcsHost models.VCSHostType, templateOverrides map[string]string) string {
	var resultsTmplData []projectResultTmplData
	numPlanSuccesses := 0
	numPolicyCheckSuccesses := 0
	numVersionSuccesses := 0

	for _, result := range results {
		resultData := projectResultTmplData{
			Workspace:   result.Workspace,
			RepoRelDir:  result.RepoRelDir,
			ProjectName: result.ProjectName,
		}
		if result.Error != nil {
			tmpl := m.getProjectErrTmpl(templateOverrides, vcsHost, result.Error.Error())
			resultData.Rendered = m.renderTemplate(tmpl, struct {
				Command string
				Error   string
			}{
				Command: common.Command,
				Error:   result.Error.Error(),
			})
		} else if result.Failure != "" {
			resultData.Rendered = m.renderTemplate(m.getProjectFailureTmpl(templateOverrides), struct {
				Command string
				Failure string
			}{
				Command: common.Command,
				Failure: result.Failure,
			})
		} else if result.PlanSuccess != nil {
			resultData.Rendered = m.renderTemplate(m.getProjectPlanSuccessTmpl(templateOverrides, vcsHost, result.PlanSuccess.TerraformOutput), planSuccessData{PlanSuccess: *result.PlanSuccess, PlanSummary: result.PlanSuccess.Summary(), DisableApply: common.DisableApply, DisableRepoLocking: common.DisableRepoLocking, EnableDiffMarkdownFormat: common.EnableDiffMarkdownFormat})
			numPlanSuccesses++
		} else if result.PolicyCheckSuccess != nil {
			resultData.Rendered = m.renderTemplate(m.getProjectPolicyCheckSuccessTmpl(templateOverrides, vcsHost, result.PolicyCheckSuccess.PolicyCheckOutput), policyCheckSuccessData{PolicyCheckSuccess: *result.PolicyCheckSuccess})
			numPolicyCheckSuccesses++
		} else if result.ApplySuccess != "" {
			resultData.Rendered = m.renderTemplate(m.getProjectApplySuccessTmpl(templateOverrides, vcsHost, result.ApplySuccess), struct{ Output string }{result.ApplySuccess})
		} else if result.VersionSuccess != "" {
			resultData.Rendered = m.renderTemplate(m.getProjectVersionSuccessTmpl(templateOverrides, vcsHost, result.VersionSuccess), struct{ Output string }{result.VersionSuccess})
			numVersionSuccesses++
		} else {
			resultData.Rendered = "Found no template. This is a bug!"
		}
		resultsTmplData = append(resultsTmplData, resultData)
	}

	var tmpl *template.Template
	switch {
	case common.Command == planCommandTitle,
		common.Command == policyCheckCommandTitle:
		tmpl = m.getPlanTmpl(templateOverrides, resultsTmplData, common, numPlanSuccesses, numPolicyCheckSuccesses)
	case common.Command == approvePoliciesCommandTitle:
		tmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(approveAllProjectsTmpl))
	case common.Command == applyCommandTitle:
		tmpl = m.getApplyTmpl(templateOverrides, resultsTmplData)
	case common.Command == versionCommandTitle:
		tmpl = m.getVersionTmpl(templateOverrides, resultsTmplData, common, numVersionSuccesses)
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

func (m *MarkdownRenderer) getProjectErrTmpl(templateOverrides map[string]string, vcsHost models.VCSHostType, output string) *template.Template {
	if val, ok := templateOverrides["project_err"]; ok {
		return template.Must(template.ParseFiles(val))
	} else if m.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(wrappedErrTmpl))
	} else {
		return template.Must(template.New("").Parse(unwrappedErrTmpl))
	}
}

func (m *MarkdownRenderer) getProjectFailureTmpl(templateOverrides map[string]string) *template.Template {
	if val, ok := templateOverrides["project_failure"]; ok {
		return template.Must(template.ParseFiles(val))
	}
	return template.Must(template.New("").Parse(failureTmpl))
}

func (m *MarkdownRenderer) getProjectPlanSuccessTmpl(templateOverrides map[string]string, vcsHost models.VCSHostType, output string) *template.Template {
	if val, ok := templateOverrides["project_plan_success"]; ok {
		return template.Must(template.ParseFiles(val))
	} else if m.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(planSuccessWrappedTmpl))
	} else {
		return template.Must(template.New("").Parse(planSuccessUnwrappedTmpl))
	}
}

func (m *MarkdownRenderer) getProjectPolicyCheckSuccessTmpl(templateOverrides map[string]string, vcsHost models.VCSHostType, output string) *template.Template {
	if val, ok := templateOverrides["project_policy_check_success"]; ok {
		return template.Must(template.ParseFiles(val))
	} else if m.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(policyCheckSuccessWrappedTmpl))
	} else {
		return template.Must(template.New("").Parse(policyCheckSuccessUnwrappedTmpl))
	}
}

func (m *MarkdownRenderer) getProjectApplySuccessTmpl(templateOverrides map[string]string, vcsHost models.VCSHostType, output string) *template.Template {
	if val, ok := templateOverrides["project_apply_success"]; ok {
		return template.Must(template.ParseFiles(val))
	} else if m.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(applyWrappedSuccessTmpl))
	} else {
		return template.Must(template.New("").Parse(applyUnwrappedSuccessTmpl))
	}
}

func (m *MarkdownRenderer) getProjectVersionSuccessTmpl(templateOverrides map[string]string, vcsHost models.VCSHostType, output string) *template.Template {
	if val, ok := templateOverrides["project_version_success"]; ok {
		return template.Must(template.ParseFiles(val))
	} else if m.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(versionWrappedSuccessTmpl))
	} else {
		return template.Must(template.New("").Parse(versionUnwrappedSuccessTmpl))
	}
}

func (m *MarkdownRenderer) getPlanTmpl(templateOverrides map[string]string, resultsTmplData []projectResultTmplData, common commonData, numPlanSuccesses int, numPolicyCheckSuccesses int) *template.Template {
	if file_name, ok := templateOverrides["plan"]; ok {
		if content, err := ioutil.ReadFile(file_name); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	switch {
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectPlanSuccessTmpl))
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses == 0:
		return template.Must(template.New("").Parse(singleProjectPlanUnsuccessfulTmpl))
	case len(resultsTmplData) == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectPlanSuccessTmpl))
	case len(resultsTmplData) == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses == 0:
		return template.Must(template.New("").Parse(singleProjectPlanUnsuccessfulTmpl))
	default:
		return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(multiProjectPlanTmpl))
	}
}

func (m *MarkdownRenderer) getApplyTmpl(templateOverrides map[string]string, resultsTmplData []projectResultTmplData) *template.Template {
	if file_name, ok := templateOverrides["apply"]; ok {
		if content, err := ioutil.ReadFile(file_name); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	if len(resultsTmplData) == 1 {
		return template.Must(template.New("").Parse(singleProjectApplyTmpl))
	} else {
		return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(multiProjectApplyTmpl))
	}
}

func (m *MarkdownRenderer) getVersionTmpl(templateOverrides map[string]string, resultsTmplData []projectResultTmplData, common commonData, numVersionSuccesses int) *template.Template {
	if file_name, ok := templateOverrides["version"]; ok {
		if content, err := ioutil.ReadFile(file_name); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	switch {
	case len(resultsTmplData) == 1 && common.Command == versionCommandTitle && numVersionSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectVersionSuccessTmpl))
	case len(resultsTmplData) == 1 && common.Command == versionCommandTitle && numVersionSuccesses == 0:
		return template.Must(template.New("").Parse(singleProjectVersionUnsuccessfulTmpl))
	default:
		return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(multiProjectVersionTmpl))
	}
}

//go:embed templates/singleProjectApply.tmpl
var singleProjectApplyTmpl string

//go:embed templates/singleProjectPlanSuccess.tmpl
var singleProjectPlanSuccessTmpl string

//go:embed templates/singleProjectPlanUnsuccessful.tmpl
var singleProjectPlanUnsuccessfulTmpl string

//go:embed templates/singleProjectVersionSuccess.tmpl
var singleProjectVersionSuccessTmpl string

//go:embed templates/singleProjectVersionUnsuccessful.tmpl
var singleProjectVersionUnsuccessfulTmpl string

//go:embed templates/approveAllProjects.tmpl
var approveAllProjectsTmpl string

//go:embed templates/multiProjectPlan.tmpl
var multiProjectPlanTmpl string

//go:embed templates/multiProjectApply.tmpl
var multiProjectApplyTmpl string

//go:embed templates/multiProjectApply.tmpl
var multiProjectVersionTmpl string

//go:embed templates/planSuccessUnwrapped.tmpl
var planSuccessUnwrappedTmpl string

//go:embed templates/planSuccessWrapped.tmpl
var planSuccessWrappedTmpl string

//go:embed templates/policyCheckSuccessUnwrapped.tmpl
var policyCheckSuccessUnwrappedTmpl string

//go:embed templates/policyCheckSuccessWrapped.tmpl
var policyCheckSuccessWrappedTmpl string

//go:embed templates/applyUnwrappedSuccess.tmpl
var applyUnwrappedSuccessTmpl string

//go:embed templates/applyWrappedSuccess.tmpl
var applyWrappedSuccessTmpl string

//go:embed templates/versionUnwrappedSuccess.tmpl
var versionUnwrappedSuccessTmpl string

//go:embed templates/versionWrappedSuccess.tmpl
var versionWrappedSuccessTmpl string

//go:embed templates/unwrappedErr.tmpl
var unwrappedErrTmpl string

//go:embed templates/unwrappedErrWithLog.tmpl
var unwrappedErrWithLogTmpl string

//go:embed templates/wrappedErr.tmpl
var wrappedErrTmpl string

//go:embed templates/failure.tmpl
var failureTmpl string

//go:embed templates/failureWithLog.tmpl
var failureWithLogTmpl string
