package terraform_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/neptune/terraform"
	"github.com/stretchr/testify/assert"
)

func TestCommandArguments_Build(t *testing.T) {
	t.Run("with flags", func(t *testing.T) {
		c := terraform.NewSubCommand(terraform.Show)

		c.WithFlags(terraform.Flag{
			Value: "json",
		})

		assert.Equal(t, []string{"show", "-json"}, c.Build())
	})

	t.Run("with input", func(t *testing.T) {
		c := terraform.NewSubCommand(terraform.Apply)

		c.WithInput("input.tfplan")

		assert.Equal(t, []string{"apply", "input.tfplan"}, c.Build())
	})

	t.Run("with args", func(t *testing.T) {
		c := terraform.NewSubCommand(terraform.Init)

		c.WithArgs(terraform.Argument{
			Key:   "input",
			Value: "false",
		})

		assert.Equal(t, []string{"init", "-input=false"}, c.Build())
	})

	t.Run("dedups last first", func(t *testing.T) {
		c := terraform.NewSubCommand(terraform.Init)

		c.WithArgs(
			terraform.Argument{
				Key:   "input",
				Value: "false",
			},
			terraform.Argument{
				Key:   "a",
				Value: "b",
			},
			terraform.Argument{
				Key:   "input",
				Value: "true",
			},
			terraform.Argument{
				Key:   "c",
				Value: "d",
			},
		)

		assert.Equal(t, []string{"init", "-a=b", "-c=d", "-input=true"}, c.Build())
	})
}
