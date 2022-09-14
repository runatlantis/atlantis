package terraform

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

type Operation string

const (
	Init  Operation = "init"
	Plan  Operation = "plan"
	Apply Operation = "apply"
	Show  Operation = "show"
)

// argument is the key value pair passed into the terraform command
type Argument struct {
	Key   string
	Value string
}

func (a Argument) build() string {
	return fmt.Sprintf("-%s=%s", a.Key, a.Value)
}

// takes in a list of key/value argument pairs and parses them.
// terraform arguments are expected to be in a certain form
// ie. "-input=false" where input and false are the key values respectively.
func NewArgumentList(args []string) ([]Argument, error) {
	arguments := []Argument{}
	for _, arg := range args {
		typedArgument, err := newArgument(arg)
		if err != nil {
			return []Argument{}, errors.Wrap(err, "building argument list")
		}
		arguments = append(arguments, typedArgument)
	}
	return arguments, nil
}

func newArgument(arg string) (Argument, error) {
	// Remove any forward dashes and spaces
	arg = strings.TrimLeft(arg, "- ")
	coll := strings.Split(arg, "=")

	if len(coll) != 2 {
		return Argument{}, errors.New(fmt.Sprintf("cannot parse argument: %s. argument can only have one =", arg))
	}

	return Argument{
		Key:   coll[0],
		Value: coll[1],
	}, nil
}

type CommandArguments struct {
	Command     Operation
	CommandArgs []Argument
	ExtraArgs   []Argument
}

func NewCommandArguments(command Operation, commandArgs []Argument, extraArgs ...Argument) (*CommandArguments, error) {

	return &CommandArguments{
		Command:     command,
		ExtraArgs:   extraArgs,
		CommandArgs: commandArgs,
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

	return append([]string{string(t.Command)}, finalArgs...)
}
