package cmd_test

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/hootsuite/atlantis/cmd"
	"github.com/hootsuite/atlantis/server"
	. "github.com/hootsuite/atlantis/testing"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// passedConfig is set to whatever config ended up being passed to NewServer.
// Used for testing.
var passedConfig server.Config

type ServerCreatorMock struct{}

func (s *ServerCreatorMock) NewServer(config server.Config) (cmd.ServerStarter, error) {
	passedConfig = config
	return &ServerStarterMock{}, nil
}

type ServerStarterMock struct{}

func (s *ServerStarterMock) Start() error {
	return nil
}

func TestExecute_NoConfigFlag(t *testing.T) {
	t.Log("If there is no config flag specified Execute should return nil.")
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  "",
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Ok(t, err)
}

func TestExecute_ConfigFileExtension(t *testing.T) {
	t.Log("If the config file doesn't have an extension then error.")
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  "does-not-exist",
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Equals(t, "invalid config: reading does-not-exist: Unsupported Config Type \"\"", err.Error())
}

func TestExecute_ConfigFileMissing(t *testing.T) {
	t.Log("If the config file doesn't exist then error.")
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  "does-not-exist.yaml",
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Equals(t, "invalid config: reading does-not-exist.yaml: open does-not-exist.yaml: no such file or directory", err.Error())
}

func TestExecute_ConfigFileExists(t *testing.T) {
	t.Log("If the config file exists then there should be no error.")
	tmpFile := tempFile(t, "")
	defer os.Remove(tmpFile) // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  tmpFile,
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Ok(t, err)
}

func TestExecute_InvalidConfig(t *testing.T) {
	t.Log("If the config file contains invalid yaml there should be an error.")
	tmpFile := tempFile(t, "invalidyaml")
	defer os.Remove(tmpFile) // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  tmpFile,
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Assert(t, strings.Contains(err.Error(), "unmarshal errors"), "should be an unmarshal error")
}

func TestExecute_ValidateLogLevel(t *testing.T) {
	t.Log("Should validate log level.")
	c := setup(map[string]interface{}{
		cmd.LogLevelFlag: "invalid",
		cmd.GHUserFlag:   "user",
		cmd.GHTokenFlag:  "token",
	})
	err := c.Execute()
	Assert(t, err != nil, "should be an error")
	Equals(t, "invalid log level: not one of debug, info, warn, error", err.Error())
}

func TestExecute_ValidateVCSConfig(t *testing.T) {
	expErr := "--gh-user/--gh-token or --gitlab-user/--gitlab-token must be set"
	cases := []struct {
		description string
		flags       map[string]interface{}
		expectError bool
	}{
		{
			"no config set",
			nil,
			true,
		},
		{
			"just github token set",
			map[string]interface{}{
				cmd.GHTokenFlag: "token",
			},
			true,
		},
		{
			"just gitlab token set",
			map[string]interface{}{
				cmd.GitlabTokenFlag: "token",
			},
			true,
		},
		{
			"just github user set",
			map[string]interface{}{
				cmd.GHUserFlag: "user",
			},
			true,
		},
		{
			"just gitlab user set",
			map[string]interface{}{
				cmd.GitlabUserFlag: "user",
			},
			true,
		},
		{
			"github user and gitlab token set",
			map[string]interface{}{
				cmd.GHUserFlag:      "user",
				cmd.GitlabTokenFlag: "token",
			},
			true,
		},
		{
			"gitlab user and github token set",
			map[string]interface{}{
				cmd.GitlabUserFlag: "user",
				cmd.GHTokenFlag:    "token",
			},
			true,
		},
		{
			"github user and github token set and should be successful",
			map[string]interface{}{
				cmd.GHUserFlag:  "user",
				cmd.GHTokenFlag: "token",
			},
			false,
		},
		{
			"gitlab user and gitlab token set and should be successful",
			map[string]interface{}{
				cmd.GitlabUserFlag:  "user",
				cmd.GitlabTokenFlag: "token",
			},
			false,
		},
		{
			"github and gitlab user and github and gitlab token set and should be successful",
			map[string]interface{}{
				cmd.GHUserFlag:      "user",
				cmd.GHTokenFlag:     "token",
				cmd.GitlabUserFlag:  "user",
				cmd.GitlabTokenFlag: "token",
			},
			false,
		},
	}
	for _, testCase := range cases {
		t.Log("Should validate vcs config when " + testCase.description)
		c := setup(testCase.flags)
		err := c.Execute()
		if testCase.expectError {
			Assert(t, err != nil, "should be an error")
			Equals(t, expErr, err.Error())
		} else {
			Ok(t, err)
		}
	}
}

func TestExecute_Defaults(t *testing.T) {
	t.Log("Should set the defaults for all unspecified flags.")
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:      "user",
		cmd.GHTokenFlag:     "token",
		cmd.GitlabUserFlag:  "gitlab-user",
		cmd.GitlabTokenFlag: "gitlab-token",
	})
	err := c.Execute()
	Ok(t, err)

	Equals(t, "user", passedConfig.GithubUser)
	Equals(t, "token", passedConfig.GithubToken)
	Equals(t, "", passedConfig.GithubWebHookSecret)
	Equals(t, "gitlab-user", passedConfig.GitlabUser)
	Equals(t, "gitlab-token", passedConfig.GitlabToken)
	Equals(t, "", passedConfig.GitlabWebHookSecret)
	// Get our hostname since that's what gets defaulted to
	hostname, err := os.Hostname()
	Ok(t, err)
	Equals(t, "http://"+hostname+":4141", passedConfig.AtlantisURL)

	// Get our home dir since that's what gets defaulted to
	dataDir, err := homedir.Expand("~/.atlantis")
	Ok(t, err)
	Equals(t, dataDir, passedConfig.DataDir)
	Equals(t, "github.com", passedConfig.GithubHostname)
	Equals(t, "gitlab.com", passedConfig.GitlabHostname)
	Equals(t, "info", passedConfig.LogLevel)
	Equals(t, false, passedConfig.RequireApproval)
	Equals(t, 4141, passedConfig.Port)
}

func TestExecute_ExpandHomeDir(t *testing.T) {
	t.Log("Should expand the ~ in the home dir to the actual home dir.")
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
		cmd.DataDirFlag: "~/this/is/a/path",
	})
	err := c.Execute()
	Ok(t, err)

	home, err := homedir.Dir()
	Ok(t, err)
	Equals(t, home+"/this/is/a/path", passedConfig.DataDir)
}

func TestExecute_GithubUser(t *testing.T) {
	t.Log("Should remove the @ from the github username if it's passed.")
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:  "@user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Ok(t, err)

	Equals(t, "user", passedConfig.GithubUser)
}

func TestExecute_GitlabUser(t *testing.T) {
	t.Log("Should remove the @ from the gitlab username if it's passed.")
	c := setup(map[string]interface{}{
		cmd.GitlabUserFlag:  "@user",
		cmd.GitlabTokenFlag: "token",
	})
	err := c.Execute()
	Ok(t, err)

	Equals(t, "user", passedConfig.GitlabUser)
}

func TestExecute_Flags(t *testing.T) {
	t.Log("Should use all flags that are set.")
	c := setup(map[string]interface{}{
		cmd.AtlantisURLFlag:     "url",
		cmd.DataDirFlag:         "path",
		cmd.GHHostnameFlag:      "ghhostname",
		cmd.GHUserFlag:          "user",
		cmd.GHTokenFlag:         "token",
		cmd.GHWebHookSecret:     "secret",
		cmd.GitlabHostnameFlag:  "gitlab-hostname",
		cmd.GitlabUserFlag:      "gitlab-user",
		cmd.GitlabTokenFlag:     "gitlab-token",
		cmd.GitlabWebHookSecret: "gitlab-secret",
		cmd.LogLevelFlag:        "debug",
		cmd.PortFlag:            8181,
		cmd.RequireApprovalFlag: true,
	})
	err := c.Execute()
	Ok(t, err)

	Equals(t, "url", passedConfig.AtlantisURL)
	Equals(t, "path", passedConfig.DataDir)
	Equals(t, "ghhostname", passedConfig.GithubHostname)
	Equals(t, "user", passedConfig.GithubUser)
	Equals(t, "token", passedConfig.GithubToken)
	Equals(t, "secret", passedConfig.GithubWebHookSecret)
	Equals(t, "gitlab-hostname", passedConfig.GitlabHostname)
	Equals(t, "gitlab-user", passedConfig.GitlabUser)
	Equals(t, "gitlab-token", passedConfig.GitlabToken)
	Equals(t, "gitlab-secret", passedConfig.GitlabWebHookSecret)
	Equals(t, "debug", passedConfig.LogLevel)
	Equals(t, 8181, passedConfig.Port)
	Equals(t, true, passedConfig.RequireApproval)
}

func TestExecute_ConfigFile(t *testing.T) {
	t.Log("Should use all the values from the config file.")
	tmpFile := tempFile(t, `---
atlantis-url: "url"
data-dir: "path"
gh-hostname: "ghhostname"
gh-user: "user"
gh-token: "token"
gh-webhook-secret: "secret"
gitlab-hostname: "gitlab-hostname"
gitlab-user: "gitlab-user"
gitlab-token: "gitlab-token"
gitlab-webhook-secret: "gitlab-secret"
log-level: "debug"
port: 8181
require-approval: true`)
	defer os.Remove(tmpFile) // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.ConfigFlag: tmpFile,
	})

	err := c.Execute()
	Ok(t, err)
	Equals(t, "url", passedConfig.AtlantisURL)
	Equals(t, "path", passedConfig.DataDir)
	Equals(t, "ghhostname", passedConfig.GithubHostname)
	Equals(t, "user", passedConfig.GithubUser)
	Equals(t, "token", passedConfig.GithubToken)
	Equals(t, "secret", passedConfig.GithubWebHookSecret)
	Equals(t, "gitlab-hostname", passedConfig.GitlabHostname)
	Equals(t, "gitlab-user", passedConfig.GitlabUser)
	Equals(t, "gitlab-token", passedConfig.GitlabToken)
	Equals(t, "gitlab-secret", passedConfig.GitlabWebHookSecret)
	Equals(t, "debug", passedConfig.LogLevel)
	Equals(t, 8181, passedConfig.Port)
	Equals(t, true, passedConfig.RequireApproval)
}

func TestExecute_EnvironmentOverride(t *testing.T) {
	t.Log("Environment variables should override config file flags.")
	tmpFile := tempFile(t, "gh-user: config\ngh-token: config2")
	defer os.Remove(tmpFile)                   // nolint: errcheck
	os.Setenv("ATLANTIS_GH_TOKEN", "override") // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.ConfigFlag: tmpFile,
	})
	err := c.Execute()
	Ok(t, err)
	Equals(t, "override", passedConfig.GithubToken)
}

func TestExecute_FlagConfigOverride(t *testing.T) {
	t.Log("Flags should override config file flags.")
	os.Setenv("ATLANTIS_GH_TOKEN", "env-var") // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "override",
	})
	err := c.Execute()
	Ok(t, err)
	Equals(t, "override", passedConfig.GithubToken)
}

func TestExecute_FlagEnvVarOverride(t *testing.T) {
	t.Log("Flags should override environment variables.")
	tmpFile := tempFile(t, "gh-user: config\ngh-token: config2")
	defer os.Remove(tmpFile) // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.ConfigFlag:  tmpFile,
		cmd.GHTokenFlag: "override",
	})

	err := c.Execute()
	Ok(t, err)
	Equals(t, "override", passedConfig.GithubToken)
}

func TestExecute_EnvVars(t *testing.T) {
	t.Log("Setting flags by env var should work.")
	os.Setenv("ATLANTIS_GH_TOKEN", "gh-token")                    // nolint: errcheck
	os.Setenv("ATLANTIS_GH_WEBHOOK_SECRET", "gh-webhook")         // nolint: errcheck
	os.Setenv("ATLANTIS_GITLAB_TOKEN", "gitlab-token")            // nolint: errcheck
	os.Setenv("ATLANTIS_GITLAB_WEBHOOK_SECRET", "gitlab-webhook") // nolint: errcheck
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:     "user",
		cmd.GitlabUserFlag: "user",
	})
	err := c.Execute()
	Ok(t, err)
	Equals(t, "gh-token", passedConfig.GithubToken)
	Equals(t, "gh-webhook", passedConfig.GithubWebHookSecret)
	Equals(t, "gitlab-token", passedConfig.GitlabToken)
	Equals(t, "gitlab-webhook", passedConfig.GitlabWebHookSecret)
}

func setup(flags map[string]interface{}) *cobra.Command {
	viper := viper.New()
	for k, v := range flags {
		viper.Set(k, v)
	}
	c := &cmd.ServerCmd{
		ServerCreator: &ServerCreatorMock{},
		Viper:         viper,
		SilenceOutput: true,
	}
	return c.Init()
}

func tempFile(t *testing.T, contents string) string {
	f, err := ioutil.TempFile("", "")
	Ok(t, err)
	newName := f.Name() + ".yaml"
	err = os.Rename(f.Name(), newName)
	Ok(t, err)
	ioutil.WriteFile(newName, []byte(contents), 0644) // nolint: errcheck
	return newName
}
