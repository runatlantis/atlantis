// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package events

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/i18n"
)

var (
	planCommandTitle            = command.Plan.String()
	applyCommandTitle           = command.Apply.String()
	policyCheckCommandTitle     = command.PolicyCheck.String()
	approvePoliciesCommandTitle = command.ApprovePolicies.String()
	versionCommandTitle         = command.Version.String()
	importCommandTitle          = command.Import.String()
	stateCommandTitle           = command.State.String()
	// maxUnwrappedLines is the maximum number of lines the Terraform output
	// can be before we wrap it in an expandable template.
	maxUnwrappedLines = 12

	//go:embed templates/*.tmpl templates/i18n/*/*.tmpl
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
	translator                *i18n.Translator
}

// commonData is data that all responses have.
type commonData struct {
	Command                   string
	CommandName               string
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
	localizationConfigs ...i18n.TranslatorConfig,
) *MarkdownRenderer {
	localizationConfig := i18n.TranslatorConfig{LanguageCode: i18n.DefaultLanguage}
	if len(localizationConfigs) > 0 {
		localizationConfig = localizationConfigs[0]
	}

	languageCode := i18n.DefaultLanguage
	customLanguageConfigFile := localizationConfig.CatalogPath
	if localizationConfig.LanguageCode != "" {
		languageCode = localizationConfig.LanguageCode
	}
	translator := i18n.NewTranslatorOrDefault(i18n.TranslatorConfig{
		LanguageCode: languageCode,
		CatalogPath:  customLanguageConfigFile,
	})

	var templates *template.Template
	templates, _ = template.New("").Funcs(sprig.TxtFuncMap()).ParseFS(templatesFS, "templates/*.tmpl")
	localizedPattern := fmt.Sprintf("templates/i18n/%s/*.tmpl", translator.LanguageCode())
	if localizedTemplates, err := fs.Glob(templatesFS, localizedPattern); err == nil && len(localizedTemplates) > 0 {
		if localized, localizedErr := templates.ParseFS(templatesFS, localizedPattern); localizedErr == nil {
			templates = localized
		}
	}
	if overrides, err := templates.ParseGlob(fmt.Sprintf("%s/*.tmpl", markdownTemplateOverridesDir)); err == nil {
		// doesn't override if templates directory doesn't exist
		templates = overrides
	}
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
		translator:                translator,
	}
}

// Render formats the data into a markdown string.
// nolint: interfacer
func (m *MarkdownRenderer) Render(ctx *command.Context, res command.Result, cmd PullCommand) string {
	commandNameStr := cmd.CommandName().String()
	commandStr := m.translator.CommandTitle(commandNameStr)
	var vcsRequestType string
	if ctx.Pull.BaseRepo.VCSHost.Type == models.Gitlab {
		vcsRequestType = m.translator.MergeRequestLabel()
	} else {
		vcsRequestType = m.translator.PullRequestLabel()
	}

	common := commonData{
		Command:                   commandStr,
		CommandName:               commandNameStr,
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
		} else if result.PolicyCheckResults != nil && common.CommandName == policyCheckCommandTitle {
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
		} else if result.PolicyCheckResults != nil && common.CommandName == approvePoliciesCommandTitle {
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
		} else if result.Error == nil && result.Failure == "" {
			resultData.Rendered = "Found no template. This is a bug!"
		}
		// Render error or failure templates. Done outside of previous block so that other context can be rendered for use here.
		if result.Error != nil {
			tmpl := templates.Lookup("unwrappedErr")
			if m.shouldUseWrappedTmpl(vcsHost, result.Error.Error()) {
				tmpl = templates.Lookup("wrappedErr")
			}
			resultData.Rendered = m.renderTemplateTrimSpace(tmpl, errData{result.Error.Error(), resultData.Rendered, common})
			if common.CommandName == applyCommandTitle {
				numApplyErrors++
			}
		} else if result.Failure != "" {
			resultData.Rendered = m.renderTemplateTrimSpace(templates.Lookup("failure"), failureData{result.Failure, resultData.Rendered, common})
			if common.CommandName == applyCommandTitle {
				numApplyFailures++
			}
		}
		resultsTmplData = append(resultsTmplData, resultData)
	}

	var tmpl *template.Template
	switch {
	case len(resultsTmplData) == 1 && common.CommandName == planCommandTitle && numPlanSuccesses > 0:
		tmpl = templates.Lookup("singleProjectPlanSuccess")
	case len(resultsTmplData) == 1 && common.CommandName == planCommandTitle && numPlanSuccesses == 0:
		tmpl = templates.Lookup("singleProjectPlanUnsuccessful")
	case len(resultsTmplData) == 1 && common.CommandName == policyCheckCommandTitle && numPolicyCheckSuccesses > 0:
		tmpl = templates.Lookup("singleProjectPlanSuccess")
	case len(resultsTmplData) == 1 && common.CommandName == policyCheckCommandTitle && numPolicyCheckSuccesses == 0:
		tmpl = templates.Lookup("singleProjectPolicyUnsuccessful")
	case len(resultsTmplData) == 1 && common.CommandName == versionCommandTitle && numVersionSuccesses > 0:
		tmpl = templates.Lookup("singleProjectVersionSuccess")
	case len(resultsTmplData) == 1 && common.CommandName == versionCommandTitle && numVersionSuccesses == 0:
		tmpl = templates.Lookup("singleProjectVersionUnsuccessful")
	case len(resultsTmplData) == 1 && common.CommandName == applyCommandTitle:
		tmpl = templates.Lookup("singleProjectApply")
	case len(resultsTmplData) == 1 && common.CommandName == importCommandTitle:
		tmpl = templates.Lookup("singleProjectImport")
	case len(resultsTmplData) == 1 && common.CommandName == stateCommandTitle:
		switch common.SubCommand {
		case "rm":
			tmpl = templates.Lookup("singleProjectStateRm")
		default:
			return fmt.Sprintf("no template matched–this is a bug: command=%s, command_name=%s, subcommand=%s", common.Command, common.CommandName, common.SubCommand)
		}
	case common.CommandName == planCommandTitle:
		tmpl = templates.Lookup("multiProjectPlan")
	case common.CommandName == policyCheckCommandTitle:
		if numPolicyCheckSuccesses == len(results) {
			tmpl = templates.Lookup("multiProjectPolicy")
		} else {
			tmpl = templates.Lookup("multiProjectPolicyUnsuccessful")
		}
	case common.CommandName == approvePoliciesCommandTitle:
		if numPolicyApprovalSuccesses == len(results) {
			tmpl = templates.Lookup("approveAllProjects")
		} else {
			tmpl = templates.Lookup("multiProjectPolicyUnsuccessful")
		}
	case common.CommandName == applyCommandTitle:
		tmpl = templates.Lookup("multiProjectApply")
	case common.CommandName == versionCommandTitle:
		tmpl = templates.Lookup("multiProjectVersion")
	case common.CommandName == importCommandTitle:
		tmpl = templates.Lookup("multiProjectImport")
	case common.CommandName == stateCommandTitle:
		switch common.SubCommand {
		case "rm":
			tmpl = templates.Lookup("multiProjectStateRm")
		default:
			return fmt.Sprintf("no template matched–this is a bug: command=%s, command_name=%s, subcommand=%s", common.Command, common.CommandName, common.SubCommand)
		}
	default:
		return fmt.Sprintf("no template matched–this is a bug: command=%s, command_name=%s", common.Command, common.CommandName)
	}

	switch common.CommandName {
	case planCommandTitle:
		numPlanFailures := len(results) - numPlanSuccesses
		return m.renderTemplateTrimSpace(tmpl, planResultData{resultsTmplData, common, numPlansWithChanges, numPlansWithNoChanges, numPlanFailures})
	case applyCommandTitle:
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

func (m *MarkdownRenderer) renderTemplateTrimSpace(tmpl *template.Template, data any) string {
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		return fmt.Sprintf("Failed to render template, this is a bug: %v", err)
	}
	return strings.TrimSpace(buf.String())
}
