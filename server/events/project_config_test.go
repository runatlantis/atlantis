package events_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/hootsuite/atlantis/server/events"
	. "github.com/hootsuite/atlantis/testing_util"
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
	writeAtlantisConfigFile([]byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile)
	Equals(t, c.Exists("/tmp"), true)
}

func TestRead_InvalidConfig(t *testing.T) {
	t.Log("when the config file has invalid yaml, we expect an error")
	str := []byte(`---invalid`)
	writeAtlantisConfigFile(str)
	defer os.Remove(tempConfigFile)
	_, err := c.Read("/tmp")
	Assert(t, err != nil, "expect an error")
}

func TestRead_ValidConfig(t *testing.T) {
	t.Log("when the config file has valid yaml, it should be parsed")
	writeAtlantisConfigFile([]byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile)
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

func writeAtlantisConfigFile(s []byte) error {
	return ioutil.WriteFile(tempConfigFile, s, 0644)
}
