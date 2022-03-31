package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	runtime_models "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"
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

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_conftest_client.go SourceResolver
// SourceResolver resolves the policy set to a local fs path
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
	binURL := fmt.Sprintf("%s/conftest_%s_%s_%s.tar.gz", versionURLPrefix, v.Original(), strings.Title(runtime.GOOS), conftestArch)
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
		log.Warn("failed to get default conftest version. Will attempt request scoped lazy loads %s", err.Error())
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

func (c *ConfTestExecutorWorkflow) Run(ctx models.ProjectCommandContext, executablePath string, envs map[string]string, workdir string, extraArgs []string) (string, error) {
	policyArgs := []Arg{}
	policySetNames := []string{}
	ctx.Log.Debug("policy sets, %s ", ctx.PolicySets)
	for _, policySet := range ctx.PolicySets.PolicySets {
		path, err := c.SourceResolver.Resolve(policySet)

		// Let's not fail the whole step because of a single failure. Log and fail silently
		if err != nil {
			ctx.Log.Err("Error resolving policyset %s. err: %s", policySet.Name, err.Error())
			continue
		}

		policyArg := NewPolicyArg(path)
		policyArgs = append(policyArgs, policyArg)

		policySetNames = append(policySetNames, policySet.Name)
	}

	inputFile := filepath.Join(workdir, ctx.GetShowResultFileName())

	args := ConftestTestCommandArgs{
		PolicyArgs: policyArgs,
		ExtraArgs:  extraArgs,
		InputFile:  inputFile,
		Command:    executablePath,
	}

	serializedArgs, err := args.build()

	if err != nil {
		ctx.Log.Warn("No policies have been configured")
		return "", nil
		// TODO: enable when we can pass policies in otherwise e2e tests with policy checks fail
		// return "", errors.Wrap(err, "building args")
	}

	initialOutput := fmt.Sprintf("Checking plan against the following policies: \n  %s\n", strings.Join(policySetNames, "\n  "))
	cmdOutput, err := c.Exec.CombinedOutput(serializedArgs, envs, workdir)

	return c.sanitizeOutput(inputFile, initialOutput+cmdOutput), err

}

func (c *ConfTestExecutorWorkflow) sanitizeOutput(inputFile string, output string) string {
	return strings.Replace(output, inputFile, "<redacted plan file>", -1)
}

func (c *ConfTestExecutorWorkflow) EnsureExecutorVersion(log logging.SimpleLogging, v *version.Version) (string, error) {
	// we have no information to proceed so fail hard
	if c.DefaultConftestVersion == nil && v == nil {
		return "", errors.New("no conftest version configured/specified")
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
