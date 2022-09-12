package terraform

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type Operation int

const (
	Init Operation = iota
	Plan
	Apply
)

func (t Operation) String() string {
	switch t {
	case Init:
		return "init"
	case Plan:
		return "plan"
	case Apply:
		return "apply"
	}
	return ""
}

// argument is the key value pair passed into the terraform command
type argument struct {
	Key   string
	Value string
}

func (a argument) build() string {
	return fmt.Sprintf("-%s=%s", a.Key, a.Value)
}

func newArgumentList(args []string) ([]argument, error) {
	arguments := []argument{}
	for _, arg := range args {
		typedArgument, err := newArgument(arg)
		if err != nil {
			return []argument{}, errors.Wrap(err, "building argument list")
		}
		arguments = append(arguments, typedArgument)
	}
	return arguments, nil
}

func newArgument(arg string) (argument, error) {
	// Remove any forward dashes and spaces
	arg = strings.TrimLeft(arg, "- ")
	coll := strings.Split(arg, "=")

	if len(coll) != 2 {
		return argument{}, errors.New(fmt.Sprintf("cannot parse argument: %s. argument can only have one =", arg))
	}

	return argument{
		Key:   coll[0],
		Value: coll[1],
	}, nil
}

type CommandArguments struct {
	Command     Operation
	CommandArgs []argument
	ExtraArgs   []argument
}

func NewCommandArguments(command Operation, commandArgs []string, extraArgs []string) (*CommandArguments, error) {
	extraArguments, err := newArgumentList(extraArgs)
	if err != nil {
		return nil, errors.Wrap(err, "parsing extra arguments")
	}

	commandArguments, err := newArgumentList(commandArgs)
	if err != nil {
		return nil, errors.Wrap(err, "parsing command arguments")
	}

	return &CommandArguments{
		Command:     command,
		ExtraArgs:   extraArguments,
		CommandArgs: commandArguments,
	}, nil
}

func (t CommandArguments) Build() []string {
	finalArgs := []string{}
	usedExtraArgsIndex := map[int]bool{}
	for _, arg := range t.CommandArgs {

		overrideIndex := -1
		for i, overrideArg := range t.ExtraArgs {
			if overrideArg.Key == arg.Key {
				usedExtraArgsIndex[i] = true
				overrideIndex = i
			}
		}

		if overrideIndex != -1 {
			finalArgs = append(finalArgs, t.ExtraArgs[overrideIndex].build())
		} else {
			finalArgs = append(finalArgs, arg.build())
		}
	}

	// Append unused extra arguments to final args
	for i, extraArg := range t.ExtraArgs {
		if _, ok := usedExtraArgsIndex[i]; !ok {
			finalArgs = append(finalArgs, extraArg.build())
		}
	}

	return append([]string{t.Command.String()}, finalArgs...)
}
