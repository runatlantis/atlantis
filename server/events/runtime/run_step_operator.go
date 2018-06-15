package runtime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
)

// RunStepOperator runs custom commands.
type RunStepOperator struct {
}

func (r *RunStepOperator) Run(ctx models.ProjectCommandContext, command []string, path string) (string, error) {
	if len(command) < 1 {
		return "", errors.New("no commands for run step")
	}

	cmd := exec.Command("sh", "-c", strings.Join(command, " ")) // #nosec
	cmd.Dir = path
	out, err := cmd.CombinedOutput()

	commandStr := strings.Join(command, " ")
	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, commandStr, path, out)
		ctx.Log.Debug("error: %s", err)
		return string(out), err
	}
	ctx.Log.Info("successfully ran %q in %q", commandStr, path)
	return string(out), nil
}
