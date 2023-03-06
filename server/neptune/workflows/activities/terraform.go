package activities

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	key "github.com/runatlantis/atlantis/server/neptune/context"

	"github.com/hashicorp/go-version"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/neptune/logger"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/file"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/temporal"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

// TerraformClientError can be used to assert a non-retryable error type for
// callers of this activity
type TerraformClientError struct {
	err error
}

func (e TerraformClientError) Error() string {
	return e.err.Error()
}

func NewTerraformClientError(err error) *TerraformClientError {
	return &TerraformClientError{
		err: err,
	}
}

func wrapTerraformError(err error, message string) TerraformClientError {
	// double wrap here to get specifics + error type for temporal to not retry
	return TerraformClientError{
		err: errors.Wrap(err, message),
	}
}

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

// Setting the buffer size to 10mb
const bufioScannerBufferSize = 10 * 1024 * 1024

type TerraformClient interface {
	RunCommand(ctx context.Context, request *terraform.RunCommandRequest, options ...terraform.RunOptions) error
}

type streamer interface {
	RegisterJob(id string) chan string
}

type gitCredentialsRefresher interface {
	Refresh(ctx context.Context, token int64) error
}

type terraformActivities struct {
	TerraformClient        TerraformClient
	DefaultTFVersion       *version.Version
	StreamHandler          streamer
	GHAppConfig            githubapp.Config
	GitCLICredentials      gitCredentialsRefresher
	GitCredentialsFileLock *file.RWLock
}

func NewTerraformActivities(
	client TerraformClient,
	defaultTfVersion *version.Version,
	streamHandler streamer,
	gitCredentialsRefresher gitCredentialsRefresher,
	gitCredentialsFileLock *file.RWLock,
) *terraformActivities { //nolint:revive // avoiding refactor while adding linter action
	return &terraformActivities{
		TerraformClient:        client,
		DefaultTFVersion:       defaultTfVersion,
		StreamHandler:          streamHandler,
		GitCLICredentials:      gitCredentialsRefresher,
		GitCredentialsFileLock: gitCredentialsFileLock,
	}
}

func getEnvs(directEnvs map[string]string, dynamicEnvs []EnvVar) (map[string]string, error) {
	envs := make(map[string]string)

	for k, v := range directEnvs {
		envs[k] = v
	}

	for _, e := range dynamicEnvs {
		v, err := e.GetValue()

		if err != nil {
			return envs, errors.Wrap(err, fmt.Sprintf("loading dynamic env var with name %s", e.Name))
		}

		envs[e.Name] = v
	}

	return envs, nil
}

// Terraform Init
type TerraformInitRequest struct {
	Args []terraform.Argument
	// deprecated: Use DynamicEnvs instead
	// remove once we are sure this isn't used
	Envs                 map[string]string
	DynamicEnvs          []EnvVar
	JobID                string
	TfVersion            string
	Path                 string
	GithubInstallationID int64
}

type TerraformInitResponse struct {
	Output string
}

func (t *terraformActivities) TerraformInit(ctx context.Context, request TerraformInitRequest) (TerraformInitResponse, error) {
	cancel := temporal.StartHeartbeat(ctx, temporal.HeartbeatTimeout)
	defer cancel()

	// Resolve the tf version to be used for this operation
	tfVersion, err := t.resolveVersion(request.TfVersion)
	if err != nil {
		return TerraformInitResponse{}, err
	}

	args := []terraform.Argument{
		DisableInputArg,
	}
	args = append(args, request.Args...)

	envs, err := getEnvs(request.Envs, request.DynamicEnvs)

	if err != nil {
		return TerraformInitResponse{}, err
	}

	r := &terraform.RunCommandRequest{
		RootPath:          request.Path,
		SubCommand:        terraform.NewSubCommand(terraform.Init).WithArgs(args...),
		AdditionalEnvVars: envs,
		Version:           tfVersion,
	}

	err = t.GitCLICredentials.Refresh(ctx, request.GithubInstallationID)
	if err != nil {
		logger.Warn(ctx, "Error refreshing git cli credentials. This is bug and will likely cause fetching of private modules to fail", key.ErrKey, err)
	}

	// terraform init clones repos using git cli auth of which we chose git global configs.
	// let's ensure we are locking access to this file so it's not rewritten to during the duration of our
	// operation
	t.GitCredentialsFileLock.RLock()
	defer t.GitCredentialsFileLock.RUnlock()

	out, err := t.runCommandWithOutputStream(ctx, request.JobID, r)
	if err != nil {
		logger.Error(ctx, out)
		return TerraformInitResponse{}, wrapTerraformError(err, "running init command")
	}
	return TerraformInitResponse{}, nil
}

// Terraform Plan

type TerraformPlanRequest struct {
	Args []terraform.Argument
	// deprecated: Use DynamicEnvs instead
	// remove once we are sure this isn't used
	Envs        map[string]string
	DynamicEnvs []EnvVar
	JobID       string
	TfVersion   string
	Path        string
	Mode        *terraform.PlanMode
}

type TerraformPlanResponse struct {
	PlanFile string
	Output   string
	Summary  terraform.PlanSummary
}

func (t *terraformActivities) TerraformPlan(ctx context.Context, request TerraformPlanRequest) (TerraformPlanResponse, error) {
	cancel := temporal.StartHeartbeat(ctx, temporal.HeartbeatTimeout)
	defer cancel()
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
	var flags []terraform.Flag

	if request.Mode != nil {
		flags = append(flags, request.Mode.ToFlag())
	}

	envs, err := getEnvs(request.Envs, request.DynamicEnvs)

	if err != nil {
		return TerraformPlanResponse{}, err
	}

	planRequest := &terraform.RunCommandRequest{
		RootPath:          request.Path,
		SubCommand:        terraform.NewSubCommand(terraform.Plan).WithArgs(args...).WithFlags(flags...),
		AdditionalEnvVars: envs,
		Version:           tfVersion,
	}
	out, err := t.runCommandWithOutputStream(ctx, request.JobID, planRequest)

	if err != nil {
		logger.Error(ctx, out)
		return TerraformPlanResponse{}, wrapTerraformError(err, "running plan command")
	}

	// let's run terraform show right after to get the plan as a structured object
	showRequest := &terraform.RunCommandRequest{
		RootPath: request.Path,
		SubCommand: terraform.NewSubCommand(terraform.Show).
			WithFlags(terraform.Flag{
				Value: "json",
			}).
			WithInput(planFile),
		AdditionalEnvVars: envs,
		Version:           tfVersion,
	}

	showResultBuffer := &bytes.Buffer{}
	err = t.TerraformClient.RunCommand(ctx, showRequest, terraform.RunOptions{
		StdOut: showResultBuffer,
		StdErr: showResultBuffer,
	})

	// we shouldn't fail our activity just because show failed. Summaries aren't that critical.
	if err != nil {
		logger.Error(ctx, "error with terraform show", key.ErrKey, err)
	}

	summary, err := terraform.NewPlanSummaryFromJSON(showResultBuffer.Bytes())

	if err != nil {
		logger.Error(ctx, "error building plan summary", key.ErrKey, err)
	}

	return TerraformPlanResponse{
		PlanFile: filepath.Join(request.Path, PlanOutputFile),
		Summary:  summary,
	}, nil
}

// Terraform Apply

type TerraformApplyRequest struct {
	Args []terraform.Argument
	// deprecated: Use DynamicEnvs instead
	// remove once we are sure this isn't used
	Envs        map[string]string
	DynamicEnvs []EnvVar
	JobID       string
	TfVersion   string
	Path        string
	PlanFile    string
}

type TerraformApplyResponse struct {
	Output string
}

func (t *terraformActivities) TerraformApply(ctx context.Context, request TerraformApplyRequest) (TerraformApplyResponse, error) {
	cancel := temporal.StartHeartbeat(ctx, temporal.HeartbeatTimeout)
	defer cancel()
	tfVersion, err := t.resolveVersion(request.TfVersion)
	if err != nil {
		return TerraformApplyResponse{}, err
	}

	planFile := request.PlanFile
	args := []terraform.Argument{DisableInputArg}
	args = append(args, request.Args...)

	envs, err := getEnvs(request.Envs, request.DynamicEnvs)

	if err != nil {
		return TerraformApplyResponse{}, err
	}

	applyRequest := &terraform.RunCommandRequest{
		RootPath:          request.Path,
		SubCommand:        terraform.NewSubCommand(terraform.Apply).WithInput(planFile).WithArgs(args...),
		AdditionalEnvVars: envs,
		Version:           tfVersion,
	}
	out, err := t.runCommandWithOutputStream(ctx, request.JobID, applyRequest)

	if err != nil {
		logger.Error(ctx, out)
		return TerraformApplyResponse{}, wrapTerraformError(err, "running apply command")
	}

	return TerraformApplyResponse{}, nil
}

func (t *terraformActivities) runCommandWithOutputStream(ctx context.Context, jobID string, request *terraform.RunCommandRequest) (string, error) {
	reader, writer := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	var err error
	go func() {
		defer wg.Done()
		defer func() {
			if e := writer.Close(); e != nil {
				logger.Error(ctx, "closing pipe writer", key.ErrKey, e)
			}
		}()
		err = t.TerraformClient.RunCommand(ctx, request, terraform.RunOptions{
			StdOut: writer,
			StdErr: writer,
		})
	}()

	s := bufio.NewScanner(reader)

	buf := []byte{}
	s.Buffer(buf, bufioScannerBufferSize)

	var output strings.Builder
	ch := t.StreamHandler.RegisterJob(jobID)
	for s.Scan() {
		_, err := output.WriteString(s.Text())
		if err != nil {
			logger.Warn(ctx, "unable to write tf output to buffer")
		}
		ch <- s.Text()
	}

	close(ch)

	wg.Wait()

	return output.String(), err
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
