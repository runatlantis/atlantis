package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/urfave/cli"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var baseConfig *AtlantisConfig = &AtlantisConfig{
	GithubHostname: stringPtr("test-gh-hostname"),
	GithubUsername: stringPtr("test-gh-username"),
	GithubPassword: stringPtr("test-gh-password"),
	SSHKey:         stringPtr("test-ssh-key"),
	AssumeRole:     stringPtr("test-assume-role"),
	Port:           9999,
	ScratchDir:     stringPtr("test-scratch-dir"),
	AWSRegion:      stringPtr("test-aws-region"),
	S3Bucket:       stringPtr("test-s3-bucket"),
	LogLevel:       stringPtr("info"),
	RequireApproval: false,
}

func TestValidateConfig_missing_config_file_should_error(t *testing.T) {
	ctx := buildTestContext()
	ok(t, ctx.Set(configFileFlag, "non-existent-file"))
	_, err := validateConfig(ctx)
	assert(t, err != nil, "expected an error")
	assert(t, strings.Contains(err.Error(), "couldn't read config file"), "error did not contain expected message, was: %q", err.Error())
}

func TestValidateConfig_non_json_config_file_should_error(t *testing.T) {
	configFile := writeConfigFile(t, "invalid_json")
	defer os.Remove(configFile)
	ctx := buildTestContext()

	ok(t, ctx.Set(configFileFlag, configFile))
	_, err := validateConfig(ctx)
	assert(t, err != nil, "expected an error")
	assert(t, strings.Contains(err.Error(), "failed to parse config file"), "error did not contain expected message, was: %q", err.Error())
}

func TestValidateConfig_should_use_defaults(t *testing.T) {
	ctx := buildTestContext()
	ok(t, ctx.Set(ghUsernameFlag, "gh-username"))
	ok(t, ctx.Set(ghPasswordFlag, "gh-password"))
	conf, err := validateConfig(ctx)
	ok(t, err)

	equals(t, defaultGHHostname, conf.githubHostname)
	equals(t, defaultPort, conf.port)
	equals(t, defaultScratchDir, conf.scratchDir)
	equals(t, defaultRegion, conf.awsRegion)
	equals(t, defaultS3Bucket, conf.s3Bucket)
	equals(t, defaultLogLevel, conf.logLevel)
	equals(t, false, conf.requireApproval)
	equals(t, "", conf.sshKey)
	equals(t, "", conf.awsAssumeRole)
}

func TestValidateConfig_config_file_should_work(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()

	ok(t, ctx.Set(configFileFlag, configFile))
	conf, err := validateConfig(ctx)
	ok(t, err)
	equals(t, "test-gh-hostname", conf.githubHostname)
	equals(t, "test-gh-username", conf.githubUsername)
	equals(t, "test-gh-password", conf.githubPassword)
	equals(t, "test-ssh-key", conf.sshKey)
	equals(t, "test-assume-role", conf.awsAssumeRole)
	equals(t, 9999, conf.port)
	equals(t, "test-scratch-dir", conf.scratchDir)
	equals(t, "test-aws-region", conf.awsRegion)
	equals(t, "test-s3-bucket", conf.s3Bucket)
	equals(t, false, conf.requireApproval)
	equals(t, "info", conf.logLevel)
}

func TestValidateConfig_flags_should_work(t *testing.T) {
	ctx := buildTestContext()
	ok(t, ctx.Set(ghHostnameFlag, "gh-hostname"))
	ok(t, ctx.Set(ghUsernameFlag, "gh-username"))
	ok(t, ctx.Set(ghPasswordFlag, "gh-password"))
	ok(t, ctx.Set(sshKeyFlag, "ssh-key"))
	ok(t, ctx.Set(awsAssumeRoleFlag, "assume-role"))
	ok(t, ctx.Set(portFlag, "8888"))
	ok(t, ctx.Set(scratchDirFlag, "scratch-dir"))
	ok(t, ctx.Set(awsRegionFlag, "aws-region"))
	ok(t, ctx.Set(s3BucketFlag, "s3-bucket"))
	ok(t, ctx.Set(logLevelFlag, "debug"))
	ok(t, ctx.Set(requireApprovalFlag, "true"))

	conf, err := validateConfig(ctx)
	ok(t, err)
	equals(t, "gh-hostname", conf.githubHostname)
	equals(t, "gh-username", conf.githubUsername)
	equals(t, "gh-password", conf.githubPassword)
	equals(t, "ssh-key", conf.sshKey)
	equals(t, "assume-role", conf.awsAssumeRole)
	equals(t, 8888, conf.port)
	equals(t, "scratch-dir", conf.scratchDir)
	equals(t, "aws-region", conf.awsRegion)
	equals(t, "s3-bucket", conf.s3Bucket)
	equals(t, true, conf.requireApproval)
	equals(t, "debug", conf.logLevel)
}

func TestValidateConfig_flags_should_override_config_file(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()

	ok(t, ctx.Set(configFileFlag, configFile))

	// override all flags
	ok(t, ctx.Set(ghHostnameFlag, "overridden-gh-hostname"))
	ok(t, ctx.Set(ghUsernameFlag, "overridden-gh-username"))
	ok(t, ctx.Set(ghPasswordFlag, "overridden-gh-password"))
	ok(t, ctx.Set(sshKeyFlag, "overridden-ssh-key"))
	ok(t, ctx.Set(awsAssumeRoleFlag, "overridden-assume-role"))
	ok(t, ctx.Set(portFlag, "8888"))
	ok(t, ctx.Set(scratchDirFlag, "overridden-scratch-dir"))
	ok(t, ctx.Set(awsRegionFlag, "overridden-aws-region"))
	ok(t, ctx.Set(s3BucketFlag, "overridden-s3-bucket"))
	ok(t, ctx.Set(requireApprovalFlag, "true"))
	ok(t, ctx.Set(logLevelFlag, "debug"))

	conf, err := validateConfig(ctx)
	ok(t, err)
	equals(t, "overridden-gh-hostname", conf.githubHostname)
	equals(t, "overridden-gh-username", conf.githubUsername)
	equals(t, "overridden-gh-password", conf.githubPassword)
	equals(t, "overridden-ssh-key", conf.sshKey)
	equals(t, "overridden-assume-role", conf.awsAssumeRole)
	equals(t, 8888, conf.port)
	equals(t, "overridden-scratch-dir", conf.scratchDir)
	equals(t, "overridden-aws-region", conf.awsRegion)
	equals(t, "overridden-s3-bucket", conf.s3Bucket)
	equals(t, true, conf.requireApproval)
	equals(t, "debug", conf.logLevel)
}

func TestValidateConfig_missing_required_flags_should_error(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, flag := range []string{ghUsernameFlag, ghPasswordFlag} {
		ctx := buildTestContext()
		ok(t, ctx.Set(configFileFlag, configFile))
		ok(t, ctx.Set(flag, ""))
		_, err := validateConfig(ctx)
		assert(t, err != nil, "expected an error")
		expected := fmt.Sprintf("Error: must specify the --%s flag", flag)
		assert(t, strings.Contains(err.Error(), expected), "error did not contain expected message, was: %q, expected: %q", err.Error(), expected)
	}
}

func TestValidateConfig_invalid_log_level_should_error(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()
	ok(t, ctx.Set(configFileFlag, configFile))
	ok(t, ctx.Set(logLevelFlag, "invalid-level"))
	_, err := validateConfig(ctx)
	assert(t, err != nil, "expected an error")
	expected := fmt.Sprintf("Invalid log level")
	assert(t, strings.Contains(err.Error(), expected), "error did not contain expected message, was: %q, expected: %q", err.Error(), expected)
}

func TestValidateConfig_valid_log_levels_should_validate(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, level := range []string{"debug", "info", "warn", "error"} {
		ctx := buildTestContext()
		ok(t, ctx.Set(configFileFlag, configFile))
		ok(t, ctx.Set(logLevelFlag, level))
		conf, err := validateConfig(ctx)
		assert(t, err == nil, "Did not expect error for valid log level %q: %v", level, err)
		equals(t, level, conf.logLevel)
	}
}

func TestValidateConfig_uppercase_log_levels_should_validate(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, level := range []string{"DEBUG", "INFO", "WARN", "ERROR"} {
		ctx := buildTestContext()
		ok(t, ctx.Set(configFileFlag, configFile))
		ok(t, ctx.Set(logLevelFlag, level))
		conf, err := validateConfig(ctx)
		assert(t, err == nil, "Did not expect error for valid log level %q: %v", level, err)
		equals(t, strings.ToLower(level), conf.logLevel)
	}
}

func writeJSONConfigFile(t *testing.T, config *AtlantisConfig) string {
	bytes, err := json.Marshal(config)
	ok(t, err)
	return writeConfigFile(t, string(bytes))
}

func writeConfigFile(t *testing.T, contents string) string {
	path := "/tmp/atlantis-test-config-file"
	invalidJSON := []byte(contents)
	ok(t, ioutil.WriteFile(path, invalidJSON, 0644))
	return path
}

func buildTestContext() *cli.Context {
	app := cli.NewApp()
	configureCli(app)
	return cli.NewContext(app, flagSet("test", app.Flags), nil)
}

func flagSet(name string, flags []cli.Flag) *flag.FlagSet {
	set := flag.NewFlagSet(name, 0)
	for _, f := range flags {
		f.Apply(set)
	}
	return set
}

func stringPtr(v string) *string { return &v }
