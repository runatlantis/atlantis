package runtime

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// RunStep runs custom commands.
type RunStep struct {
	Commands []string
	Meta     StepMeta
}

func (r *RunStep) Run() (string, error) {
	if len(r.Commands) < 1 {
		return "", errors.New("no commands for run step")
	}
	path := r.Meta.AbsolutePath

	cmd := exec.Command("sh", "-c", strings.Join(r.Commands, " ")) // #nosec
	cmd.Dir = path
	out, err := cmd.CombinedOutput()

	commandStr := strings.Join(r.Commands, " ")
	if err != nil {
		err = fmt.Errorf("%s: running %q in %q: \n%s", err, commandStr, path, out)
		r.Meta.Log.Debug("error: %s", err)
		return string(out), err
	}
	r.Meta.Log.Info("successfully ran %q in %q", commandStr, path)
	return string(out), nil
}
