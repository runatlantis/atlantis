package metrics

import (
	"io"

	tally "github.com/uber-go/tally/v4"
)

// NewNoopScope creates a no-op metrics scope that doesn't log anything.
// This is useful for tests where metrics logging is not needed and would
// create excessive debug output.
func NewNoopScope() (tally.Scope, io.Closer, error) {
	scope := tally.NoopScope
	closer := nopCloser{}
	return scope, closer, nil
}

type nopCloser struct{}

func (nopCloser) Close() error {
	return nil
}