package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/viper"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func setupValidate(flags map[string]any, t *testing.T) *ServerValidateCmd {
	t.Helper()
	v := viper.New()
	for k, val := range flags {
		v.Set(k, val)
	}
	return &ServerValidateCmd{
		Viper:  v,
		Logger: logging.NewNoopLogger(t),
	}
}

func TestValidate_MissingConfigFlag(t *testing.T) {
	t.Log("Should error when --config is not provided.")
	cmd := setupValidate(nil, t)
	c := cmd.Init()
	err := c.Execute()
	ErrContains(t, "--config is required", err)
}

func TestValidate_ConfigFileNotFound(t *testing.T) {
	t.Log("Should error when config file does not exist.")
	cmd := setupValidate(map[string]any{
		ConfigFlag: "does-not-exist.yaml",
	}, t)
	c := cmd.Init()
	err := c.Execute()
	ErrContains(t, "does-not-exist.yaml", err)
}

func TestValidate_ValidConfig(t *testing.T) {
	t.Log("Should succeed when config contains only known keys.")
	tmpFile := tempFile(t, "repo-allowlist: github.com/test/*\nport: 4141\n")
	defer os.Remove(tmpFile) // nolint: errcheck
	cmd := setupValidate(map[string]any{
		ConfigFlag: tmpFile,
	}, t)
	c := cmd.Init()
	err := c.Execute()
	Ok(t, err)
}

func TestValidate_UnknownKey(t *testing.T) {
	t.Log("Should error when config contains unknown keys.")
	tmpFile := tempFile(t, "repo-allowlist: github.com/test/*\nallow-draft-pr: true\nparallel_apply: true\n")
	defer os.Remove(tmpFile) // nolint: errcheck
	cmd := setupValidate(map[string]any{
		ConfigFlag: tmpFile,
	}, t)
	c := cmd.Init()
	err := c.Execute()
	Assert(t, err != nil, "should error on unknown keys")
	ErrContains(t, "unknown keys", err)
	ErrContains(t, "allow-draft-pr", err)
	ErrContains(t, "parallel_apply", err)
}

func TestValidate_InvalidYAML(t *testing.T) {
	t.Log("Should error on invalid YAML.")
	tmpFile := tempFile(t, "invalid: yaml: content: [")
	defer os.Remove(tmpFile) // nolint: errcheck
	cmd := setupValidate(map[string]any{
		ConfigFlag: tmpFile,
	}, t)
	c := cmd.Init()
	err := c.Execute()
	Assert(t, err != nil, "should error on invalid YAML")
}

func TestFindUnknownKeys_AllKnown(t *testing.T) {
	t.Log("Should return empty slice when all keys are known.")
	v := viper.New()
	v.Set("repo-allowlist", "github.com/test/*")
	v.Set("port", 4141)
	v.Set("allow-draft-prs", true)
	unknowns := findUnknownKeys(v)
	Equals(t, 0, len(unknowns))
}

func TestFindUnknownKeys_WithUnknown(t *testing.T) {
	t.Log("Should return unknown keys sorted alphabetically.")
	v := viper.New()
	v.Set("repo-allowlist", "github.com/test/*")
	v.Set("tofu-version", "1.11.5")
	v.Set("allow-draft-pr", true)
	unknowns := findUnknownKeys(v)
	Equals(t, 2, len(unknowns))
	Equals(t, "allow-draft-pr", unknowns[0])
	Equals(t, "tofu-version", unknowns[1])
}

func TestBuildKnownKeys_ContainsExpectedKeys(t *testing.T) {
	t.Log("Should contain known keys from UserConfig mapstructure tags.")
	keys := buildKnownKeys()
	expectedKeys := []string{
		"repo-allowlist", "port", "allow-draft-prs", "parallel-plan",
		"parallel-apply", "autodiscover-mode", "config",
	}
	for _, key := range expectedKeys {
		if _, ok := keys[key]; !ok {
			t.Errorf("expected key %q not found in known keys", key)
		}
	}
}

func TestRun_UnknownKeyWarning(t *testing.T) {
	t.Log("Server run() should warn about unknown keys in config.")
	cfgContents := "repo-allowlist: github.com/test/*\ntofu-version: 1.11.5\n"
	tmpFile := tempFile(t, cfgContents)
	defer os.Remove(tmpFile) // nolint: errcheck

	v := viper.New()
	v.Set(ConfigFlag, tmpFile)
	v.Set(GHUserFlag, "user")
	v.Set(GHTokenFlag, "token")
	v.Set(RepoAllowlistFlag, "*")

	logger := &captureLogger{}
	s := &ServerCmd{
		ServerCreator: &ServerCreatorMock{},
		Viper:         v,
		SilenceOutput: true,
		Logger:        logger,
	}
	c := s.Init()
	_ = c.Execute()

	found := false
	for _, msg := range logger.warnings {
		if strings.Contains(msg, "unknown keys") && strings.Contains(msg, "tofu-version") {
			found = true
			break
		}
	}
	Assert(t, found, "expected warning about unknown key 'tofu-version'")
}

// captureLogger captures log messages for testing.
type captureLogger struct {
	warnings []string
}

func (l *captureLogger) Debug(string, ...any)              {}
func (l *captureLogger) Info(string, ...any)               {}
func (l *captureLogger) Warn(format string, a ...any) {
	l.warnings = append(l.warnings, fmt.Sprintf(format, a...))
}
func (l *captureLogger) Err(string, ...any)                {}
func (l *captureLogger) Log(logging.LogLevel, string, ...any) {}
func (l *captureLogger) SetLevel(logging.LogLevel)         {}
func (l *captureLogger) With(...any) logging.SimpleLogging  { return l }
func (l *captureLogger) WithHistory(...any) logging.SimpleLogging { return l }
func (l *captureLogger) GetHistory() string                { return "" }
func (l *captureLogger) Flush() error                      { return nil }
