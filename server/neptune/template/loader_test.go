package template

import (
	"fmt"
	"os"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/stretchr/testify/assert"
)

var testRepo = models.Repo{
	VCSHost: models.VCSHost{
		Hostname: models.Github.String(),
	},
	FullName: "test-repo",
}

func TestLoader_TemplateOverride(t *testing.T) {
	globalCfg := valid.GlobalCfg{
		Repos: []valid.Repo{
			{
				ID: testRepo.ID(),
				TemplateOverrides: map[string]string{
					string(LegacyApplyComment): "testdata/custom.tmpl",
				},
			},
		},
	}

	loader := NewLoader[any](globalCfg)

	output, err := loader.Load(LegacyApplyComment, testRepo, nil)
	assert.NoError(t, err)

	templateContent, err := os.ReadFile(globalCfg.MatchingRepo(testRepo.ID()).TemplateOverrides[string(LegacyApplyComment)])
	assert.NoError(t, err)

	assert.Equal(t, output, string(templateContent))
}

func TestLoader_NoTemplateOverride(t *testing.T) {
	globalCfg := valid.GlobalCfg{
		Repos: []valid.Repo{
			{
				ID: testRepo.ID(),
			},
		},
	}

	loader := NewLoader[any](globalCfg)

	output, err := loader.Load(LegacyApplyComment, testRepo, nil)
	assert.NoError(t, err)

	templateContent, err := os.ReadFile(fmt.Sprintf("templates/%s.tmpl", LegacyApplyComment))
	assert.NoError(t, err)

	assert.Equal(t, output, string(templateContent))
}
