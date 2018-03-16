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
package events_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	. "github.com/runatlantis/atlantis/testing"
)

var tempConfigFile = "/tmp/" + events.ProjectConfigFile
var projectConfigFileStr = `
---
terraform_version: "0.0.1"
pre_init:
  commands:
  - "echo"
  - "pre_init"
pre_get:
  commands:
  - "echo"
  - "pre_get"
pre_plan:
  commands:
  - "echo"
  - "pre_plan"
post_plan:
  commands:
  - "echo"
  - "post_plan"
pre_apply:
  commands:
  - "echo"
  - "pre_apply"
post_apply:
  commands:
  - "echo"
  - "post_apply"
extra_arguments:
- command_name: "init"
  arguments: ["arg", "init"]
- command_name: "get"
  arguments: ["arg", "get"]
- command_name: "plan"
  arguments: ["arg", "plan"]
- command_name: "apply"
  arguments: ["arg", "apply"]
`

var c events.ProjectConfigManager

func TestExists_InvalidPath(t *testing.T) {
	t.Log("given a path to a directory that doesn't exist Exists should return false")
	Equals(t, c.Exists("/invalid/path"), false)
}

func TestExists_ValidPath(t *testing.T) {
	t.Log("given a path to a directory with an atlantis config file, Exists returns true")
	writeAtlantisConfigFile(t, []byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile) // nolint: errcheck
	Equals(t, c.Exists("/tmp"), true)
}

func TestRead_InvalidConfig(t *testing.T) {
	t.Log("when the config file has invalid yaml, we expect an error")
	str := []byte(`---invalid`)
	writeAtlantisConfigFile(t, str)
	defer os.Remove(tempConfigFile) // nolint: errcheck
	_, err := c.Read("/tmp")
	Assert(t, err != nil, "expect an error")
}

func TestRead_ValidConfig(t *testing.T) {
	t.Log("when the config file has valid yaml, it should be parsed")
	writeAtlantisConfigFile(t, []byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile) // nolint: errcheck
	config, err := c.Read("/tmp")
	Ok(t, err)
	Equals(t, []string{"echo", "pre_init"}, config.PreInit)
	Equals(t, []string{"echo", "pre_get"}, config.PreGet)
	Equals(t, []string{"echo", "pre_plan"}, config.PrePlan)
	Equals(t, []string{"echo", "post_plan"}, config.PostPlan)
	Equals(t, []string{"echo", "pre_apply"}, config.PreApply)
	Equals(t, []string{"echo", "post_apply"}, config.PostApply)
	Equals(t, []string{"arg", "init"}, config.GetExtraArguments("init"))
	Equals(t, []string{"arg", "get"}, config.GetExtraArguments("get"))
	Equals(t, []string{"arg", "plan"}, config.GetExtraArguments("plan"))
	Equals(t, []string{"arg", "apply"}, config.GetExtraArguments("apply"))
	Equals(t, 0, len(config.GetExtraArguments("not-specified")))
}

func writeAtlantisConfigFile(t *testing.T, s []byte) {
	err := ioutil.WriteFile(tempConfigFile, s, 0644)
	Ok(t, err)
}
