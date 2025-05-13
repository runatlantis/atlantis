package raw_test

import (
	"io"
	"strings"

	"errors"

	yaml "github.com/goccy/go-yaml"
)

// Bool is a helper routine that allocates a new bool value
// to store v and returns a pointer to it.
func Bool(v bool) *bool { return &v }

// Int is a helper routine that allocates a new int value
// to store v and returns a pointer to it.
func Int(v int) *int { return &v }

// String is a helper routine that allocates a new string value
// to store v and returns a pointer to it.
func String(v string) *string { return &v }

// Helper function to unmarshal from strings
func unmarshalString(in string, out interface{}) error {
	decoder := yaml.NewDecoder(strings.NewReader(in), yaml.Strict())

	err := decoder.Decode(out)
	if errors.Is(err, io.EOF) {
		return nil
	}
	return err
}
