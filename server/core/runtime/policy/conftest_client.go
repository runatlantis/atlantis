// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter/v2"

	version "github.com/hashicorp/go-version"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	runtime_models "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	DefaultConftestVersionEnvKey = "DEFAULT_CONFTEST_VERSION"
	conftestBinaryName           = "conftest"
	conftestDownloadURLPrefix    = "https://github.com/open-policy-agent/conftest/releases/download/v"
)

type Arg struct {
	Param  string
	Option string
}

func (a Arg) build(env map[string]string) []string {
	return append([]string{a.Option}, splitConfiguredArg(a.Param, env)...)
}

type configuredArgQuote uint8

const (
	configuredArgUnquoted configuredArgQuote = iota
	configuredArgSingleQuoted
	configuredArgDoubleQuoted
)

// splitConfiguredArg parses the shell-like whitespace and quoting supported by
// administrator-configured policy paths and extra_args without interpreting
// the result as shell source. Environment references expand only while they are
// unquoted or double quoted. Single quotes and backslashes preserve a literal
// dollar sign, as they did when the command was run by a shell.
//
// Expansion happens while quote and escape provenance is still available. An
// expanded value remains within its source argument even when it contains
// whitespace. A malformed value is passed through as one literal argument
// because its expansion intent cannot be determined safely.
func splitConfiguredArg(value string, env map[string]string) []string {
	lookup := func(key string) string {
		// LocalExec appends the process environment after workflow values,
		// and exec.Cmd resolves duplicates in favor of the later entry.
		if val, ok := os.LookupEnv(key); ok {
			return val
		}
		return env[key]
	}

	var args []string
	var word strings.Builder
	var expandable strings.Builder
	quote := configuredArgUnquoted
	wordStarted := false

	flushExpandable := func() {
		if expandable.Len() == 0 {
			return
		}
		word.WriteString(os.Expand(expandable.String(), lookup))
		expandable.Reset()
	}
	flushWord := func() {
		flushExpandable()
		if !wordStarted {
			return
		}
		args = append(args, word.String())
		word.Reset()
		wordStarted = false
	}

	for i := 0; i < len(value); i++ {
		char := value[i]
		switch quote {
		case configuredArgSingleQuoted:
			if char == '\'' {
				quote = configuredArgUnquoted
				continue
			}
			word.WriteByte(char)
			wordStarted = true
		case configuredArgDoubleQuoted:
			switch char {
			case '"':
				flushExpandable()
				quote = configuredArgUnquoted
			case '\\':
				flushExpandable()
				if i+1 >= len(value) {
					return []string{value}
				}
				i++
				word.WriteByte(value[i])
				wordStarted = true
			default:
				expandable.WriteByte(char)
				wordStarted = true
			}
		case configuredArgUnquoted:
			switch char {
			case '\'', '"':
				flushExpandable()
				wordStarted = true
				if char == '\'' {
					quote = configuredArgSingleQuoted
				} else {
					quote = configuredArgDoubleQuoted
				}
			case '\\':
				flushExpandable()
				if i+1 >= len(value) {
					return []string{value}
				}
				i++
				word.WriteByte(value[i])
				wordStarted = true
			case ' ', '\t', '\r', '\n':
				flushWord()
			case '#':
				if wordStarted {
					expandable.WriteByte(char)
					continue
				}
				for i+1 < len(value) && value[i+1] != '\n' {
					i++
				}
			default:
				expandable.WriteByte(char)
				wordStarted = true
			}
		}
	}

	if quote != configuredArgUnquoted {
		return []string{value}
	}
	flushWord()
	return args
}

func NewPolicyArg(parameter string) Arg {
	return Arg{
		Param:  parameter,
		Option: "-p",
	}
}

type ConftestTestCommandArgs struct {
	PolicyArgs []Arg
	ExtraArgs  []string
	InputFile  string
	Command    string
	// Env are the environment variables the conftest process will receive.
	// Administrator-configured values are expanded against them, since the
	// previous shell-based command line expanded them too.
	Env map[string]string
}

func (c ConftestTestCommandArgs) build() ([]string, error) {

	if len(c.PolicyArgs) == 0 {
		return []string{}, errors.New("no policies specified")
	}

	// add the subcommand
	commandArgs := []string{c.Command, "test"}

	for _, a := range c.PolicyArgs {
		commandArgs = append(commandArgs, a.build(c.Env)...)
	}

	// add hardcoded options
	commandArgs = append(commandArgs, c.InputFile, "--no-color")

	// add extra args provided through server config
	for _, extraArg := range c.ExtraArgs {
		commandArgs = append(commandArgs, splitConfiguredArg(extraArg, c.Env)...)
	}

	return commandArgs, nil
}

// SourceResolver resolves the policy set to a local fs path
//
//go:generate go tool pegomock generate --package mocks -o mocks/mock_conftest_client.go SourceResolver
type SourceResolver interface {
	Resolve(policySet valid.PolicySet) (string, error)
}

// LocalSourceResolver resolves a local policy set to a local fs path
type LocalSourceResolver struct {
}

func (p *LocalSourceResolver) Resolve(policySet valid.PolicySet) (string, error) {
	return policySet.Path, nil

}

// SourceResolverProxy proxies to underlying source resolvers dynamically
type SourceResolverProxy struct {
	localSourceResolver SourceResolver
}

func (p *SourceResolverProxy) Resolve(policySet valid.PolicySet) (string, error) {
	switch source := policySet.Source; source {
	case valid.LocalPolicySet:
		return p.localSourceResolver.Resolve(policySet)
	default:
		return "", fmt.Errorf("unable to resolve policy set source %s", source)
	}
}

//go:generate go tool pegomock generate --package mocks -o mocks/mock_downloader.go Downloader

type Downloader interface {
	GetAny(dst, src string) error
}

type ConfTestGoGetterVersionDownloader struct{}

func (c ConfTestGoGetterVersionDownloader) GetAny(dst, src string) error {
	_, err := getter.GetAny(context.Background(), dst, src)
	return err
}

type ConfTestVersionDownloader struct {
	downloader Downloader
}

func (c ConfTestVersionDownloader) downloadConfTestVersion(v *version.Version, destPath string) (runtime_models.FilePath, error) {
	versionURLPrefix := fmt.Sprintf("%s%s", conftestDownloadURLPrefix, v.Original())

	conftestPlatform := getPlatform()
	if conftestPlatform == "" {
		return runtime_models.LocalFilePath(""), fmt.Errorf("don't know where to find conftest for %s on %s", runtime.GOOS, runtime.GOARCH)
	}

	// download binary in addition to checksum file
	binURL := fmt.Sprintf("%s/conftest_%s_%s.tar.gz", versionURLPrefix, v.Original(), conftestPlatform)
	checksumURL := fmt.Sprintf("%s/checksums.txt", versionURLPrefix)

	// underlying implementation uses go-getter so the URL is formatted as such.
	// i know i know, I'm assuming an interface implementation with my inputs.
	// realistically though the interface just exists for testing so ¯\_(ツ)_/¯
	fullSrcURL := fmt.Sprintf("%s?checksum=file:%s", binURL, checksumURL)

	if err := c.downloader.GetAny(destPath, fullSrcURL); err != nil {
		return runtime_models.LocalFilePath(""), fmt.Errorf("downloading conftest version %s at %q: %w", v.String(), fullSrcURL, err)
	}

	binPath := filepath.Join(destPath, "conftest")

	return runtime_models.LocalFilePath(binPath), nil
}

// ConfTestExecutorWorkflow runs a versioned conftest binary with the args built from the project context.
// Project context defines whether conftest runs a local policy set or runs a test on a remote policy set.
type ConfTestExecutorWorkflow struct {
	SourceResolver         SourceResolver
	VersionCache           cache.ExecutionVersionCache
	DefaultConftestVersion *version.Version
	Exec                   runtime_models.Exec
}

func NewConfTestExecutorWorkflow(log logging.SimpleLogging, versionRootDir string, conftestDownloder Downloader) *ConfTestExecutorWorkflow {
	downloader := ConfTestVersionDownloader{
		downloader: conftestDownloder,
	}
	version, err := getDefaultVersion()

	if err != nil {
		// conftest default versions are not essential to service startup so let's not block on it.
		log.Info("failed to get default conftest version. Will attempt request scoped lazy loads %s", err.Error())
	}

	versionCache := cache.NewExecutionVersionLayeredLoadingCache(
		conftestBinaryName,
		versionRootDir,
		downloader.downloadConfTestVersion,
	)

	return &ConfTestExecutorWorkflow{
		VersionCache:           versionCache,
		DefaultConftestVersion: version,
		SourceResolver: &SourceResolverProxy{
			localSourceResolver: &LocalSourceResolver{},
		},
		Exec: runtime_models.LocalExec{},
	}
}

func (c *ConfTestExecutorWorkflow) Run(ctx command.ProjectContext, executablePath string, envs map[string]string, workdir string, extraArgs []string) (string, error) {
	ctx.Log.Debug("policy sets, %v ", ctx.PolicySets)

	inputFile := filepath.Join(workdir, ctx.GetShowResultFileName())
	var policySetResults []models.PolicySetResult
	var combinedErr error

	for _, policySet := range ctx.PolicySets.PolicySets {
		path, resolveErr := c.SourceResolver.Resolve(policySet)

		// Let's not fail the whole step because of a single failure. Log and fail silently
		if resolveErr != nil {
			ctx.Log.Err("Error resolving policyset %s. err: %s", policySet.Name, resolveErr.Error())
			continue
		}

		args := ConftestTestCommandArgs{
			PolicyArgs: []Arg{NewPolicyArg(path)},
			ExtraArgs:  extraArgs,
			InputFile:  inputFile,
			Command:    executablePath,
			Env:        envs,
		}

		serializedArgs, _ := args.build()
		cmdOutput, cmdErr := c.Exec.CombinedOutput(serializedArgs, envs, workdir)

		if cmdErr != nil {
			// Since we're running conftest for each policyset, individual command errors should be concatenated.
			if isValidConftestOutput(cmdOutput) {
				combinedErr = errors.Join(combinedErr, fmt.Errorf("policy_set: %s: conftest: some policies failed", policySet.Name))
			} else {
				combinedErr = errors.Join(combinedErr, fmt.Errorf("policy_set: %s: conftest: %s", policySet.Name, cmdOutput))
			}
		}

		passed := true
		if cmdErr != nil || hasFailures(cmdOutput) {
			passed = false
		}

		// Sanitize before hashing so that hashes are stable across runs
		// (temp file paths vary) and match what ends up in PolicyOutput.
		sanitizedOutput := c.sanitizeOutput(inputFile, cmdOutput)

		result, regexErr := models.NewPolicySetResult(
			policySet.Name,
			sanitizedOutput,
			passed,
			policySet.ApproveCount,
			policySet.PolicyItemRegex,
		)
		if regexErr != nil {
			// RegexValidator runs at config-parse time so this is in theory
			// unreachable. Fail closed with a synthetic failing result so the
			// project surfaces the misconfiguration rather than silently
			// passing without this policy set.
			ctx.Log.Err("invalid policy_item_regex for policy set %q: %v", policySet.Name, regexErr)
			policySetResults = append(policySetResults, models.PolicySetResult{
				PolicySetName:    policySet.Name,
				PolicyOutput:     fmt.Sprintf("invalid policy_item_regex %q: %v", policySet.PolicyItemRegex, regexErr),
				ReqApprovalCount: policySet.ApproveCount,
				PolicyItemRegex:  policySet.PolicyItemRegex,
			})
			continue
		}
		policySetResults = append(policySetResults, *result)
	}

	if policySetResults == nil {
		ctx.Log.Warn("no policies have been configured.")
		return "", nil
		// TODO: enable when we can pass policies in otherwise e2e tests with policy checks fail
		// return "", errors.Wrap(err, "building args")
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(policySetResults); err != nil {
		return "", errors.New("cannot marshal data into []PolicySetResult. data")
	}
	marshaledStatus := bytes.TrimRight(buf.Bytes(), "\n")

	// Write policy check results to a file which can be used by custom workflow run steps for metrics, notifications, etc.
	policyCheckResultFile := filepath.Join(workdir, ctx.GetPolicyCheckResultFileName())
	if writeErr := os.WriteFile(policyCheckResultFile, marshaledStatus, 0600); writeErr != nil {
		combinedErr = errors.Join(combinedErr, writeErr)
	}

	return string(marshaledStatus), combinedErr

}

func (c *ConfTestExecutorWorkflow) sanitizeOutput(inputFile string, output string) string {
	return strings.ReplaceAll(output, inputFile, "<redacted plan file>")
}

func (c *ConfTestExecutorWorkflow) EnsureExecutorVersion(log logging.SimpleLogging, v *version.Version) (string, error) {
	// we have no information to proceed, so fallback to `conftest` command or fail hard
	if c.DefaultConftestVersion == nil && v == nil {
		localPath, err := c.Exec.LookPath(conftestBinaryName)
		if err == nil {
			log.Info("conftest version is not specified, so fallback to conftest command")
			return localPath, nil
		}
		return "", errors.New("no conftest version configured/specified or not found conftest command")
	}

	var versionToRetrieve *version.Version

	if v == nil {
		versionToRetrieve = c.DefaultConftestVersion
	} else {
		versionToRetrieve = v
	}

	localPath, err := c.VersionCache.Get(versionToRetrieve)

	if err != nil {
		return "", err
	}

	return localPath, nil

}

func getDefaultVersion() (*version.Version, error) {
	// ensure version is not default version.
	// first check for the env var and if that doesn't exist use the local executable version
	defaultVersion, exists := os.LookupEnv(DefaultConftestVersionEnvKey)

	if !exists {
		return nil, fmt.Errorf("%s not set", DefaultConftestVersionEnvKey)
	}

	wrappedVersion, err := version.NewVersion(defaultVersion)

	if err != nil {
		return nil, fmt.Errorf("wrapping version %s: %w", defaultVersion, err)
	}
	return wrappedVersion, nil
}

// Checks if output from conftest is a valid output.
func isValidConftestOutput(output string) bool {

	r := regexp.MustCompile(`^(WARN|FAIL|\[)`)
	if match := r.FindString(output); match != "" {
		return true
	}
	return false
}

// hasFailures checks whether any conftest policies have failed
func hasFailures(output string) bool {
	r := regexp.MustCompile(`([1-9]([0-9]?)* failure|failures": \[)`)
	if match := r.FindString(output); match != "" {
		return true
	}
	return false
}

func getPlatform() string {
	platform := runtime.GOOS + "_" + runtime.GOARCH

	switch platform {
	case "linux_amd64":
		return "Linux_x86_64"
	case "linux_arm64":
		return "Linux_arm64"
	case "darwin_amd64":
		return "Darwin_x86_64"
	case "darwin_arm64":
		return "Darwin_arm64"
	default:
		return ""
	}
}
