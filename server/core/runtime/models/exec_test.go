// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package models_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/runtime/models"
	. "github.com/runatlantis/atlantis/testing"
)

func TestLocalExec_CombinedOutput_EmptyArgs(t *testing.T) {
	e := models.LocalExec{}
	_, err := e.CombinedOutput([]string{}, nil, "")
	Assert(t, err != nil, "expected error for empty args")
	Equals(t, "no command specified", err.Error())
}

func TestLocalExec_CombinedOutput_RunsCommand(t *testing.T) {
	e := models.LocalExec{}
	out, err := e.CombinedOutput([]string{"echo", "hello"}, nil, "")
	Ok(t, err)
	Equals(t, "hello\n", out)
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
