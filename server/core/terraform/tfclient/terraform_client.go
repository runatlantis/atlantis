// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package tfclient handles the actual running of terraform commands.
package tfclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	"github.com/mitchellh/go-homedir"

	"github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/core/terraform/ansi"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
)

var LogStreamingValidCmds = [...]string{"init", "plan", "apply"}

const versionCommandTimeout = 10 * time.Second

//go:generate go tool pegomock generate --package mocks -o mocks/mock_terraform_client.go Client

type Client interface {
	// RunCommandWithVersion executes terraform with args in path. If v is nil,
	// it will use the default Terraform version. workspace is the Terraform
	// workspace which should be set as an environment variable.
	RunCommandWithVersion(ctx command.ProjectContext, path string, args []string, envs map[string]string, d terraform.Distribution, v *version.Version, workspace string) (string, error)

	// EnsureVersion makes sure that terraform version `v` is available to use
	EnsureVersion(log logging.SimpleLogging, d terraform.Distribution, v *version.Version) error

	// DetectVersion extracts required_version from Terraform/OpenTofu configuration in the specified project directory.
	// Non-exact constraints are resolved using the provided distribution, or the client default distribution when nil.
	// Returns nil if unable to determine the version.
	DetectVersion(log logging.SimpleLogging, d terraform.Distribution, projectDirectory string) *version.Version
}

type DefaultClient struct {
	// Distribution handles logic specific to the TF distribution being used by Atlantis
	distribution terraform.Distribution

	// defaultVersion is the default version of terraform to use if another
	// version isn't specified.
	defaultVersion *version.Version
	// We will run terraform with the TF_PLUGIN_CACHE_DIR env var set to this
	// directory inside our data dir.
	terraformPluginCacheDir string
	binDir                  string
	// overrideTF can be used to override the terraform binary during testing
	// with another binary, ex. echo.
	overrideTF string
	// settings for the downloader.
	downloadBaseURL string
	downloadAllowed bool
	// versions maps from the string representation of a tf version (ex. 0.11.10)
	// to the absolute path of that binary on disk (if it exists).
	// Use versionsLock to control access.
	versions map[string]string
	// versionLocks serializes install/remove operations for each Terraform version.
	// Use versionsLock to control access.
	versionLocks map[string]*sync.Mutex
	// downloadLock serializes downloader installs that share intermediate paths in binDir.
	downloadLock *sync.Mutex

	// versionsLock is used to ensure versions isn't being concurrently written to.
	versionsLock *sync.Mutex

	// usePluginCache determines whether or not to set the TF_PLUGIN_CACHE_DIR env var
	usePluginCache bool

	projectCmdOutputHandler jobs.ProjectCommandOutputHandler
}

// versionRegex extracts the version from `terraform version` output.
//
//	    Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//		   => 0.12.0-alpha4
//
//	    Terraform v0.11.10
//		   => 0.11.10
//
//	    OpenTofu v1.0.0
//		   => 1.0.0
var versionRegex = regexp.MustCompile("(?:Terraform|OpenTofu) v(.*?)(\\s.*)?\n")

// NewClientWithDefaultVersion creates a new terraform client and pre-fetches the default version
func NewClientWithDefaultVersion(
	log logging.SimpleLogging,
	distribution terraform.Distribution,
	binDir string,
	cacheDir string,
	tfeToken string,
	tfeHostname string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloadAllowed bool,
	usePluginCache bool,
	fetchAsync bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
) (*DefaultClient, error) {
	var finalDefaultVersion *version.Version
	var localVersion *version.Version
	versions := make(map[string]string)
	versionLocks := make(map[string]*sync.Mutex)
	var versionsLock sync.Mutex
	var downloadLock sync.Mutex

	localPath, err := exec.LookPath(distribution.BinName())
	if err != nil && defaultVersionStr == "" {
		return nil, fmt.Errorf("%s not found in $PATH. Set --%s or download terraform from https://developer.hashicorp.com/terraform/downloads", distribution.BinName(), defaultVersionFlagName)
	}
	if err == nil {
		localVersion, err = getVersion(localPath, distribution.BinName())
		if err != nil {
			return nil, err
		}
		versions[localVersion.String()] = localPath
		if defaultVersionStr == "" {

			// If they haven't set a default version, then whatever they had
			// locally is now the default.
			finalDefaultVersion = localVersion
		}
	}

	if defaultVersionStr != "" {
		defaultVersion, err := version.NewVersion(defaultVersionStr)
		if err != nil {
			return nil, err
		}
		finalDefaultVersion = defaultVersion
		ensureVersionFunc := func() {
			// Since ensureVersion might end up downloading terraform,
			// we call it asynchronously so as to not delay server startup.
			_, err := ensureVersion(log, distribution, versions, versionLocks, &versionsLock, &downloadLock, defaultVersion, binDir, tfDownloadURL, tfDownloadAllowed, true)
			if err != nil {
				log.Err("could not download %s %s: %s", distribution.BinName(), defaultVersion.String(), err)
			}
		}

		if fetchAsync {
			go ensureVersionFunc()
		} else {
			ensureVersionFunc()
		}
	}

	// If tfeToken is set, we try to create a ~/.terraformrc file.
	if tfeToken != "" {
		home, err := homedir.Dir()
		if err != nil {
			return nil, fmt.Errorf("getting home dir to write ~/.terraformrc file: %w", err)
		}
		if err := generateRCFile(tfeToken, tfeHostname, home); err != nil {
			return nil, err
		}
	}
	return &DefaultClient{
		distribution:            distribution,
		defaultVersion:          finalDefaultVersion,
		terraformPluginCacheDir: cacheDir,
		binDir:                  binDir,
		downloadBaseURL:         tfDownloadURL,
		downloadAllowed:         tfDownloadAllowed,
		versionsLock:            &versionsLock,
		versions:                versions,
		versionLocks:            versionLocks,
		downloadLock:            &downloadLock,
		usePluginCache:          usePluginCache,
		projectCmdOutputHandler: projectCmdOutputHandler,
	}, nil

}

func NewTestClient(
	log logging.SimpleLogging,
	distribution terraform.Distribution,
	binDir string,
	cacheDir string,
	tfeToken string,
	tfeHostname string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloadAllowed bool,
	usePluginCache bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
) (*DefaultClient, error) {
	return NewClientWithDefaultVersion(
		log,
		distribution,
		binDir,
		cacheDir,
		tfeToken,
		tfeHostname,
		defaultVersionStr,
		defaultVersionFlagName,
		tfDownloadURL,
		tfDownloadAllowed,
		usePluginCache,
		false,
		projectCmdOutputHandler,
	)
}

// NewClient constructs a terraform client.
// tfeToken is an optional terraform enterprise token.
// defaultVersionStr is an optional default terraform version to use unless
// a specific version is set.
// defaultVersionFlagName is the name of the flag that sets the default terraform
// version.
// Will asynchronously download the required version if it doesn't exist already.
func NewClient(
	log logging.SimpleLogging,
	distribution terraform.Distribution,
	binDir string,
	cacheDir string,
	tfeToken string,
	tfeHostname string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloadAllowed bool,
	usePluginCache bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
) (*DefaultClient, error) {
	return NewClientWithDefaultVersion(
		log,
		distribution,
		binDir,
		cacheDir,
		tfeToken,
		tfeHostname,
		defaultVersionStr,
		defaultVersionFlagName,
		tfDownloadURL,
		tfDownloadAllowed,
		usePluginCache,
		true,
		projectCmdOutputHandler,
	)
}

func (c *DefaultClient) DefaultDistribution() terraform.Distribution {
	return c.distribution
}

// Version returns the default version of Terraform we use if no other version
// is defined.
func (c *DefaultClient) DefaultVersion() *version.Version {
	return c.defaultVersion
}

// TerraformBinDir returns the directory where we download Terraform binaries.
func (c *DefaultClient) TerraformBinDir() string {
	return c.binDir
}

// ExtractExactRegex attempts to extract an exact version number from the provided string as a fallback.
// The function expects the version string to be in one of the following formats: "= x.y.z", "=x.y.z", or "x.y.z" where x, y, and z are integers.
// If the version string matches one of these formats, the function returns a slice containing the exact version number.
// If the version string does not match any of these formats, the function logs a debug message and returns nil.
func (c *DefaultClient) ExtractExactRegex(log logging.SimpleLogging, version string) []string {
	re := regexp.MustCompile(`^=?\s*([0-9.]+)\s*$`)
	matched := re.FindStringSubmatch(version)
	if len(matched) == 0 {
		log.Debug("exact version regex not found in the version %q", version)
		return nil
	}
	// The first element of the slice is the entire string, so we want the second element (the first capture group)
	tfVersions := []string{matched[1]}
	log.Debug("extracted exact version %q from version %q", tfVersions[0], version)
	return tfVersions
}

// DetectVersion extracts required_version from Terraform/OpenTofu configuration in the specified project directory.
// If downloads are allowed, non-exact constraints are resolved against the provided distribution, or the client
// default distribution when nil, and the highest satisfying version is selected.
// Returns nil if unable to determine the version.
func (c *DefaultClient) DetectVersion(log logging.SimpleLogging, d terraform.Distribution, projectDirectory string) *version.Version {
	requiredCore := c.detectRequiredCore(log, d, projectDirectory)

	if len(requiredCore) != 1 {
		log.Info("cannot determine which version to use from terraform configuration, detected %d possibilities.", len(requiredCore))
		return nil
	}
	requiredVersionSetting := requiredCore[0]
	log.Debug("Found required_version setting of %q", requiredVersionSetting)

	if !c.downloadAllowed {
		log.Debug("terraform downloads disabled.")
		matched := c.ExtractExactRegex(log, requiredVersionSetting)
		if len(matched) == 0 {
			log.Debug("did not specify exact version in terraform configuration, found %q", requiredVersionSetting)
			return nil
		}

		version, err := version.NewVersion(matched[0])
		if err != nil {
			log.Err("error parsing version string: %s", err)
			return nil
		}
		return version
	}

	downloadVersion, err := c.effectiveDistribution(d).ResolveConstraint(context.Background(), requiredVersionSetting)
	if err != nil {
		log.Err("%s", err)
		return nil
	}

	return downloadVersion
}

// detectRequiredCore returns the required_version constraints from configuration files.
// For OpenTofu distribution, it uses a local parser that supports .tofu/.tofu.json files
// with proper precedence (.tofu overrides same-basename .tf). For Terraform distribution,
// it uses hashicorp/terraform-config-inspect which only reads .tf/.tf.json files.
func (c *DefaultClient) detectRequiredCore(log logging.SimpleLogging, d terraform.Distribution, projectDirectory string) []string {
	dist := c.effectiveDistribution(d)
	if dist.BinName() == "tofu" {
		constraints, err := detectRequiredCoreFromTofu(projectDirectory)
		if err != nil {
			log.Err("trying to detect required version from OpenTofu config: %s", err)
		}
		if len(constraints) == 0 && err != nil {
			return nil
		}
		return constraints
	}

	module, diags := tfconfig.LoadModule(projectDirectory)
	if diags.HasErrors() {
		log.Err("trying to detect required version: %s", diags.Error())
	}
	if module == nil {
		return nil
	}
	return module.RequiredCore
}

// See Client.EnsureVersion.
func (c *DefaultClient) EnsureVersion(log logging.SimpleLogging, d terraform.Distribution, v *version.Version) error {
	if v == nil {
		v = c.defaultVersion
	}

	d = c.effectiveDistribution(d)
	_, err := ensureVersion(log, d, c.versions, c.versionLocks, c.versionsLock, c.downloadLock, v, c.binDir, c.downloadBaseURL, c.downloadAllowed, true)
	if err != nil {
		return err
	}

	return nil
}

// See Client.RunCommandWithVersion.
func (c *DefaultClient) RunCommandWithVersion(ctx command.ProjectContext, path string, args []string, customEnvVars map[string]string, d terraform.Distribution, v *version.Version, workspace string) (string, error) {
	if isAsyncEligibleCommand(args[0]) {
		_, outCh := c.RunCommandAsync(ctx, path, args, customEnvVars, d, v, workspace)

		var lines []string
		var err error
		for line := range outCh {
			if line.Err != nil {
				err = line.Err
				break
			}
			lines = append(lines, line.Line)
		}
		output := strings.Join(lines, "\n")

		// sanitize output by stripping out any ansi characters.
		output = ansi.Strip(output)
		return fmt.Sprintf("%s\n", output), err
	}
	tfCmd, cmd, err := c.prepExecCmd(ctx, d, v, workspace, path, args, customEnvVars)
	if err != nil {
		return "", err
	}
	unlock, err := c.lockPluginCache(ctx.Log, args[0], cmd.Env)
	if err != nil {
		return "", err
	}
	defer unlock()

	start := time.Now()
	out, err := cmd.CombinedOutput()
	dur := time.Since(start)
	log := ctx.Log.With("duration", dur)
	if err != nil {
		err = fmt.Errorf("running '%s' in '%s': %w", tfCmd, path, err)
		log.Err("%s", err.Error())
		return ansi.Strip(string(out)), err
	}
	log.Info("Successfully ran '%s' in '%s'", tfCmd, path)

	return ansi.Strip(string(out)), nil
}

// prepExecCmd builds a ready to execute command based on the version of terraform
// v, and args. It returns a printable representation of the command that will
// be run and the actual command.
func (c *DefaultClient) prepExecCmd(ctx command.ProjectContext, d terraform.Distribution, v *version.Version, workspace string, path string, args []string, customEnvVars map[string]string) (string, *exec.Cmd, error) {
	argv, display, envVars, err := c.prepCmd(ctx.Log, d, v, workspace, path, args, customEnvVars, ctx.ExpandableArgs, ctx.CommentArgs)
	if err != nil {
		return "", nil, err
	}
	cmd := exec.Command(argv[0], argv[1:]...) // #nosec G204 -- argv[0] is a resolved Terraform binary path, and the arguments are passed as a vector rather than as shell source
	cmd.Dir = path
	cmd.Env = envVars
	return display, cmd, nil
}

// prepCmd prepares the argument vector and the set of environment variables for
// running terraform. The vector is executed directly, without a shell, so that
// a metacharacter in an argument is never interpreted. Arguments still undergo
// environment variable expansion, because extra_args is documented to be able
// to refer to $WORKSPACE and friends, but that expansion is performed here
// rather than by a shell, so it neither word splits nor globs.
func (c *DefaultClient) prepCmd(log logging.SimpleLogging, d terraform.Distribution, v *version.Version, workspace string, path string, args []string, customEnvVars map[string]string, expandable []string, literal []string) ([]string, string, []string, error) {

	if v == nil {
		v = c.defaultVersion
	}

	var binPath string
	if c.overrideTF != "" {
		// This is only set during testing.
		binPath = c.overrideTF
	} else {
		var err error
		d = c.effectiveDistribution(d)
		binPath, err = ensureVersion(log, d, c.versions, c.versionLocks, c.versionsLock, c.downloadLock, v, c.binDir, c.downloadBaseURL, c.downloadAllowed, true)
		if err != nil {
			return nil, "", nil, err
		}
	}

	// We add custom variables so that if `extra_args` is specified with env
	// vars then they'll be substituted.
	envVars := []string{
		// Will de-emphasize specific commands to run in output.
		"TF_IN_AUTOMATION=true",
		// Cache plugins so terraform init runs faster.
		fmt.Sprintf("WORKSPACE=%s", workspace),
		fmt.Sprintf("ATLANTIS_TERRAFORM_VERSION=%s", v.String()),
		fmt.Sprintf("DIR=%s", path),
	}
	if c.usePluginCache {
		envVars = append(envVars, fmt.Sprintf("TF_PLUGIN_CACHE_DIR=%s", c.terraformPluginCacheDir))
	}
	// Append current Atlantis process's environment variables, ex.
	// AWS_ACCESS_KEY.
	envVars = append(envVars, os.Environ()...)
	// Appended last so that they win over the process environment, and before
	// expansion so that an extra_args entry can refer to a variable a workflow
	// set with an `env` step.
	for key, val := range customEnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}

	argv := make([]string, 0, len(args)+1)
	argv = append(argv, binPath)
	expand := envLookup(envVars)
	mayExpand := expansionSet(expandable, literal)
	for _, arg := range args {
		if mayExpand[arg] {
			argv = append(argv, os.Expand(arg, expand))
			continue
		}
		argv = append(argv, arg)
	}

	// The display string keeps variable references unexpanded. It ends up in
	// logs and in the error comments Atlantis posts on pull requests, and an
	// extra_args entry may well reference a variable holding a credential.
	display := displayCmd(append([]string{binPath}, args...))

	return argv, display, envVars, nil
}

// expansionSet returns the arguments whose environment variable references may
// be expanded. Only operator-configured extra_args qualify. Anything that also
// appears in literal, which is the set of arguments taken from a pull request
// comment, is excluded, so that a commenter cannot have a value expanded by
// repeating an operator's argument back.
func expansionSet(expandable []string, literal []string) map[string]bool {
	if len(expandable) == 0 {
		return nil
	}
	fromComment := make(map[string]bool, len(literal))
	for _, arg := range literal {
		fromComment[arg] = true
	}
	set := make(map[string]bool, len(expandable))
	for _, arg := range expandable {
		if !fromComment[arg] {
			set[arg] = true
		}
	}
	return set
}

// envLookup returns a lookup function over environ, with later entries winning
// over earlier ones. That matches how the child process resolves duplicates, so
// expansion here sees the same values the process will.
func envLookup(environ []string) func(string) string {
	values := make(map[string]string, len(environ))
	for _, kv := range environ {
		if key, value, found := strings.Cut(kv, "="); found {
			values[key] = value
		}
	}
	return func(key string) string {
		return values[key]
	}
}

// displayCmd renders an argument vector for logs and error messages. It is
// never executed.
func displayCmd(argv []string) string {
	return strings.Join(argv, " ")
}

func (c *DefaultClient) effectiveDistribution(d terraform.Distribution) terraform.Distribution {
	if d != nil {
		return d
	}
	return c.distribution
}

// RunCommandAsync prepares terraform and acquires any plugin-cache lock, then
// returns channels for realtime output and stdin. An error is always the final
// output value, so callers may stop reading it.
func (c *DefaultClient) RunCommandAsync(ctx command.ProjectContext, path string, args []string, customEnvVars map[string]string, d terraform.Distribution, v *version.Version, workspace string) (chan<- string, <-chan models.Line) {
	argv, display, envVars, err := c.prepCmd(ctx.Log, d, v, workspace, path, args, customEnvVars, ctx.ExpandableArgs, ctx.CommentArgs)
	if err != nil {
		outCh := make(chan models.Line)
		inCh := make(chan string)
		go func() {
			defer close(outCh)
			defer close(inCh)
			outCh <- models.Line{Err: err}
		}()
		return inCh, outCh
	}

	unlock, err := c.lockPluginCache(ctx.Log, args[0], envVars)
	if err != nil {
		outCh := make(chan models.Line, 1)
		inCh := make(chan string)
		outCh <- models.Line{Err: err}
		close(outCh)
		close(inCh)
		return inCh, outCh
	}

	runner := models.NewArgvCommandRunner(argv, display, envVars, path, !ctx.SuppressJobOutput, c.projectCmdOutputHandler)
	inCh, runnerOutCh := runner.RunCommandAsync(ctx)
	outCh := make(chan models.Line)
	go func() {
		defer close(outCh)
		defer unlock()
		// ShellCommandRunner guarantees that an error is its final send before
		// closing runnerOutCh. Callers may therefore stop after receiving it
		// without preventing this goroutine from releasing the cache lock.
		for line := range runnerOutCh {
			outCh <- line
		}
	}()
	return inCh, outCh
}

// Terraform's plugin cache is not safe when init writes to it while another
// Terraform command is reading from it.
func (c *DefaultClient) lockPluginCache(log logging.SimpleLogging, command string, environ []string) (func(), error) {
	cacheDir := envLookup(environ)("TF_PLUGIN_CACHE_DIR")
	if cacheDir == "" {
		return func() {}, nil
	}
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		return nil, fmt.Errorf("creating terraform plugin cache directory: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pluginCacheLockTimeout)
	defer cancel()
	unlock, err := lockFile(ctx, log, filepath.Join(cacheDir, ".atlantis.lock"), command == "init")
	if err != nil {
		return nil, fmt.Errorf("locking terraform plugin cache: %w", err)
	}
	return unlock, nil
}

// MustConstraint will parse one or more constraints from the given
// constraint string. The string must be a comma-separated list of
// constraints. It panics if there is an error.
func MustConstraint(v string) version.Constraints {
	c, err := version.NewConstraint(v)
	if err != nil {
		panic(err)
	}
	return c
}

// ensureVersion returns the path to a terraform binary of version v.
// It will download this version if we don't have it.
func ensureVersion(
	log logging.SimpleLogging,
	dist terraform.Distribution,
	versions map[string]string,
	versionLocks map[string]*sync.Mutex,
	versionsLock *sync.Mutex,
	downloadLock *sync.Mutex,
	v *version.Version,
	binDir string,
	downloadURL string,
	downloadsAllowed bool,
	redownloadOnFailedExecution bool,
) (string, error) {
	if dist == nil {
		return "", errors.New("terraform distribution is nil")
	}

	binPath, err := findOrDownloadVersionBinaryPath(log, dist, versions, versionLocks, versionsLock, downloadLock, v, binDir, downloadURL, downloadsAllowed)
	if err != nil {
		return "", err
	}
	if !redownloadOnFailedExecution {
		return binPath, nil
	}

	binName := dist.BinName()
	if err := validateVersionBinary(binPath, binName); err == nil {
		return binPath, nil
	} else if !downloadsAllowed {
		return "", invalidVersionBinaryError(binPath, binName, err)
	}

	log.Warn("%s binary %s failed execution validation, attempting to re-download", binName, binPath)
	binPath, err = redownloadVersionBinary(log, dist, versions, versionLocks, versionsLock, downloadLock, v, binPath, binDir, downloadURL)
	if err != nil {
		return "", err
	}
	if err := validateVersionBinary(binPath, binName); err != nil {
		return "", invalidVersionBinaryError(binPath, binName, err)
	}
	return binPath, nil
}

func validateVersionBinary(binPath string, binName string) error {
	_, err := getVersion(binPath, binName)
	return err
}

func invalidVersionBinaryError(binPath string, binName string, err error) error {
	return fmt.Errorf("%s binary at %s failed to execute: %w", binName, binPath, err)
}

func isManagedVersionBinary(binPath string, binDir string) bool {
	rel, err := filepath.Rel(binDir, binPath)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}

// findOrDownloadVersionBinaryPath returns the path to a terraform binary of version v.
// It will download this version if we don't have it.
func findOrDownloadVersionBinaryPath(
	log logging.SimpleLogging,
	dist terraform.Distribution,
	versions map[string]string,
	versionLocks map[string]*sync.Mutex,
	versionsLock *sync.Mutex,
	downloadLock *sync.Mutex,
	v *version.Version,
	binDir string,
	downloadURL string,
	downloadsAllowed bool,
) (string, error) {
	if binPath, ok := getVersionBinaryPath(versions, versionsLock, v); ok {
		return binPath, nil
	}

	versionLock := getVersionOperationLock(versionLocks, versionsLock, v)
	versionLock.Lock()
	defer versionLock.Unlock()

	if binPath, ok := getVersionBinaryPath(versions, versionsLock, v); ok {
		return binPath, nil
	}

	// This tf version might not yet be in the versions map even though it
	// exists on disk. This would happen if users have manually added
	// terraform{version} binaries. In this case we don't want to re-download.
	binFile := dist.BinName() + v.String()
	if binPath, err := exec.LookPath(binFile); err == nil {
		setVersionBinaryPath(versions, versionsLock, v, binPath)
		return binPath, nil
	}

	// The version might also not be in the versions map if it's in our bin dir.
	// This could happen if Atlantis was restarted without losing its disk.
	dest := filepath.Join(binDir, binFile)
	if _, err := os.Stat(dest); err == nil {
		setVersionBinaryPath(versions, versionsLock, v, dest)
		return dest, nil
	}
	if !downloadsAllowed {
		return "", fmt.Errorf(
			"could not find %s version %s in PATH or %s, and downloads are disabled",
			dist.BinName(),
			v.String(),
			binDir,
		)
	}

	log.Info("could not find %s version %s in PATH or %s", dist.BinName(), v.String(), binDir)
	execPath, err := downloadVersionBinary(log, dist, downloadLock, v, binDir, downloadURL)
	if err != nil {
		return "", err
	}
	setVersionBinaryPath(versions, versionsLock, v, execPath)
	return execPath, nil
}

func redownloadVersionBinary(
	log logging.SimpleLogging,
	dist terraform.Distribution,
	versions map[string]string,
	versionLocks map[string]*sync.Mutex,
	versionsLock *sync.Mutex,
	downloadLock *sync.Mutex,
	v *version.Version,
	binPath string,
	binDir string,
	downloadURL string,
) (string, error) {
	versionLock := getVersionOperationLock(versionLocks, versionsLock, v)
	versionLock.Lock()
	defer versionLock.Unlock()

	if currentPath, ok := getVersionBinaryPath(versions, versionsLock, v); ok && currentPath != binPath {
		return currentPath, nil
	} else if ok {
		if err := validateVersionBinary(currentPath, dist.BinName()); err == nil {
			return currentPath, nil
		}
	}
	deleteVersionBinaryPath(versions, versionsLock, v, binPath)
	if isManagedVersionBinary(binPath, binDir) {
		if err := os.Remove(binPath); err != nil && !os.IsNotExist(err) {
			return "", fmt.Errorf("removing cached %s binary for redownload at %s: %w", dist.BinName(), binPath, err)
		}
	}

	execPath, err := downloadVersionBinary(log, dist, downloadLock, v, binDir, downloadURL)
	if err != nil {
		return "", err
	}
	setVersionBinaryPath(versions, versionsLock, v, execPath)
	return execPath, nil
}

func downloadVersionBinary(
	log logging.SimpleLogging,
	dist terraform.Distribution,
	downloadLock *sync.Mutex,
	v *version.Version,
	binDir string,
	downloadURL string,
) (string, error) {
	downloadLock.Lock()
	defer downloadLock.Unlock()

	log.Info("downloading %s version %s from download URL %s", dist.BinName(), v.String(), downloadURL)

	execPath, err := dist.Downloader().Install(context.Background(), binDir, downloadURL, v)
	if err != nil {
		return "", fmt.Errorf("error downloading %s version %s: %w", dist.BinName(), v.String(), err)
	}

	log.Info("Downloaded %s %s to %s", dist.BinName(), v.String(), execPath)
	return execPath, nil
}

func getVersionOperationLock(versionLocks map[string]*sync.Mutex, versionsLock *sync.Mutex, v *version.Version) *sync.Mutex {
	versionsLock.Lock()
	defer versionsLock.Unlock()

	lockKey := v.String()
	versionLock, ok := versionLocks[lockKey]
	if !ok {
		versionLock = &sync.Mutex{}
		versionLocks[lockKey] = versionLock
	}
	return versionLock
}

func getVersionBinaryPath(versions map[string]string, versionsLock *sync.Mutex, v *version.Version) (string, bool) {
	versionsLock.Lock()
	defer versionsLock.Unlock()
	binPath, ok := versions[v.String()]
	return binPath, ok
}

func setVersionBinaryPath(versions map[string]string, versionsLock *sync.Mutex, v *version.Version, binPath string) {
	versionsLock.Lock()
	defer versionsLock.Unlock()
	versions[v.String()] = binPath
}

func deleteVersionBinaryPath(versions map[string]string, versionsLock *sync.Mutex, v *version.Version, binPath string) {
	versionsLock.Lock()
	defer versionsLock.Unlock()
	if versions[v.String()] == binPath {
		delete(versions, v.String())
	}
}

// generateRCFile generates a .terraformrc file containing config for tfeToken
// and hostname tfeHostname.
// It will create the file in home/.terraformrc.
func generateRCFile(tfeToken string, tfeHostname string, home string) error {
	const rcFilename = ".terraformrc"
	rcFile := filepath.Join(home, rcFilename)
	config := fmt.Sprintf(rcFileContents, tfeHostname, tfeToken)

	// If there is already a .terraformrc file and its contents aren't exactly
	// what we would have written to it, then we error out because we don't
	// want to overwrite anything.
	if _, err := os.Stat(rcFile); err == nil {
		currContents, err := os.ReadFile(rcFile) // nolint: gosec
		if err != nil {
			return fmt.Errorf("trying to read %s to ensure we're not overwriting it: %w", rcFile, err)
		}
		if config != string(currContents) {
			return fmt.Errorf("can't write TFE token to %s because that file has contents that would be overwritten", rcFile)
		}
		// Otherwise we don't need to write the file because it already has
		// what we need.
		return nil
	}

	if err := os.WriteFile(rcFile, []byte(config), 0600); err != nil {
		return fmt.Errorf("writing generated %s file with TFE token to %s: %w", rcFilename, rcFile, err)
	}
	return nil
}

func isAsyncEligibleCommand(cmd string) bool {
	for _, validCmd := range LogStreamingValidCmds {
		if validCmd == cmd {
			return true
		}
	}
	return false
}

func getVersion(tfBinary string, binName string) (*version.Version, error) {
	ctx, cancel := context.WithTimeout(context.Background(), versionCommandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, tfBinary, "version") // #nosec
	cmd.Env = terraformVersionEnv(os.Environ())
	versionOutBytes, err := cmd.Output()
	versionOutput := string(versionOutBytes)
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("running %s version timed out after %s", binName, versionCommandTimeout)
	}
	if err != nil {
		return nil, fmt.Errorf("running %s version: %s: %w", binName, versionOutput, err)
	}
	match := versionRegex.FindStringSubmatch(versionOutput)
	if len(match) <= 1 {
		return nil, fmt.Errorf("could not parse %s version from %s", binName, versionOutput)
	}
	return version.NewVersion(match[1])
}

func terraformVersionEnv(env []string) []string {
	out := make([]string, 0, len(env)+3)
	for _, envVar := range env {
		key, _, ok := strings.Cut(envVar, "=")
		if ok && (key == "CHECKPOINT_DISABLE" || key == "TF_CLI_ARGS" || key == "TF_CLI_ARGS_version") {
			continue
		}
		out = append(out, envVar)
	}
	return append(out, "CHECKPOINT_DISABLE=1", "TF_CLI_ARGS=", "TF_CLI_ARGS_version=")
}

// rcFileContents is a format string to be used with Sprintf that can be used
// to generate the contents of a ~/.terraformrc file for authenticating with
// Terraform Enterprise.
var rcFileContents = `credentials "%s" {
  token = %q
}`
