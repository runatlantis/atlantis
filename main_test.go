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
	. "github.com/hootsuite/atlantis/testing_util"
)

var hostname = "hostname"
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
	AtlantisURL:   stringPtr("http://test-atlantis-url"),
	RequireApproval: false,
	DataDir: stringPtr("/datadir"),
}

func TestValidateConfig_missing_config_file_should_error(t *testing.T) {
	ctx := buildTestContext()
	Ok(t, ctx.Set(configFileFlag, "non-existent-file"))
	_, err := validateConfig(ctx, hostname)
	Assert(t, err != nil, "expected an error")
	Assert(t, strings.Contains(err.Error(), "couldn't read config file"), "error did not contain expected message, was: %q", err.Error())
}

func TestValidateConfig_non_json_config_file_should_error(t *testing.T) {
	configFile := writeConfigFile(t, "invalid_json")
	defer os.Remove(configFile)
	ctx := buildTestContext()

	Ok(t, ctx.Set(configFileFlag, configFile))
	_, err := validateConfig(ctx, hostname)
	Assert(t, err != nil, "expected an error")
	Assert(t, strings.Contains(err.Error(), "failed to parse config file"), "error did not contain expected message, was: %q", err.Error())
}

func TestValidateConfig_should_use_defaults(t *testing.T) {
	ctx := buildTestContext()
	Ok(t, ctx.Set(ghUsernameFlag, "gh-username"))
	Ok(t, ctx.Set(ghPasswordFlag, "gh-password"))
	conf, err := validateConfig(ctx, hostname)
	Ok(t, err)

	Equals(t, defaultGHHostname, conf.githubHostname)
	Equals(t, defaultPort, conf.port)
	Equals(t, defaultScratchDir, conf.scratchDir)
	Equals(t, defaultRegion, conf.awsRegion)
	Equals(t, defaultS3Bucket, conf.s3Bucket)
	Equals(t, defaultLogLevel, conf.logLevel)
	Equals(t, false, conf.requireApproval)
	Equals(t, "", conf.sshKey)
	Equals(t, "", conf.awsAssumeRole)
	Equals(t, fmt.Sprintf("http://hostname:%d", defaultPort), conf.atlantisURL)
	Equals(t, defaultDataDir, conf.dataDir)
}

func TestValidateConfig_config_file_should_work(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()

	Ok(t, ctx.Set(configFileFlag, configFile))
	conf, err := validateConfig(ctx, hostname)
	Ok(t, err)
	Equals(t, "test-gh-hostname", conf.githubHostname)
	Equals(t, "test-gh-username", conf.githubUsername)
	Equals(t, "test-gh-password", conf.githubPassword)
	Equals(t, "test-ssh-key", conf.sshKey)
	Equals(t, "test-assume-role", conf.awsAssumeRole)
	Equals(t, 9999, conf.port)
	Equals(t, "test-scratch-dir", conf.scratchDir)
	Equals(t, "test-aws-region", conf.awsRegion)
	Equals(t, "test-s3-bucket", conf.s3Bucket)
	Equals(t, "info", conf.logLevel)
	Equals(t, "http://test-atlantis-url", conf.atlantisURL)
	Equals(t, false, conf.requireApproval)
	Equals(t, "/datadir", conf.dataDir)
}

func TestValidateConfig_flags_should_work(t *testing.T) {
	ctx := buildTestContext()
	Ok(t, ctx.Set(ghHostnameFlag, "gh-hostname"))
	Ok(t, ctx.Set(ghUsernameFlag, "gh-username"))
	Ok(t, ctx.Set(ghPasswordFlag, "gh-password"))
	Ok(t, ctx.Set(sshKeyFlag, "ssh-key"))
	Ok(t, ctx.Set(awsAssumeRoleFlag, "assume-role"))
	Ok(t, ctx.Set(portFlag, "8888"))
	Ok(t, ctx.Set(scratchDirFlag, "scratch-dir"))
	Ok(t, ctx.Set(awsRegionFlag, "aws-region"))
	Ok(t, ctx.Set(s3BucketFlag, "s3-bucket"))
	Ok(t, ctx.Set(logLevelFlag, "debug"))
	Ok(t, ctx.Set(atlantisURLFlag, "http://new-atlantis-url"))
	Ok(t, ctx.Set(requireApprovalFlag, "true"))
	Ok(t, ctx.Set(dataDirFlag, "/new/dir"))

	conf, err := validateConfig(ctx, hostname)
	Ok(t, err)
	Equals(t, "gh-hostname", conf.githubHostname)
	Equals(t, "gh-username", conf.githubUsername)
	Equals(t, "gh-password", conf.githubPassword)
	Equals(t, "ssh-key", conf.sshKey)
	Equals(t, "assume-role", conf.awsAssumeRole)
	Equals(t, 8888, conf.port)
	Equals(t, "scratch-dir", conf.scratchDir)
	Equals(t, "aws-region", conf.awsRegion)
	Equals(t, "s3-bucket", conf.s3Bucket)
	Equals(t, "debug", conf.logLevel)
	Equals(t, "http://new-atlantis-url", conf.atlantisURL)
	Equals(t, true, conf.requireApproval)
	Equals(t, "/new/dir", conf.dataDir)
}

func TestValidateConfig_flags_should_override_config_file(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()

	Ok(t, ctx.Set(configFileFlag, configFile))

	// override all flags
	Ok(t, ctx.Set(ghHostnameFlag, "overridden-gh-hostname"))
	Ok(t, ctx.Set(ghUsernameFlag, "overridden-gh-username"))
	Ok(t, ctx.Set(ghPasswordFlag, "overridden-gh-password"))
	Ok(t, ctx.Set(sshKeyFlag, "overridden-ssh-key"))
	Ok(t, ctx.Set(awsAssumeRoleFlag, "overridden-assume-role"))
	Ok(t, ctx.Set(portFlag, "8888"))
	Ok(t, ctx.Set(scratchDirFlag, "overridden-scratch-dir"))
	Ok(t, ctx.Set(awsRegionFlag, "overridden-aws-region"))
	Ok(t, ctx.Set(s3BucketFlag, "overridden-s3-bucket"))
	Ok(t, ctx.Set(logLevelFlag, "debug"))
	Ok(t, ctx.Set(atlantisURLFlag, "overridden-url"))
	Ok(t, ctx.Set(requireApprovalFlag, "true"))
	Ok(t, ctx.Set(dataDirFlag, "/overridden/dir"))

	conf, err := validateConfig(ctx, hostname)
	Ok(t, err)
	Equals(t, "overridden-gh-hostname", conf.githubHostname)
	Equals(t, "overridden-gh-username", conf.githubUsername)
	Equals(t, "overridden-gh-password", conf.githubPassword)
	Equals(t, "overridden-ssh-key", conf.sshKey)
	Equals(t, "overridden-assume-role", conf.awsAssumeRole)
	Equals(t, 8888, conf.port)
	Equals(t, "overridden-scratch-dir", conf.scratchDir)
	Equals(t, "overridden-aws-region", conf.awsRegion)
	Equals(t, "overridden-s3-bucket", conf.s3Bucket)
	Equals(t, "debug", conf.logLevel)
	Equals(t, "overridden-url", conf.atlantisURL)
	Equals(t, true, conf.requireApproval)
	Equals(t, "/overridden/dir", conf.dataDir)
}

func TestValidateConfig_missing_required_flags_should_error(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, flag := range []string{ghUsernameFlag, ghPasswordFlag} {
		ctx := buildTestContext()
		Ok(t, ctx.Set(configFileFlag, configFile))
		Ok(t, ctx.Set(flag, ""))
		_, err := validateConfig(ctx, hostname)
		Assert(t, err != nil, "expected an error")
		expected := fmt.Sprintf("Error: must specify the --%s flag", flag)
		Assert(t, strings.Contains(err.Error(), expected), "error did not contain expected message, was: %q, expected: %q", err.Error(), expected)
	}
}

func TestValidateConfig_invalid_log_level_should_error(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)
	ctx := buildTestContext()
	Ok(t, ctx.Set(configFileFlag, configFile))
	Ok(t, ctx.Set(logLevelFlag, "invalid-level"))
	_, err := validateConfig(ctx, hostname)
	Assert(t, err != nil, "expected an error")
	expected := fmt.Sprintf("Invalid log level")
	Assert(t, strings.Contains(err.Error(), expected), "error did not contain expected message, was: %q, expected: %q", err.Error(), expected)
}

func TestValidateConfig_valid_log_levels_should_validate(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, level := range []string{"debug", "info", "warn", "error"} {
		ctx := buildTestContext()
		Ok(t, ctx.Set(configFileFlag, configFile))
		Ok(t, ctx.Set(logLevelFlag, level))
			conf, err := validateConfig(ctx, hostname)
		Assert(t, err == nil, "Did not expect error for valid log level %q: %v", level, err)
		Equals(t, level, conf.logLevel)
	}
}

func TestValidateConfig_uppercase_log_levels_should_validate(t *testing.T) {
	configFile := writeJSONConfigFile(t, baseConfig)
	defer os.Remove(configFile)

	for _, level := range []string{"DEBUG", "INFO", "WARN", "ERROR"} {
		ctx := buildTestContext()
		Ok(t, ctx.Set(configFileFlag, configFile))
		Ok(t, ctx.Set(logLevelFlag, level))
		conf, err := validateConfig(ctx, hostname)
		Assert(t, err == nil, "Did not expect error for valid log level %q: %v", level, err)
		Equals(t, strings.ToLower(level), conf.logLevel)
	}
}

func writeJSONConfigFile(t *testing.T, config *AtlantisConfig) string {
	bytes, err := json.Marshal(config)
	Ok(t, err)
	return writeConfigFile(t, string(bytes))
}

func writeConfigFile(t *testing.T, contents string) string {
	path := "/tmp/atlantis-test-config-file"
	invalidJSON := []byte(contents)
	Ok(t, ioutil.WriteFile(path, invalidJSON, 0644))
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
