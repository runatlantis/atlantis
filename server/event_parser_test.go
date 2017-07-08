package server_test

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
)

func TestDetermineCommandInvalid(t *testing.T) {
	t.Log("given a comment that does not match the regex should return an error")
	e := server.EventParser{"user"}
	comments := []string{
		// just the executable, no command
		"run",
		"atlantis",
		"@user",
		// invalid command
		"run slkjd",
		"atlantis slkjd",
		"@user slkjd",
		"atlantis plans",
		// whitespace
		"  atlantis plan",
		// misc
		"related comment mentioning atlantis",
	}
	for _, c := range comments {
		_, e := e.DetermineCommand(buildComment(c))
		Assert(t, e != nil, "expected error for comment: "+c)
	}
}

func TestDetermineCommandHelp(t *testing.T) {
	t.Log("given a help comment, should match")
	e := server.EventParser{"user"}
	comments := []string{
		"run help",
		"atlantis help",
		"@user help",
		"atlantis help --verbose",
	}
	for _, c := range comments {
		command, e := e.DetermineCommand(buildComment(c))
		Ok(t, e)
		Equals(t, server.Help, command.Name)
	}
}

func TestDetermineCommandPermutations(t *testing.T) {
	e := server.EventParser{"user"}

	execNames := []string{"run", "atlantis", "@user"}
	commandNames := []server.CommandName{server.Plan, server.Apply}
	envs := []string{"", "default", "env", "env-dash", "env_underscore", "camelEnv"}
	verboses := []bool{true, false}

	// test all permutations
	for _, exec := range execNames {
		for _, name := range commandNames {
			for _, env := range envs {
				for _, v := range verboses {
					vFlag := ""
					if v == true {
						vFlag = "--verbose"
					}

					comment := fmt.Sprintf("%s %s %s %s", exec, name.String(), env, vFlag)
					t.Log("testing comment: " + comment)
					c, err := e.DetermineCommand(buildComment(comment))
					Ok(t, err)
					Equals(t, name, c.Name)
					if env == "" {
						Equals(t, "default", c.Environment)
					} else {
						Equals(t, env, c.Environment)
					}
					Equals(t, v, c.Verbose)
				}
			}
		}
	}
}

func buildComment(c string) *github.IssueCommentEvent {
	return &github.IssueCommentEvent{
		Comment: &github.IssueComment{
			Body: github.String(c),
		},
	}
}
