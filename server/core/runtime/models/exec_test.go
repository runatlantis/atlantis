// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestLocalExec_CombinedOutput_DoesNotInterpretArgsAsShell(t *testing.T) {
	// Arguments reaching conftest include a filename built from the project
	// name and workspace, both of which come from a repo-level atlantis.yaml.
	// They must be passed to the process as a single argument rather than
	// spliced into a shell command line.
	cases := []string{
		"a;id",
		"a&&id",
		"a|id",
		"a$(id)b",
		"a`id`b",
		"a>/dev/null",
	}
	for _, payload := range cases {
		t.Run(payload, func(t *testing.T) {
			e := models.LocalExec{}

			out, err := e.CombinedOutput([]string{"echo", payload}, nil, "")

			Ok(t, err)
			Equals(t, payload+"\n", out)
			Assert(t, !strings.Contains(out, "uid="),
				"payload %q was interpreted by a shell, output was %q", payload, out)
		})
	}
}

func TestLocalExec_CombinedOutput_ErrsOnEmptyArgs(t *testing.T) {
	e := models.LocalExec{}

	_, err := e.CombinedOutput(nil, nil, "")

	ErrEquals(t, "no command specified", err)
}

func TestLocalExec_CombinedOutput_PassesEnvVars(t *testing.T) {
	e := models.LocalExec{}

	out, err := e.CombinedOutput(
		[]string{"sh", "-c", "echo $MYVAR"},
		map[string]string{"MYVAR": "testvalue"},
		"",
	)

	Ok(t, err)
	Equals(t, "testvalue\n", out)
}

func TestLocalExec_CombinedOutput_ProcessEnvOverridesCustomEnv(t *testing.T) {
	t.Setenv("ATLANTIS_LOCAL_EXEC_PRECEDENCE", "process-value")
	e := models.LocalExec{}

	out, err := e.CombinedOutput(
		[]string{"sh", "-c", `printf %s "$ATLANTIS_LOCAL_EXEC_PRECEDENCE"`},
		map[string]string{"ATLANTIS_LOCAL_EXEC_PRECEDENCE": "workflow-value"},
		"",
	)

	Ok(t, err)
	Equals(t, "process-value", out)
}
