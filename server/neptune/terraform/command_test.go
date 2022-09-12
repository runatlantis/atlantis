package terraform_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/stretchr/testify/assert"
)

func TestCommandArguments_Build(t *testing.T) {
	t.Run("empty extra args", func(t *testing.T) {
		c, err := terraform.NewCommandArguments(terraform.Init, []string{"-input=false"}, []string{})
		assert.Nil(t, err)

		assert.Equal(t, []string{"init", "-input=false"}, c.Build())
	})

	t.Run("empty command args with extra args", func(t *testing.T) {
		c, err := terraform.NewCommandArguments(terraform.Init, []string{}, []string{"-input=false"})
		assert.Nil(t, err)

		assert.Equal(t, []string{"init", "-input=false"}, c.Build())
	})

	t.Run("empty command args and empty extra args", func(t *testing.T) {
		c, err := terraform.NewCommandArguments(terraform.Init, []string{}, []string{})
		assert.Nil(t, err)

		assert.Equal(t, []string{"init"}, c.Build())
	})

	t.Run("extra args replaces command args", func(t *testing.T) {
		c, err := terraform.NewCommandArguments(terraform.Init, []string{"-input=false", "-a=b"}, []string{"-input=true", "-c=d"})
		assert.Nil(t, err)

		assert.Equal(t, []string{"init", "-input=true", "-a=b", "-c=d"}, c.Build())
	})
}
