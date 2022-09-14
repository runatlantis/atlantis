package activities

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/terraform/ansi"
	"github.com/runatlantis/atlantis/server/neptune/terraform"
)

var DisableInputArg = terraform.Argument{
	Key:   "input",
	Value: "false",
}

var RefreshArg = terraform.Argument{
	Key:   "refresh",
	Value: "true",
}

const (
	outArgKey      = "out"
	PlanOutputFile = "output.tfplan"
)

type TerraformClient interface {
	RunCommand(ctx context.Context, jobID string, path string, subcommand *terraform.SubCommand, customEnvVars map[string]string, v *version.Version) <-chan terraform.Line
}

type terraformActivities struct {
	TerraformClient  TerraformClient
	DefaultTFVersion *version.Version
}

func NewTerraformActivities(client TerraformClient, defaultTfVersion *version.Version) *terraformActivities {
	return &terraformActivities{
		TerraformClient:  client,
		DefaultTFVersion: defaultTfVersion,
	}
}

// Terraform Init
type TerraformInitRequest struct {
	Args      []terraform.Argument
	Envs      map[string]string
	JobID     string
	TfVersion string
	Path      string
}

type TerraformInitResponse struct {
	Output string
}

func (t *terraformActivities) TerraformInit(ctx context.Context, request TerraformInitRequest) (TerraformInitResponse, error) {
	// Resolve the tf version to be used for this operation
	tfVersion, err := t.resolveVersion(request.TfVersion)
	if err != nil {
		return TerraformInitResponse{}, err
	}

	args := []terraform.Argument{
		DisableInputArg,
	}
	args = append(args, request.Args...)
	cmd := terraform.NewSubCommand(terraform.Init).WithArgs(args...)

	ch := t.TerraformClient.RunCommand(ctx, request.JobID, request.Path, cmd, request.Envs, tfVersion)
	_, err = t.readCommandOutput(ch)
	if err != nil {
		return TerraformInitResponse{}, errors.Wrap(err, "processing command output")
	}
	return TerraformInitResponse{}, nil
}

// Terraform Plan
type TerraformPlanRequest struct {
	Args      []terraform.Argument
	Envs      map[string]string
	JobID     string
	TfVersion string
	Path      string
}

type TerraformPlanResponse struct {
	PlanFile string
	Output   string
}

func (t *terraformActivities) TerraformPlan(ctx context.Context, request TerraformPlanRequest) (TerraformPlanResponse, error) {
	tfVersion, err := t.resolveVersion(request.TfVersion)
	if err != nil {
		return TerraformPlanResponse{}, err
	}
	planFile := filepath.Join(request.Path, PlanOutputFile)

	args := []terraform.Argument{
		DisableInputArg,
		RefreshArg,
		{
			Key:   outArgKey,
			Value: planFile,
		},
	}
	args = append(args, request.Args...)

	cmd := terraform.NewSubCommand(terraform.Plan).WithArgs(args...)
	ch := t.TerraformClient.RunCommand(ctx, request.JobID, request.Path, cmd, request.Envs, tfVersion)
	_, err = t.readCommandOutput(ch)
	if err != nil {
		return TerraformPlanResponse{}, errors.Wrap(err, "processing command output")
	}

	return TerraformPlanResponse{
		PlanFile: planFile,
	}, nil
}

// Terraform Apply

type TerraformApplyRequest struct {
}

func (t *terraformActivities) TerraformApply(ctx context.Context, request TerraformApplyRequest) error {
	return nil
}

func (t *terraformActivities) resolveVersion(v string) (*version.Version, error) {
	// Use default version if configured version is empty
	if v == "" {
		return t.DefaultTFVersion, nil
	}

	version, err := version.NewVersion(v)
	if err != nil {
		return nil, errors.Wrap(err, "resolving terraform version")
	}

	if version != nil {
		return version, nil
	}
	return t.DefaultTFVersion, nil
}

func (t *terraformActivities) readCommandOutput(ch <-chan terraform.Line) (string, error) {
	var err error
	var lines []string
	for line := range ch {
		if line.Err != nil {
			err = errors.Wrap(line.Err, "executing command")
			break
		}
		lines = append(lines, line.Line)
	}
	if err != nil {
		return "", err
	}
	output := strings.Join(lines, "\n")
	// sanitize output by stripping out any ansi characters.
	output = ansi.Strip(output)
	return output, nil
}
