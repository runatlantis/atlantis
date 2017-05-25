package main

import (
	"log"
	"os"
	"testing"

	"github.com/hootsuite/atlantis/logging"
)

var level = logging.Info
var logger = &logging.SimpleLogger{
	Source: "server",
	Log:    log.New(os.Stderr, "", log.LstdFlags),
	Level:  level,
}

func TestPreRunCreateScript_empty(t *testing.T) {
	scriptName, err := createScript(nil)
	assert(t, scriptName == "", "there should not be a script name")
	assert(t, err == nil, "there should not be an error")
}

func TestPreRunCreateScript_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	scriptName, err := createScript(cmds)
	assert(t, scriptName != "", "there should be a script name")
	assert(t, err == nil, "there should not be an error")
}

func TestPreRunExecuteScript_invalid(t *testing.T) {
	cmds := []string{"invalid", "command"}
	scriptName, _ := createScript(cmds)
	_, err := execute(scriptName)
	assert(t, err != nil, "there should be an error")
}

func TestPreRunExecuteScript_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	scriptName, _ := createScript(cmds)
	output, err := execute(scriptName)
	assert(t, err == nil, "there should not be an error")
	assert(t, output != "", "there should be output")
}

func TestPreRun_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	prePlan := PrePlan{Commands: cmds}
	preApply := PreApply{Commands: cmds}
	var config Config
	config.PrePlan = prePlan
	config.PreApply = preApply
	config.StashPath = "/some/path"
	err := PreRun(&config, logger, "/some/path", &Command{environment: "staging", commandType: Plan})
	assert(t, err == nil, "should not error")

}

func TestPreRun_partial_valid(t *testing.T) {
	cmds := []string{"echo", "date"}
	prePlan := PrePlan{Commands: cmds}
	var config Config
	config.PrePlan = prePlan
	config.StashPath = "/some/path"
	err := PreRun(&config, logger, "/some/path", &Command{environment: "staging", commandType: Plan})
	assert(t, err == nil, "should not error")

}
