package events

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/runatlantis/atlantis/server/events/yaml"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/stretchr/testify/require"
)

func TestRepoBranch(t *testing.T) {
	globalYAML := `repos:
  - id: gitlab.com/nico.lab/strokeviewer/main
    branch: /release/.*/
    apply_requirements: [approved, mergeable]
    allowed_overrides: [workflow]
    allowed_workflows: [test, production]
    allow_custom_workflows: true

workflows:
  test:
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
  - name: development/anton
    branch: /release/.*/
    dir: terraform/development/anton
    workflow: test
    autoplan:
      when_modified:
        - "../../../build/build.json"
        - "../../../build/_env/*"
        - "../../../build/_settings/*"
        - "../../_modules/**/*"
        - "**/*"
  - name: test/nl
    branch: /master/
    dir: terraform/test/nl
    workflow: production
    autoplan:
      when_modified:
        - "../../../build/build.json"
        - "../../../build/_env/*"
        - "../../../build/_settings/*"
        - "../../_modules/**/*"
        - "**/*"
`

	tmp, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer func() {
		os.RemoveAll(tmp)
	}()

	globalYAMLPath := filepath.Join(tmp, "config.yaml")
	err = ioutil.WriteFile(globalYAMLPath, []byte(globalYAML), 0600)
	require.NoError(t, err)

	globalCfgArgs := valid.GlobalCfgArgs{
		AllowRepoCfg:  false,
		MergeableReq:  false,
		ApprovedReq:   false,
		UnDivergedReq: false,
	}

	parser := &yaml.ParserValidator{}
	global, err := parser.ParseGlobalCfg(globalYAMLPath, valid.NewGlobalCfgFromArgs(globalCfgArgs))
	require.NoError(t, err)

	repoYAMLPath := filepath.Join(tmp, "atlantis.yaml")
	err = ioutil.WriteFile(repoYAMLPath, []byte(repoYAML), 0600)
	require.NoError(t, err)

	repo, err := parser.ParseRepoCfg(tmp, global, "gitlab.com/nico.lab/strokeviewer/main", "release/nico-1040-atlantis-test")
	require.NoError(t, err)

	require.Equal(t, 1, len(repo.Projects))

	t.Logf("Projects: %+v", repo.Projects)
}
