package events

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/core/config"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/stretchr/testify/require"
)

func TestRepoBranch(t *testing.T) {
	globalYAML := `repos:
  - id: github.com/foo/bar
    branch: /release/.*/
    apply_requirements: [approved, mergeable]
    allowed_overrides: [workflow]
    allowed_workflows: [development, production]
    allow_custom_workflows: true
workflows:
  development:
    plan:
      steps:
        - run: 'echo "Executing test workflow: terraform plan in ..."'
        - init:
            extra_args: ["-upgrade"]
        - plan
    apply:
      steps:
        - run: 'echo "Executing test workflow: terraform apply in ..."'
        - apply
  production:
    plan:
      steps:
        - run: 'echo "Executing production workflow: terraform plan in ..."'
        - init:
            extra_args: ["-upgrade"]
        - plan
    apply:
      steps:
        - run: 'echo "Executing production workflow: terraform apply in ..."'
        - apply
`

	repoYAML := `version: 3
projects:
  - name: development
    branch: /main/
    dir: terraform/development
    workflow: development
    autoplan:
      when_modified:
        - "**/*"
  - name: production
    branch: /production/
    dir: terraform/production
    workflow: production
    autoplan:
      when_modified:
        - "**/*"
`

	tmp := t.TempDir()

	globalYAMLPath := filepath.Join(tmp, "config.yaml")
	err := os.WriteFile(globalYAMLPath, []byte(globalYAML), 0600)
	require.NoError(t, err)

	globalCfgArgs := valid.GlobalCfgArgs{
		MergeableReq:  false,
		ApprovedReq:   false,
		UnDivergedReq: false,
	}

	parser := &config.ParserValidator{}
	global, err := parser.ParseGlobalCfg(globalYAMLPath, valid.NewGlobalCfgFromArgs(globalCfgArgs))
	require.NoError(t, err)

	repoYAMLPath := filepath.Join(tmp, "atlantis.yaml")
	err = os.WriteFile(repoYAMLPath, []byte(repoYAML), 0600)
	require.NoError(t, err)

	repo, err := parser.ParseRepoCfg(tmp, global, "github.com/foo/bar", "main")
	require.NoError(t, err)

	require.Equal(t, 1, len(repo.Projects))

	t.Logf("Projects: %+v", repo.Projects)
}
