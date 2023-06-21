package runtime

import (
	"errors"
	"fmt"
	"strings"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/command"
)

// EnvStepRunner set environment variables.
type MultiEnvStepRunner struct {
	RunStepRunner *RunStepRunner
}

// Run runs the multienv step command.
// The command must return a json string containing the array of name-value pairs that are being added as extra environment variables
func (r *MultiEnvStepRunner) Run(ctx command.ProjectContext, command string, path string, envs map[string]string) (string, error) {
	res, err := r.RunStepRunner.Run(ctx, command, path, envs, false, valid.PostProcessRunOutputShow)
	if err != nil {
		return "", err
	}

	if len(res) == 0 {
		return "No dynamic environment variable added", nil
	}

	var sb strings.Builder
	sb.WriteString("Dynamic environment variables added:\n")

	envVars := strings.Split(res, ",")
	for _, item := range envVars {
		// Only split after the first = found in case the environment variable value has
		// = in it (as might be the case with access tokens)
		nameValue := strings.SplitN(strings.TrimRight(item, "\n"), "=", 2)
		if len(nameValue) == 2 {
			envs[nameValue[0]] = nameValue[1]
			sb.WriteString(nameValue[0])
			sb.WriteString("\n")
		} else {
			return "", fmt.Errorf("Invalid environment variable definition: %s", item)
		}
	}
	return sb.String(), nil
}

func parseMultienvLine(in string) ([]string, error) {
	in = strings.TrimSpace(in)
	if in == "" {
		return nil, nil
	}
	if len(in) < 3 {
		return nil, errors.New("invalid syntax") // TODO
	}

	var res []string
	var inValue, dquoted, squoted, escaped bool
	var i int

	for j, r := range in {
		if !inValue {
			if r == '=' {
				inValue = true
				res = append(res, in[i:j])
				i = j + 1
			}
			if r == ' ' || r == '\t' {
				return nil, errInvalidKeySyntax
			}
			if r == ',' && len(res) > 0 {
				i = j + 1
			}
			continue
		}

		if r == '"' && !squoted {
			if j == i && !dquoted { // value is double quoted
				dquoted = true
				i = j + 1
			} else if dquoted && in[j-1] != '\\' {
				res = append(res, unescape(in[i:j], escaped))
				i = j + 1
				dquoted = false
				inValue = false
			} else if in[j-1] != '\\' {
				return nil, errMisquoted
			} else if in[j-1] == '\\' {
				escaped = true
			}
			continue
		}

		if r == '\'' && !dquoted {
			if j == i && !squoted { // value is double quoted
				squoted = true
				i = j + 1
			} else if squoted && in[j-1] != '\\' {
				res = append(res, in[i:j])
				i = j + 1
				squoted = false
				inValue = false
			}
			continue
		}

		if r == ',' && !dquoted && !squoted && inValue {
			res = append(res, in[i:j])
			i = j + 1
			inValue = false
		}
	}

	if i < len(in) {
		if !inValue {
			return nil, errRemaining
		}
		res = append(res, unescape(in[i:], escaped))
		inValue = false
	}
	if dquoted || squoted {
		return nil, errMisquoted
	}
	if inValue {
		return nil, errRemaining
	}

	return res, nil
}

func unescape(s string, escaped bool) string {
	if escaped {
		return strings.ReplaceAll(strings.ReplaceAll(s, `\\`, `\`), `\"`, `"`)
	}
	return s
}

var (
	errInvalidKeySyntax = errors.New("invalid key syntax")
	errMisquoted        = errors.New("misquoted")
	errRemaining        = errors.New("remaining unparsed data")
)
