package server_test

import (
	"testing"

	"github.com/google/go-github/github"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing_util"
	"strings"
)

func TestDetermineCommandInvalid(t *testing.T) {
	t.Log("given a comment that does not match the regex should return an error")
	e := server.EventParser{"user", "token"}
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
	e := server.EventParser{"user", "token"}
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
	e := server.EventParser{"user", "token"}

	execNames := []string{"run", "atlantis", "@user"}
	commandNames := []server.CommandName{server.Plan, server.Apply}
	envs := []string{"", "default", "env", "env-dash", "env_underscore", "camelEnv"}
	flagCases := [][]string{
		[]string{},
		[]string{"--verbose"},
		[]string{"-key=value"},
		[]string{"-key", "value"},
		[]string{"-key1=value1", "-key2=value2"},
		[]string{"-key1=value1", "-key2", "value2"},
		[]string{"-key1", "value1", "-key2=value2"},
		[]string{"--verbose", "key2=value2"},
		[]string{"-key1=value1", "--verbose"},
	}

	// test all permutations
	for _, exec := range execNames {
		for _, name := range commandNames {
			for _, env := range envs {
				for _, flags := range flagCases {
					// If github comments end in a newline they get \r\n appended.
					// Ensure that we parse commands properly either way.
					for _, lineEnding := range []string{"", "\r\n"} {
						comment := strings.Join(append([]string{exec, name.String(), env}, flags...), " ") + lineEnding
						t.Log("testing comment: " + comment)
						c, err := e.DetermineCommand(buildComment(comment))
						Ok(t, err)
						Equals(t, name, c.Name)
						if env == "" {
							Equals(t, "default", c.Environment)
						} else {
							Equals(t, env, c.Environment)
						}
						Equals(t, stringInSlice("--verbose", flags), c.Verbose)

						// ensure --verbose never shows up in flags
						for _, f := range c.Flags {
							Assert(t, f != "--verbose", "Should not pass on the --verbose flag: %v", flags)
						}

						// check all flags are present
						for _, f := range flags {
							if f != "--verbose" {
								Contains(t, f, c.Flags)
							}
						}
					}
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

