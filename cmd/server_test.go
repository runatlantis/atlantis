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

func TestExecute_Validation(t *testing.T) {
	cases := []struct {
		description string
		flags       map[string]interface{}
		expErr      string
	}{
		{
			"Should validate log level.",
			map[string]interface{}{
				cmd.LogLevelFlag: "invalid",
				cmd.GHUserFlag:   "user",
				cmd.GHTokenFlag:  "token",
			},
			"invalid log level: not one of debug, info, warn, error",
		},
		{
			"Should ensure github user is set.",
			map[string]interface{}{
				cmd.GHTokenFlag: "token",
			},
			"--gh-user must be set",
		},
		{
			"Should ensure github token is set.",
			map[string]interface{}{
				cmd.GHUserFlag: "user",
			},
			"--gh-token must be set",
		},
	}
	for _, testCase := range cases {
		t.Log(testCase.description)
		c := setup(testCase.flags)
		err := c.Execute()
		Assert(t, err != nil, "should be an error")
		Equals(t, testCase.expErr, err.Error())
	}
}

func TestExecute_Defaults(t *testing.T) {
	t.Log("Should set the defaults for all unspecified flags.")
	c := setup(map[string]interface{}{
		cmd.GHUserFlag:  "user",
		cmd.GHTokenFlag: "token",
	})
	err := c.Execute()
	Ok(t, err)

	Equals(t, "user", passedConfig.GithubUser)
	Equals(t, "token", passedConfig.GithubToken)
	Equals(t, "", passedConfig.GithubWebHookSecret)
	// Get our hostname since that's what gets defaulted to
	hostname, err := os.Hostname()
	Ok(t, err)
	Equals(t, "http://"+hostname+":4141", passedConfig.AtlantisURL)

	// Get our home dir since that's what gets defaulted to
	dataDir, err := homedir.Expand("~/.atlantis")
	Ok(t, err)
	Equals(t, dataDir, passedConfig.DataDir)
	Equals(t, "github.com", passedConfig.GithubHostname)
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

func TestExecute_Flags(t *testing.T) {
	t.Log("Should use all flags that are set.")
	c := setup(map[string]interface{}{
		cmd.AtlantisURLFlag:     "url",
		cmd.DataDirFlag:         "path",
		cmd.GHHostnameFlag:      "ghhostname",
		cmd.GHUserFlag:          "user",
		cmd.GHTokenFlag:         "token",
		cmd.GHWebHookSecret:     "secret",
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
