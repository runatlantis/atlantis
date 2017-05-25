package main

import (
	"io/ioutil"
	"os"
	"testing"
)

var tempConfigFile = "/tmp/" + AtlantisConfigFile

func TestConfigFileExists_invalid_path(t *testing.T) {
	var c Config
	equals(t, c.Exists("/invalid/path"), false)
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
	equals(t, c.Exists("/tmp"), true)
}

func TestConfigFileRead_invalid_config(t *testing.T) {
	var c Config
	str := []byte(`---invalid`)
	writeAtlantisConfigFile(str)
	defer os.Remove(tempConfigFile)
	err := c.Read("/tmp")
	assert(t, err != nil, "expect an error")
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
	assert(t, err == nil, "should be valid yaml")
}

func writeAtlantisConfigFile(s []byte) error {
	return ioutil.WriteFile(tempConfigFile, s, 0644)
}
