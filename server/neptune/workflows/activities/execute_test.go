package activities_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/workflows/activities"
	"github.com/stretchr/testify/assert"
)

func TestEnvVar(t *testing.T) {

	t.Run("prioritizes value", func(t *testing.T) {
		subject := activities.EnvVar{
			Name:  "some-name",
			Value: "some-val",
			Command: activities.StringCommand{
				Command: "echo 'hello'",
			},
		}

		v, err := subject.GetValue()

		assert.NoError(t, err)
		assert.Equal(t, "some-val", v)
	})

	t.Run("runs command when value dne", func(t *testing.T) {
		subject := activities.EnvVar{
			Name: "some-name",
			Command: activities.StringCommand{
				Command: "echo 'hello'",
			},
		}

		v, err := subject.GetValue()

		assert.NoError(t, err)
		assert.Equal(t, "hello", v)
	})
}
