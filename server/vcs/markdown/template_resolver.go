package markdown

import (
	"os"
	"strings"
	"text/template"

	"github.com/runatlantis/atlantis/server/events/terraform/filter"

	_ "embed" // embedding files

	"github.com/Masterminds/sprig/v3"
	"github.com/runatlantis/atlantis/server/core/config/valid"
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

type planSuccessData struct {
	models.PlanSuccess
	PlanSummary              string
	PlanWasDeleted           bool
	DisableApply             bool
	EnableDiffMarkdownFormat bool
}

type policyCheckSuccessData struct {
	models.PolicyCheckSuccess
}

type ProjectOutputType int

const (
	Failure ProjectOutputType = iota
	Error
	PlanSuccess
	PolicyCheckSuccess
	ApplySuccess
	VersionSuccess
)

func (p ProjectOutputType) String() string {
	switch p {
	case Failure:
		return "project_failure"
	case Error:
		return "project_err"
	case PlanSuccess:
		return "project_plan_success"
	case PolicyCheckSuccess:
		return "project_policy_check_success"
	case ApplySuccess:
		return "project_apply_success"
	case VersionSuccess:
		return "project_version_success"
	}
	return ""
}

// Uses template overrides and server configs to resolve template
type TemplateResolver struct {
	// GitlabSupportsCommonMark is true if the version of GitLab we're
	// using supports the CommonMark markdown format.
	// If we're not configured with a GitLab client, this will be false.
	GitlabSupportsCommonMark bool
	DisableMarkdownFolding   bool
	GlobalCfg                valid.GlobalCfg
	LogFilter                filter.LogFilter
}

// Resolves templates for commands
func (t *TemplateResolver) Resolve(common commonData, baseRepo models.Repo, numPrjResults int, numPlanSuccesses int, numPolicyCheckSuccesses int, numVersionSuccesses int) *template.Template {
	// Build template override for this repo
	var templateOverrides map[string]string
	repoCfg := t.GlobalCfg.MatchingRepo(baseRepo.ID())
	if repoCfg != nil {
		templateOverrides = repoCfg.TemplateOverrides
	}

	var tmpl *template.Template
	switch {
	case common.Command == approvePoliciesCommandTitle:
		tmpl = template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(approveAllProjectsTmpl))
	case common.Command == planCommandTitle, common.Command == policyCheckCommandTitle:
		tmpl = t.getPlanTmpl(common, baseRepo, templateOverrides, numPrjResults, numPlanSuccesses, numPolicyCheckSuccesses)
	case common.Command == applyCommandTitle:
		tmpl = t.getApplyTmpl(templateOverrides, numPrjResults)
	case common.Command == versionCommandTitle:
		tmpl = t.getVersionTmpl(templateOverrides, common, numPrjResults, numVersionSuccesses)
	}

	return tmpl
}

// Resolves templates for project commands
func (t *TemplateResolver) ResolveProject(result command.ProjectResult, baseRepo models.Repo, common commonData) (*template.Template, interface{}) {

	// Build template override for this repo
	var templateOverrides map[string]string
	repoCfg := t.GlobalCfg.MatchingRepo(baseRepo.ID())
	if repoCfg != nil {
		templateOverrides = repoCfg.TemplateOverrides
	}

	var tmpl *template.Template
	var templateData interface{}

	switch {
	case result.Error != nil:
		filteredOutput := t.filterOutput(result.Error.Error())
		tmpl = t.buildTemplate(Error, baseRepo.VCSHost.Type, wrappedErrTmpl, unwrappedErrTmpl, filteredOutput, templateOverrides)
		templateData = struct {
			Command string
			Error   string
		}{
			Command: common.Command,
			Error:   filteredOutput,
		}
	case result.Failure != "":
		// use template override if specified
		if val, ok := templateOverrides["project_failure"]; ok {
			tmpl = template.Must(template.ParseFiles(val))
		} else {
			tmpl = template.Must(template.New("").Parse(failureTmpl))
		}

		templateData = struct {
			Command string
			Failure string
		}{
			Command: common.Command,
			Failure: result.Failure,
		}
	case result.PlanSuccess != nil:
		filteredOutput := t.filterOutput(result.PlanSuccess.TerraformOutput)
		tmpl = t.buildTemplate(PlanSuccess, baseRepo.VCSHost.Type, planSuccessWrappedTmpl, planSuccessUnwrappedTmpl, filteredOutput, templateOverrides)
		result.PlanSuccess.TerraformOutput = filteredOutput
		templateData = planSuccessData{
			PlanSuccess:              *result.PlanSuccess,
			PlanSummary:              result.PlanSuccess.Summary(),
			DisableApply:             common.DisableApply,
			EnableDiffMarkdownFormat: common.EnableDiffMarkdownFormat,
		}
	case result.PolicyCheckSuccess != nil:
		tmpl = t.buildTemplate(PolicyCheckSuccess, baseRepo.VCSHost.Type, policyCheckSuccessWrappedTmpl, policyCheckSuccessUnwrappedTmpl, result.PolicyCheckSuccess.PolicyCheckOutput, templateOverrides)
		templateData = policyCheckSuccessData{
			PolicyCheckSuccess: *result.PolicyCheckSuccess,
		}
	case result.ApplySuccess != "":
		filteredOutput := t.filterOutput(result.ApplySuccess)
		tmpl = t.buildTemplate(ApplySuccess, baseRepo.VCSHost.Type, applyWrappedSuccessTmpl, applyUnwrappedSuccessTmpl, filteredOutput, templateOverrides)
		templateData = struct {
			Output string
		}{
			Output: filteredOutput,
		}
	case result.VersionSuccess != "":
		tmpl = t.buildTemplate(VersionSuccess, baseRepo.VCSHost.Type, versionWrappedSuccessTmpl, versionUnwrappedSuccessTmpl, result.VersionSuccess, templateOverrides)
		templateData = struct {
			Output string
		}{
			Output: result.VersionSuccess,
		}
	}
	return tmpl, templateData

}

func (t *TemplateResolver) filterOutput(s string) string {
	var filteredLines []string
	logLines := strings.Split(s, "\n")
	for _, line := range logLines {
		if t.LogFilter.ShouldFilterLine(line) {
			continue
		}
		filteredLines = append(filteredLines, line)
	}
	return strings.Join(filteredLines, "\n")
}

func (t *TemplateResolver) buildTemplate(projectOutputType ProjectOutputType, vcsHost models.VCSHostType, wrappedTmpl string, unwrappedTmpl string, output string, templateOverrides map[string]string) *template.Template {
	// use template override is specified
	if val, ok := templateOverrides[projectOutputType.String()]; ok {
		return template.Must(template.ParseFiles(val))
	} else if t.shouldUseWrappedTmpl(vcsHost, output) {
		return template.Must(template.New("").Parse(wrappedTmpl))
	} else {
		return template.Must(template.New("").Parse(unwrappedTmpl))
	}
}

// shouldUseWrappedTmpl returns true if we should use the wrapped markdown
// templates that collapse the output to make the comment smaller on initial
// load. Some VCS providers or versions of VCS providers don't support this
// syntax.
func (t *TemplateResolver) shouldUseWrappedTmpl(vcsHost models.VCSHostType, output string) bool {
	if t.DisableMarkdownFolding {
		return false
	}

	// Bitbucket Cloud and Server don't support the folding markdown syntax.
	if vcsHost == models.BitbucketServer || vcsHost == models.BitbucketCloud {
		return false
	}

	if vcsHost == models.Gitlab && !t.GitlabSupportsCommonMark {
		return false
	}

	return strings.Count(output, "\n") > maxUnwrappedLines
}

func (t *TemplateResolver) getPlanTmpl(common commonData, baseRepo models.Repo, templateOverrides map[string]string, numPrjResults int, numPlanSuccesses int, numPolicyCheckSuccesses int) *template.Template {
	if fileName, ok := templateOverrides["plan"]; ok {
		if content, err := os.ReadFile(fileName); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	switch {
	case numPrjResults == 1 && common.Command == planCommandTitle && numPlanSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectPlanSuccessTmpl))
	case numPrjResults == 1 && common.Command == planCommandTitle && numPlanSuccesses == 0:
		return template.Must(template.New("").Parse(singleProjectPlanUnsuccessfulTmpl))
	case numPrjResults == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectPlanSuccessTmpl))
	case numPrjResults == 1 && common.Command == policyCheckCommandTitle && numPolicyCheckSuccesses == 0:
		return template.Must(template.New("").Parse(singleProjectPlanUnsuccessfulTmpl))
	default:
		return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(multiProjectPlanTmpl))
	}
}

func (t *TemplateResolver) getApplyTmpl(templateOverrides map[string]string, numPrjResults int) *template.Template {
	if fileName, ok := templateOverrides["apply"]; ok {
		if content, err := os.ReadFile(fileName); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	if numPrjResults == 1 {
		return template.Must(template.New("").Parse(singleProjectApplyTmpl))
	}
	return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(multiProjectApplyTmpl))
}

func (t *TemplateResolver) getVersionTmpl(templateOverrides map[string]string, common commonData, numPrjResults int, numVersionSuccesses int) *template.Template {
	if fileName, ok := templateOverrides["version"]; ok {
		if content, err := os.ReadFile(fileName); err == nil {
			return template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(string(content)))
		}
	}
	switch {
	case numPrjResults == 1 && common.Command == versionCommandTitle && numVersionSuccesses > 0:
		return template.Must(template.New("").Parse(singleProjectVersionSuccessTmpl))
	case numPrjResults == 1 && common.Command == versionCommandTitle && numVersionSuccesses == 0:
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
