// Package raw contains the golang representations of the YAML elements
// supported in atlantis.yaml. The structs here represent the exact data that
// comes from the file before it is parsed/validated further.
package raw

import (
	"fmt"

	version "github.com/hashicorp/go-version"
)

// VersionValidator helper function to validate binary version.
// Function implements ozzo-validation::Rule.Validate interface.
func VersionValidator(value interface{}) error {
	strPtr := value.(*string)
	if strPtr == nil {
		return nil
	}
	_, err := version.NewVersion(*strPtr)
	return fmt.Errorf("version %q could not be parsed: %w", *strPtr, err)
}
