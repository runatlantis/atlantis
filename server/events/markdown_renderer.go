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
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	planCommandTitle            = command.Plan.TitleString()
	applyCommandTitle           = command.Apply.TitleString()
	policyCheckCommandTitle     = command.PolicyCheck.TitleString()
	approvePoliciesCommandTitle = command.ApprovePolicies.TitleString()
	versionCommandTitle         = command.Version.TitleString()
	importCommandTitle          = command.Import.TitleString()
	stateCommandTitle           = command.State.TitleString()
	// maxUnwrappedLines is the maximum number of lines the Terraform output
	// can be before we wrap it in an expandable template.
	maxUnwrappedLines = 12

	//go:embed templates/*
	templatesFS embed.FS
)

// MarkdownRenderer renders responses as markdown.
type MarkdownRenderer struct {
	// gitlabSupportsCommonMark is true if the version of GitLab we're
	// using supports the CommonMark markdown format.
	// If we're not configured with a GitLab client, this will be false.
	gitlabSupportsCommonMark  bool
	disableApplyAll           bool
	disableApply              bool
	disableMarkdownFolding    bool
	disableRepoLocking        bool
	enableDiffMarkdownFormat  bool
	markdownTemplates         *template.Template
	executableName            string
	hideUnchangedPlanComments bool
	quietPolicyChecks         bool
	logger                    logging.SimpleLogging
}

// commonData is data that all responses have.
type commonData struct {
	Command                   string
	SubCommand                string
	Verbose                   bool
	Log                       string
	PlansDeleted              bool
	DisableApplyAll           bool
	DisableApply              bool
	DisableRepoLocking        bool
	EnableDiffMarkdownFormat  bool
	ExecutableName            string
	HideUnchangedPlanComments bool
	QuietPolicyChecks         bool
	VcsRequestType            string
}

// errData is data about an error response.
type errData struct {
	Error           string
	RenderedContext string
	commonData
}

// failureData is data about a failure response.
type failureData struct {
	Failure         string
	RenderedContext string
	commonData
}

type resultData struct {
	Results []projectResultTmplData
	commonData
}

type planResultData struct {
	Results []projectResultTmplData
	commonData
	NumPlansWithChanges   int
	NumPlansWithNoChanges int
	NumPlanFailures       int
}

type applyResultData struct {
	Results []projectResultTmplData
	commonData
	NumApplySuccesses int
	NumApplyFailures  int
	NumApplyErrors    int
}

type planSuccessData struct {
	models.PlanSuccess
	PlanSummary              string
	PlanWasDeleted           bool
	DisableApply             bool
	DisableRepoLocking       bool
	EnableDiffMarkdownFormat bool
	PlanStats                models.PlanSuccessStats
}

type policyCheckResultsData struct {
	models.PolicyCheckResults
	PreConftestOutput     string
	PostConftestOutput    string
	PolicyCheckSummary    string
	PolicyApprovalSummary string
	PolicyCleared         bool
	commonData
}

type projectResultTmplData struct {
	Workspace    string
	RepoRelDir   string
	ProjectName  string
	Rendered     string
	NoChanges    bool
	IsSuccessful bool
}

// Initialize templates
func NewMarkdownRenderer(
	gitlabSupportsCommonMark bool,
	disableApplyAll bool,
	disableApply bool,
	disableMarkdownFolding bool,
	disableRepoLocking bool,
	enableDiffMarkdownFormat bool,
	markdownTemplateOverridesDir string,
	executableName string,
	hideUnchangedPlanComments bool,
	quietPolicyChecks bool,
	logger logging.SimpleLogging,
) *MarkdownRenderer {
	templates := loadTemplatesWithLogging(markdownTemplateOverridesDir, logger)
	
	return &MarkdownRenderer{
		gitlabSupportsCommonMark:  gitlabSupportsCommonMark,
		disableApplyAll:           disableApplyAll,
		disableMarkdownFolding:    disableMarkdownFolding,
		disableApply:              disableApply,
		disableRepoLocking:        disableRepoLocking,
		enableDiffMarkdownFormat:  enableDiffMarkdownFormat,
		markdownTemplates:         templates,
		executableName:            executableName,
		hideUnchangedPlanComments: hideUnchangedPlanComments,
		quietPolicyChecks:         quietPolicyChecks,
		logger:                    logger,
	}
}

// loadTemplatesWithLogging loads templates with comprehensive diagnostic logging
func loadTemplatesWithLogging(markdownTemplateOverridesDir string, logger logging.SimpleLogging) *template.Template {
	// Log built-in template loading
	logger.Info("loading built-in markdown templates")
	templates, err := template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(templatesFS, "templates/*.tmpl")
	if err != nil {
		logger.Err("loading built-in templates: %v", err)
		// Return empty template set rather than panic
		return template.New("").Funcs(sprig.TxtFuncMap())
	}
	logger.Info("successfully loaded %d built-in templates", len(templates.Templates()))
	
	// Log template names for debugging
	var builtinTemplateNames []string
	for _, tmpl := range templates.Templates() {
		builtinTemplateNames = append(builtinTemplateNames, tmpl.Name())
	}
	logger.Debug("built-in template names: %v", builtinTemplateNames)

	// Log override directory check
	if markdownTemplateOverridesDir == "" {
		logger.Info("no template override directory specified")
		return templates
	}

	logger.Info("checking for template overrides in directory %q", markdownTemplateOverridesDir)
	
	// Check if directory exists
	if _, err := os.Stat(markdownTemplateOverridesDir); os.IsNotExist(err) {
		logger.Warn("template override directory does not exist %q", markdownTemplateOverridesDir)
		return templates
	} else if err != nil {
		logger.Err("accessing template override directory %q: %v", markdownTemplateOverridesDir, err)
		return templates
	}
	
	logger.Info("template override directory exists and is accessible %q", markdownTemplateOverridesDir)
	
	// List available template files
	globPattern := filepath.Join(markdownTemplateOverridesDir, "*.tmpl")
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		logger.Err("globbing template files with pattern %q: %v", globPattern, err)
		return templates
	}
	
	if len(matches) == 0 {
		logger.Warn("no template files found in override directory %q", markdownTemplateOverridesDir)
		return templates
	}
	
	logger.Info("found %d template files matching pattern %q: %v", len(matches), globPattern, matches)
	
	// Attempt to parse overrides
	logger.Info("attempting to parse template overrides from %q", globPattern)
	overrides, err := templates.ParseGlob(globPattern)
	if err != nil {
		logger.Err("parsing template overrides from %q: %v", globPattern, err)
		logger.Err("continuing with built-in templates only")
		return templates
	}
	
	logger.Info("successfully parsed template overrides, override template count: %d", len(overrides.Templates()))
	
	// Log template names for debugging
	var overrideTemplateNames []string
	for _, tmpl := range overrides.Templates() {
		overrideTemplateNames = append(overrideTemplateNames, tmpl.Name())
	}
	logger.Info("override template names: %v", overrideTemplateNames)
	
	// Apply overrides
	logger.Info("template overrides applied successfully")
	return overrides
}

// Render formats the data into a markdown string.
// nolint: interfacer
func (m *MarkdownRenderer) Render(ctx *command.Context, res command.Result, cmd PullCommand) string {
	commandStr := cases.Title(language.English).String(strings.Replace(cmd.CommandName().String(), "_", " ", -1))
	var vcsRequestType string
	if ctx.Pull.BaseRepo.VCSHost.Type == models.Gitlab {
		vcsRequestType = "Merge Request"
	} else {
		vcsRequestType = "Pull Request"
	}

	common := commonData{
		Command:                   commandStr,
		SubCommand:                cmd.SubCommandName(),
		Verbose:                   cmd.IsVerbose(),
		Log:                       ctx.Log.GetHistory(),
		PlansDeleted:              res.PlansDeleted,
		DisableApplyAll:           m.disableApplyAll || m.disableApply,
		DisableApply:              m.disableApply,
		DisableRepoLocking:        m.disableRepoLocking,
		EnableDiffMarkdownFormat:  m.enableDiffMarkdownFormat,
		ExecutableName:            m.executableName,
		HideUnchangedPlanComments: m.hideUnchangedPlanComments,
		QuietPolicyChecks:         m.quietPolicyChecks,
		VcsRequestType:            vcsRequestType,
	}

	templates := m.markdownTemplates

	if res.Error != nil {
		return m.renderTemplateTrimSpace(templates.Lookup("unwrappedErrWithLog"), errData{res.Error.Error(), "", common})
	}
	if res.Failure != "" {
		return m.renderTemplateTrimSpace(templates.Lookup("failureWithLog"), failureData{res.Failure, "", common})
	}
	return m.renderProjectResults(ctx, res.ProjectResults, common)
}

func (m *MarkdownRenderer) renderProjectResults(ctx *command.Context, results []command.ProjectResult, common commonData) string {
	vcsHost := ctx.Pull.BaseRepo.VCSHost.Type

	var resultsTmplData []projectResultTmplData
	numPlanSuccesses := 0
	numPolicyCheckSuccesses := 0
	numPolicyApprovalSuccesses := 0
	numVersionSuccesses := 0
	numPlansWithChanges := 0
	numPlansWithNoChanges := 0
	numApplySuccesses := 0
	numApplyFailures := 0
	numApplyErrors := 0

	templates := m.markdownTemplates

	for _, result := range results {
		resultData := projectResultTmplData{
			Workspace:    result.Workspace,
			RepoRelDir:   result.RepoRelDir,
			ProjectName:  result.ProjectName,
			IsSuccessful: result.IsSuccessful(),
		}
		if result.PlanSuccess != nil {
			result.PlanSuccess.TerraformOutput = strings.TrimSpace(result.PlanSuccess.TerraformOutput)
			data := planSuccessData{
				PlanSuccess:              *result.PlanSuccess,
				PlanWasDeleted:           common.PlansDeleted,
				DisableApply:             common.DisableApply,
				DisableRepoLocking:       common.DisableRepoLocking,
				EnableDiffMarkdownFormat: common.EnableDiffMarkdownFormat,
				PlanStats:                result.PlanSuccess.Stats(),
			}
			if m.shouldUseWrappedTmpl(vcsHost, result.PlanSuccess.TerraformOutput) {
				data.PlanSummary = result.PlanSuccess.Summary()
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("planSuccessWrapped"), data)
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("planSuccessUnwrapped"), data)
			}
			resultData.NoChanges = result.PlanSuccess.NoChanges()
			if result.PlanSuccess.NoChanges() {
				numPlansWithNoChanges++
			} else {
				numPlansWithChanges++
			}
			numPlanSuccesses++
		} else if result.PolicyCheckResults != nil && common.Command == policyCheckCommandTitle {
			policyCheckResults := policyCheckResultsData{
				PreConftestOutput:     result.PolicyCheckResults.PreConftestOutput,
				PostConftestOutput:    result.PolicyCheckResults.PostConftestOutput,
				PolicyCheckResults:    *result.PolicyCheckResults,
				PolicyCheckSummary:    result.PolicyCheckResults.Summary(),
				PolicyApprovalSummary: result.PolicyCheckResults.PolicySummary(),
				PolicyCleared:         result.PolicyCheckResults.PolicyCleared(),
				commonData:            common,
			}
			if m.shouldUseWrappedTmpl(vcsHost, result.PolicyCheckResults.CombinedOutput()) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("policyCheckResultsWrapped"), policyCheckResults)
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("policyCheckResultsUnwrapped"), policyCheckResults)
			}
			if result.Error == nil && result.Failure == "" {
				numPolicyCheckSuccesses++
			}
		} else if result.PolicyCheckResults != nil && common.Command == approvePoliciesCommandTitle {
			policyCheckResults := policyCheckResultsData{
				PolicyCheckResults:    *result.PolicyCheckResults,
				PolicyCheckSummary:    result.PolicyCheckResults.Summary(),
				PolicyApprovalSummary: result.PolicyCheckResults.PolicySummary(),
				PolicyCleared:         result.PolicyCheckResults.PolicyCleared(),
				commonData:            common,
			}
			if m.shouldUseWrappedTmpl(vcsHost, result.PolicyCheckResults.CombinedOutput()) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("policyCheckResultsWrapped"), policyCheckResults)
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("policyCheckResultsUnwrapped"), policyCheckResults)
			}
			if result.Error == nil && result.Failure == "" {
				numPolicyApprovalSuccesses++
			}
		} else if result.ApplySuccess != "" {
			output := strings.TrimSpace(result.ApplySuccess)
			if m.shouldUseWrappedTmpl(vcsHost, result.ApplySuccess) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("applyWrappedSuccess"), struct{ Output string }{output})
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("applyUnwrappedSuccess"), struct{ Output string }{output})
			}
			numApplySuccesses++
		} else if result.VersionSuccess != "" {
			output := strings.TrimSpace(result.VersionSuccess)
			if m.shouldUseWrappedTmpl(vcsHost, output) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("versionWrappedSuccess"), struct{ Output string }{output})
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("versionUnwrappedSuccess"), struct{ Output string }{output})
			}
			numVersionSuccesses++
		} else if result.ImportSuccess != nil {
			result.ImportSuccess.Output = strings.TrimSpace(result.ImportSuccess.Output)
			if m.shouldUseWrappedTmpl(vcsHost, result.ImportSuccess.Output) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("importSuccessWrapped"), result.ImportSuccess)
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("importSuccessUnwrapped"), result.ImportSuccess)
			}
		} else if result.StateRmSuccess != nil {
			result.StateRmSuccess.Output = strings.TrimSpace(result.StateRmSuccess.Output)
			if m.shouldUseWrappedTmpl(vcsHost, result.StateRmSuccess.Output) {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("stateRmSuccessWrapped"), result.StateRmSuccess)
			} else {
				resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("stateRmSuccessUnwrapped"), result.StateRmSuccess)
			}
			// Error out if no template was found, only if there are no errors or failures.
			// This is because some errors and failures rely on additional context rendered by templates, but not all errors or failures.
		} else if !(result.Error != nil || result.Failure != "") {
			resultData.Rendered = "Found no template. This is a bug!"
		}
		// Render error or failure templates. Done outside of previous block so that other context can be rendered for use here.
		if result.Error != nil {
			tmpl := templates.Lookup("unwrappedErr")
			if m.shouldUseWrappedTmpl(vcsHost, result.Error.Error()) {
				tmpl = templates.Lookup("wrappedErr")
			}
			resultData.Rendered = m.renderTemplateTrimSpace(tmpl, errData{result.Error.Error(), resultData.Rendered, common})
			if common.Command == applyCommandTitle {
				numApplyErrors++
			}
		} else if result.Failure != "" {
			resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("failure"), failureData{result.Failure, resultData.Rendered, common})
			if common.Command == applyCommandTitle {
				numApplyFailures++
			}
		}
		resultsTmplData = append(resultsTmplData, resultData)
	}

	var tmpl *template.Template
	switch {
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses > 0:
		tmpl = templates.Lookup("singleProjectPlanSuccess")
	case len(resultsTmplData) == 1 && common.Command == planCommandTitle && numPlanSuccesses == 0:
		tmpl = templates.Lookup("singleProjectPlanUnsuccessful")
	case len(resultsTmplData) == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses > 0:
		tmpl = templates.Lookup("singleProjectPlanSuccess")
	case len(resultsTmplData) == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses == 0:
		tmpl = templates.Lookup("singleProjectPolicyUnsuccessful")
	case len(resultsTmplData) == 1 && common.Command == versionCommandTitle && numVersionSuccesses > 0:
		tmpl = templates.Lookup("singleProjectVersionSuccess")
	case len(resultsTmplData) == 1 && common.Command == versionCommandTitle && numVersionSuccesses == 0:
		tmpl = templates.Lookup("singleProjectVersionUnsuccessful")
	case len(resultsTmplData) == 1 && common.Command == applyCommandTitle:
		tmpl = templates.Lookup("singleProjectApply")
	case len(resultsTmplData) == 1 && common.Command == importCommandTitle:
		tmpl = templates.Lookup("singleProjectImport")
	case len(resultsTmplData) == 1 && common.Command == stateCommandTitle:
		switch common.SubCommand {
		case "rm":
			tmpl = templates.Lookup("singleProjectStateRm")
		default:
			return fmt.Sprintf("no template matched–this is a bug: command=%s, subcommand=%s", common.Command, common.SubCommand)
		}
	case common.Command == planCommandTitle:
		tmpl = templates.Lookup("multiProjectPlan")
	case common.Command == policyCheckCommandTitle:
		if numPolicyCheckSuccesses == len(results) {
			tmpl = templates.Lookup("multiProjectPolicy")
		} else {
			tmpl = templates.Lookup("multiProjectPolicyUnsuccessful")
		}
	case common.Command == approvePoliciesCommandTitle:
		if numPolicyApprovalSuccesses == len(results) {
			tmpl = templates.Lookup("approveAllProjects")
		} else {
			tmpl = templates.Lookup("multiProjectPolicyUnsuccessful")
		}
	case common.Command == applyCommandTitle:
		tmpl = templates.Lookup("multiProjectApply")
	case common.Command == versionCommandTitle:
		tmpl = templates.Lookup("multiProjectVersion")
	case common.Command == importCommandTitle:
		tmpl = templates.Lookup("multiProjectImport")
	case common.Command == stateCommandTitle:
		switch common.SubCommand {
		case "rm":
			tmpl = templates.Lookup("multiProjectStateRm")
		default:
			return fmt.Sprintf("no template matched–this is a bug: command=%s, subcommand=%s", common.Command, common.SubCommand)
		}
	default:
		return fmt.Sprintf("no template matched–this is a bug: command=%s", common.Command)
	}

	switch {
	case common.Command == planCommandTitle:
		numPlanFailures := len(results) - numPlanSuccesses
		return m.renderTemplateTrimSpace(tmpl, planResultData{resultsTmplData, common, numPlansWithChanges, numPlansWithNoChanges, numPlanFailures})
	case common.Command == applyCommandTitle:
		return m.renderTemplateTrimSpace(tmpl, applyResultData{resultsTmplData, common, numApplySuccesses, numApplyFailures, numApplyErrors})
	}
	return m.renderTemplateTrimSpace(tmpl, resultData{resultsTmplData, common})
}

// shouldUseWrappedTmpl returns true if we should use the wrapped markdown
// templates that collapse the output to make the comment smaller on initial
// load. Some VCS providers or versions of VCS providers don't support this
// syntax.
func (m *MarkdownRenderer) shouldUseWrappedTmpl(vcsHost models.VCSHostType, output string) bool {
	if m.disableMarkdownFolding {
		return false
	}

	// Bitbucket Cloud and Server don't support the folding markdown syntax.
	if vcsHost == models.BitbucketServer || vcsHost == models.BitbucketCloud {
		return false
	}

	if vcsHost == models.Gitlab && !m.gitlabSupportsCommonMark {
		return false
	}

	return strings.Count(output, "\n") > maxUnwrappedLines
}

func (m *MarkdownRenderer) renderTemplateTrimSpace(tmpl *template.Template, data interface{}) string {
	if tmpl == nil {
		m.logger.Err("template is nil, template not found in template set")
		// List all available templates for debugging
		var availableTemplates []string
		for _, t := range m.markdownTemplates.Templates() {
			availableTemplates = append(availableTemplates, t.Name())
		}
		m.logger.Err("available templates: %v", availableTemplates)
		return "template not found, this indicates a template override issue"
	}
	
	templateName := tmpl.Name()
	m.logger.Debug("executing template %q", templateName)
	
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		m.logger.Err("executing template %q: %v", templateName, err)
		return fmt.Sprintf("rendering template %q: %v", templateName, err)
	}
	
	result := strings.TrimSpace(buf.String())
	m.logger.Debug("successfully executed template %q, output length: %d characters", templateName, len(result))
	
	return result
}
