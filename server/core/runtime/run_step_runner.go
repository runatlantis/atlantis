package runtime

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/terraform/ansi"
	"github.com/runatlantis/atlantis/server/jobs"
)

// RunStepRunner runs custom commands.
type RunStepRunner struct {
	TerraformExecutor TerraformExec
	DefaultTFVersion  *version.Version
	// TerraformBinDir is the directory where Atlantis downloads Terraform binaries.
	TerraformBinDir         string
	ProjectCmdOutputHandler jobs.ProjectCommandOutputHandler
}

func (r *RunStepRunner) Run(ctx command.ProjectContext, command string, path string, envs map[string]string) (string, error) {
	tfVersion := r.DefaultTFVersion
	if ctx.TerraformVersion != nil {
		tfVersion = ctx.TerraformVersion
	}

	err := r.TerraformExecutor.EnsureVersion(ctx.Log, tfVersion)
	if err != nil {
		err = fmt.Errorf("%s: Downloading terraform Version %s", err, tfVersion.String())
		ctx.Log.Debug("error: %s", err)
		return "", err
	}

	cmd := exec.Command("sh", "-c", command) // #nosec
	cmd.Dir = path

	baseEnvVars := os.Environ()
	customEnvVars := map[string]string{
		"ATLANTIS_TERRAFORM_VERSION": tfVersion.String(),
		"BASE_BRANCH_NAME":           ctx.Pull.BaseBranch,
		"BASE_REPO_NAME":             ctx.BaseRepo.Name,
		"BASE_REPO_OWNER":            ctx.BaseRepo.Owner,
		"COMMENT_ARGS":               strings.Join(ctx.EscapedCommentArgs, ","),
		"DIR":                        path,
		"HEAD_BRANCH_NAME":           ctx.Pull.HeadBranch,
		"HEAD_COMMIT":                ctx.Pull.HeadCommit,
		"HEAD_REPO_NAME":             ctx.HeadRepo.Name,
		"HEAD_REPO_OWNER":            ctx.HeadRepo.Owner,
		"PATH":                       fmt.Sprintf("%s:%s", os.Getenv("PATH"), r.TerraformBinDir),
		"PLANFILE":                   filepath.Join(path, GetPlanFilename(ctx.Workspace, ctx.ProjectName)),
		"SHOWFILE":                   filepath.Join(path, ctx.GetShowResultFileName()),
		"PROJECT_NAME":               ctx.ProjectName,
		"PULL_AUTHOR":                ctx.Pull.Author,
		"PULL_NUM":                   fmt.Sprintf("%d", ctx.Pull.Num),
		"REPO_REL_DIR":               ctx.RepoRelDir,
		"USER_NAME":                  ctx.User.Username,
		"WORKSPACE":                  ctx.Workspace,
	}

	finalEnvVars := baseEnvVars
	for key, val := range customEnvVars {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	for key, val := range envs {
		finalEnvVars = append(finalEnvVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = finalEnvVars
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("%s: unable to create stdout buffer", err)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		err = fmt.Errorf("%s: unable to create stderr buffer", err)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	if err := cmd.Start(); err != nil {
		err = fmt.Errorf("%s: unable to start command %q", err, command)
		ctx.Log.Debug("error: %s", err)
		return "", err
	}

	output := &bytes.Buffer{}
	mutex := &sync.Mutex{}

	go r.streamOutput(ctx, stdout, output, mutex)
	go r.streamOutput(ctx, stderr, output, mutex)

	err = cmd.Wait()

	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, command, path, output.String())
		ctx.Log.Debug("error: %s", err)
		return "", err
	}
	ctx.Log.Info("successfully ran %q in %q", command, path)
	return ansi.Strip(output.String()), nil
}

func (r RunStepRunner) streamOutput(ctx command.ProjectContext, reader io.Reader, buffer io.StringWriter, mutex *sync.Mutex) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		r.ProjectCmdOutputHandler.Send(ctx, line, false)
		mutex.Lock()
		buffer.WriteString(line)
		buffer.WriteString("\n")
		mutex.Unlock()
	}
}
