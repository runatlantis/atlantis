package server

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/hootsuite/atlantis/testing_util"
)

var tempConfigFile = "/tmp/" + ProjectConfigFile
var projectConfigFileStr = `
---
terraform_version: "0.0.1"
post_apply:
  commands:
    - "echo"
    - "date"
pre_apply:
  commands:
    - "echo"
    - "date"
pre_plan:
  commands:
    - "echo"
    - "date"
extra_arguments:
  - command_name: "plan"
    arguments: ["-var", "hello=world"]
`

func TestConfigFileExists_invalid_path(t *testing.T) {
	var c ConfigReader
	Equals(t, c.Exists("/invalid/path"), false)
}

func TestConfigFileExists_valid_path(t *testing.T) {
	var c ConfigReader
	writeAtlantisConfigFile([]byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile)
	Equals(t, c.Exists("/tmp"), true)
}

func TestConfigFileRead_invalid_config(t *testing.T) {
	var c ConfigReader
	str := []byte(`---invalid`)
	writeAtlantisConfigFile(str)
	defer os.Remove(tempConfigFile)
	_, err := c.Read("/tmp")
	Assert(t, err != nil, "expect an error")
}

func TestConfigFileRead_valid_config(t *testing.T) {
	var c ConfigReader
	writeAtlantisConfigFile([]byte(projectConfigFileStr))
	defer os.Remove(tempConfigFile)
	_, err := c.Read("/tmp")
	Assert(t, err == nil, "should be valid yaml")
}

func writeAtlantisConfigFile(s []byte) error {
	return ioutil.WriteFile(tempConfigFile, s, 0644)
}
