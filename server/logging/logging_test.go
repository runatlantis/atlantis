// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
