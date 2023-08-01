package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"encoding/json"
	"regexp"

	"github.com/hashicorp/go-multierror"
	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	runtime_models "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	DefaultConftestVersionEnvKey = "DEFAULT_CONFTEST_VERSION"
	conftestBinaryName           = "conftest"
	conftestDownloadURLPrefix    = "https://github.com/open-policy-agent/conftest/releases/download/v"
	conftestArch                 = "x86_64"
)

type Arg struct {
	Param  string
	Option string
}

func (a Arg) build() []string {
	return []string{a.Option, a.Param}
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
}

func (c ConftestTestCommandArgs) build() ([]string, error) {

	if len(c.PolicyArgs) == 0 {
		return []string{}, errors.New("no policies specified")
	}

	// add the subcommand
	commandArgs := []string{c.Command, "test"}

	for _, a := range c.PolicyArgs {
		commandArgs = append(commandArgs, a.build()...)
	}

	// add hardcoded options
	commandArgs = append(commandArgs, c.InputFile, "--no-color")

	// add extra args provided through server config
	commandArgs = append(commandArgs, c.ExtraArgs...)

	return commandArgs, nil
}

// SourceResolver resolves the policy set to a local fs path
//
//go:generate pegomock generate --package mocks -o mocks/mock_conftest_client.go SourceResolver
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

type ConfTestVersionDownloader struct {
	downloader terraform.Downloader
}

func (c ConfTestVersionDownloader) downloadConfTestVersion(v *version.Version, destPath string) (runtime_models.FilePath, error) {
	versionURLPrefix := fmt.Sprintf("%s%s", conftestDownloadURLPrefix, v.Original())

	// download binary in addition to checksum file
	binURL := fmt.Sprintf("%s/conftest_%s_%s_%s.tar.gz", versionURLPrefix, v.Original(), cases.Title(language.English).String(runtime.GOOS), conftestArch)
	checksumURL := fmt.Sprintf("%s/checksums.txt", versionURLPrefix)

	// underlying implementation uses go-getter so the URL is formatted as such.
	// i know i know, I'm assuming an interface implementation with my inputs.
	// realistically though the interface just exists for testing so ¯\_(ツ)_/¯
	fullSrcURL := fmt.Sprintf("%s?checksum=file:%s", binURL, checksumURL)

	if err := c.downloader.GetAny(destPath, fullSrcURL); err != nil {
		return runtime_models.LocalFilePath(""), errors.Wrapf(err, "downloading conftest version %s at %q", v.String(), fullSrcURL)
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

func NewConfTestExecutorWorkflow(log logging.SimpleLogging, versionRootDir string, conftestDownloder terraform.Downloader) *ConfTestExecutorWorkflow {
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
	ctx.Log.Debug("policy sets, %s ", ctx.PolicySets)

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
		}

		serializedArgs, _ := args.build()
		cmdOutput, cmdErr := c.Exec.CombinedOutput(serializedArgs, envs, workdir)

		if cmdErr != nil {
			// Since we're running conftest for each policyset, individual command errors should be concatenated.
			if isValidConftestOutput(cmdOutput) {
				combinedErr = multierror.Append(combinedErr, fmt.Errorf("policy_set: %s: conftest: some policies failed", policySet.Name))
			} else {
				combinedErr = multierror.Append(combinedErr, fmt.Errorf("policy_set: %s: conftest: %s", policySet.Name, cmdOutput))
			}
		}

		passed := true
		if hasFailures(cmdOutput) {
			passed = false
		}

		policySetResults = append(policySetResults, models.PolicySetResult{
			PolicySetName:  policySet.Name,
			ConftestOutput: cmdOutput,
			Passed:         passed,
			ReqApprovals:   policySet.ApproveCount,
		})
	}

	if policySetResults == nil {
		ctx.Log.Warn("no policies have been configured.")
		return "", nil
		// TODO: enable when we can pass policies in otherwise e2e tests with policy checks fail
		// return "", errors.Wrap(err, "building args")
	}

	marshaledStatus, err := json.Marshal(policySetResults)
	if err != nil {
		return "", errors.New("cannot marshal data into []PolicySetResult. data")
	}

	// Write policy check results to a file which can be used by custom workflow run steps for metrics, notifications, etc.
	policyCheckResultFile := filepath.Join(workdir, ctx.GetPolicyCheckResultFileName())
	err = os.WriteFile(policyCheckResultFile, marshaledStatus, 0600)

	combinedErr = multierror.Append(combinedErr, err)

	// Multierror will wrap combined errors in a way that the upstream functions won't be able to read it as nil.
	// Let's pass nil back if there are no wrapped errors.
	if errors.Unwrap(combinedErr) == nil {
		combinedErr = nil
	}

	output := string(marshaledStatus)

	return c.sanitizeOutput(inputFile, output), combinedErr

}

func (c *ConfTestExecutorWorkflow) sanitizeOutput(inputFile string, output string) string {
	return strings.Replace(output, inputFile, "<redacted plan file>", -1)
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
		return nil, errors.Wrapf(err, "wrapping version %s", defaultVersion)
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
