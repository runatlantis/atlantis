package run

import (
	"testing"

	version "github.com/hashicorp/go-version"
	"github.com/hootsuite/atlantis/server/logging"
	. "github.com/hootsuite/atlantis/testing"
)

var logger = logging.NewNoopLogger()
var run = &Run{}

func TestRunCreateScript_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	scriptName, err := createScript(cmds, "post_apply")
	Assert(t, scriptName != "", "there should be a script name")
	Assert(t, err == nil, "there should not be an error")
}

func TestRunExecuteScript_invalid(t *testing.T) {
	cmds := []string{"invalid", "command"}
	scriptName, _ := createScript(cmds, "post_apply")
	_, err := execute(scriptName)
	Assert(t, err != nil, "there should be an error")
}

func TestRunExecuteScript_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	scriptName, _ := createScript(cmds, "post_apply")
	output, err := execute(scriptName)
	Assert(t, err == nil, "there should not be an error")
	Assert(t, output != "", "there should be output")
}

func TestRun_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	version, _ := version.NewVersion("0.8.8")
	_, err := run.Execute(logger, cmds, "/tmp/atlantis", "staging", version, "post_apply")
	Ok(t, err)
}
