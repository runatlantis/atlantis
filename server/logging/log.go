// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"io"
	"log"
)

// SuppressDefaultLogging suppresses the default logging
func SuppressDefaultLogging() {
	// Some packages use the default logger, so we need to suppress it. (such as uber-go/tally)
	log.SetOutput(io.Discard)
}
