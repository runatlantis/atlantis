// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package logging_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/logging"
	"github.com/stretchr/testify/assert"
)

func TestStructuredLoggerSavesHistory(t *testing.T) {
	logger := logging.NewNoopLogger(t)

	historyLogger := logger.WithHistory()

	expectedStr := "[DBUG] Hello World\n[INFO] foo bar\n"

	historyLogger.Debug("Hello World")
	historyLogger.Info("foo bar")

	assert.Equal(t, expectedStr, historyLogger.GetHistory())
}
