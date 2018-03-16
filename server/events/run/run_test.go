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
//
package run

import (
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
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
	v, _ := version.NewVersion("0.8.8")
	_, err := run.Execute(logger, cmds, "/tmp/atlantis", "staging", v, "post_apply")
	Ok(t, err)
}
