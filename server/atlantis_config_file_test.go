package server

import (
	"io/ioutil"
	"os"
	"testing"
	. "github.com/hootsuite/atlantis/testing_util"
)

var tempConfigFile = "/tmp/" + AtlantisConfigFile

func TestConfigFileExists_invalid_path(t *testing.T) {
	var c Config
	Equals(t, c.Exists("/invalid/path"), false)
}

func TestConfigFileExists_valid_path(t *testing.T) {
	var c Config
	var str = `
---
terraform_version: "0.0.1"
pre_apply:
  commands:
    - "echo"
    - "date"
pre_plan:
  commands:
    - "echo"
    - "date"
stash_path: "file/path"
`
	writeAtlantisConfigFile([]byte(str))
	defer os.Remove(tempConfigFile)
	Equals(t, c.Exists("/tmp"), true)
}

func TestConfigFileRead_invalid_config(t *testing.T) {
	var c Config
	str := []byte(`---invalid`)
	writeAtlantisConfigFile(str)
	defer os.Remove(tempConfigFile)
	err := c.Read("/tmp")
	Assert(t, err != nil, "expect an error")
}

func TestConfigFileRead_valid_config(t *testing.T) {
	var c Config
	var str = `
---
terraform_version: "0.0.1"
pre_apply:
  commands:
    - "echo"
    - "date"
pre_plan:
  commands:
    - "echo"
    - "date"
stash_path: "file/path"
`
	writeAtlantisConfigFile([]byte(str))
	defer os.Remove(tempConfigFile)
	err := c.Read("/tmp")
	Assert(t, err == nil, "should be valid json")
}

func writeAtlantisConfigFile(s []byte) error {
	return ioutil.WriteFile(tempConfigFile, s, 0644)
}
