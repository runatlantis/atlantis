package policy

import (
	"fmt"
	"os"
	"runtime"
	"strings"

	version "github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/events/runtime/cache"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
)

const (
	defaultConftestVersionEnvKey = "DEFAULT_CONFTEST_VERSION"
	conftestBinaryName           = "conftest"
	conftestDownloadURLPrefix    = "https://github.com/open-policy-agent/conftest/releases/download/v"
	conftestArch                 = "x86_64"
)

// SourceResolver resolves the policy set to a local fs path
type SourceResolver interface {
	Resolve(policySet models.PolicySet) (string, error)
}

// LocalSourceResolver resolves a local policy set to a local fs path
type LocalSourceResolver struct {
}

func (p *LocalSourceResolver) Resolve(policySet models.PolicySet) (string, error) {
	return "some/path", nil

}

// SourceResolverProxy proxies to underlying source resolvers dynamically
type SourceResolverProxy struct {
	localSourceResolver SourceResolver
}

func (p *SourceResolverProxy) Resolve(policySet models.PolicySet) (string, error) {
	switch source := policySet.Source; source {
	case models.LocalPolicySet:
		return p.localSourceResolver.Resolve(policySet)
	default:
		return "", errors.New(fmt.Sprintf("unable to resolve policy set source %s", source))
	}
}

type ConfTestVersionDownloader struct {
	downloader terraform.Downloader
}

func (c ConfTestVersionDownloader) downloadConfTestVersion(v *version.Version, destPath string) error {
	versionURLPrefix := fmt.Sprintf("%s%s", conftestDownloadURLPrefix, v.Original())

	// download binary in addition to checksum file
	binURL := fmt.Sprintf("%s/conftest_%s_%s_%s.tar.gz", versionURLPrefix, v.Original(), strings.Title(runtime.GOOS), conftestArch)
	checksumURL := fmt.Sprintf("%s/checksums.txt", versionURLPrefix)

	// underlying implementation uses go-getter so the URL is formatted as such.
	// i know i know, I'm assuming an interface implementation with my inputs.
	// realistically though the interface just exists for testing so ¯\_(ツ)_/¯
	fullSrcURL := fmt.Sprintf("%s?checksum=file:%s", binURL, checksumURL)

	if err := c.downloader.GetFile(destPath, fullSrcURL); err != nil {
		return errors.Wrapf(err, "downloading conftest version %s at %q", v.String(), fullSrcURL)
	}

	return nil
}

// ConfTestExecutorWorkflow runs a versioned conftest binary with the args built from the project context.
// Project context defines whether conftest runs a local policy set or runs a test on a remote policy set.
type ConfTestExecutorWorkflow struct {
	SourceResolver         SourceResolver
	VersionCache           cache.ExecutionVersionCache
	DefaultConftestVersion *version.Version
}

func NewConfTestExecutorWorkflow(log *logging.SimpleLogger, versionRootDir string) *ConfTestExecutorWorkflow {
	downloader := ConfTestVersionDownloader{
		downloader: &terraform.DefaultDownloader{},
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
	}
}

func (c *ConfTestExecutorWorkflow) Run(log *logging.SimpleLogger, executablePath string, envs map[string]string, args []string) (string, error) {
	return "success", nil

}

func (c *ConfTestExecutorWorkflow) EnsureExecutorVersion(log *logging.SimpleLogger, v *version.Version) (string, error) {
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
	defaultVersion, exists := os.LookupEnv(defaultConftestVersionEnvKey)

	if !exists {
		return nil, errors.New(fmt.Sprintf("%s not set.", defaultConftestVersionEnvKey))
	}

	wrappedVersion, err := version.NewVersion(defaultVersion)

	if err != nil {
		return nil, errors.Wrapf(err, "wrapping version %s", defaultVersion)
	}
	return wrappedVersion, nil
}

func (c *ConfTestExecutorWorkflow) ResolveArgs(ctx models.ProjectCommandContext) ([]string, error) {
	return []string{""}, nil
}
