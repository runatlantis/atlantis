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
