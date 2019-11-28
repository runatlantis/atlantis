package valid_test

import (
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

func TestAutoplan_SetEnabled(t *testing.T) {
	Equals(t, true, (DefaultValidAutoplan().SetEnabled(true)).Enabled)
	Equals(t, false, (DefaultValidAutoplan().SetEnabled(false)).Enabled)
}
