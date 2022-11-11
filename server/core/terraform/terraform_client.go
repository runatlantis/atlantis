// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// Modified hereafter by contributors to runatlantis/atlantis.
//
// Package terraform handles the actual running of terraform commands.
package terraform

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"

	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	runtime_models "github.com/runatlantis/atlantis/server/core/runtime/models"
	"github.com/runatlantis/atlantis/server/events/command"
	"github.com/runatlantis/atlantis/server/events/terraform/ansi"
	"github.com/runatlantis/atlantis/server/jobs"
	"github.com/runatlantis/atlantis/server/logging"
)

var LogStreamingValidCmds = [...]string{"init", "plan", "apply"}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_terraform_client.go Client

type Client interface {
	// RunCommandWithVersion executes terraform with args in path. If v is nil,
	// it will use the default Terraform version. workspace is the Terraform
	// workspace which should be set as an environment variable.
	RunCommandWithVersion(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, envs map[string]string, v *version.Version, workspace string) (string, error)

	// EnsureVersion makes sure that terraform version `v` is available to use
	EnsureVersion(log logging.Logger, v *version.Version) error
}

type DefaultClient struct {
	// defaultVersion is the default version of terraform to use if another
	// version isn't specified.
	defaultVersion *version.Version
	binDir         string
	// downloader downloads terraform versions.
	downloader      Downloader
	downloadBaseURL string

	versionCache   cache.ExecutionVersionCache
	commandBuilder commandBuilder
	*AsyncClient
}

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_downloader.go Downloader

// Downloader is for downloading terraform versions.
type Downloader interface {
	GetFile(dst, src string, opts ...getter.ClientOption) error
	GetAny(dst, src string, opts ...getter.ClientOption) error
}

// versionRegex extracts the version from `terraform version` output.
//
//	    Terraform v0.12.0-alpha4 (2c36829d3265661d8edbd5014de8090ea7e2a076)
//		   => 0.12.0-alpha4
//
//	    Terraform v0.11.10
//		   => 0.11.10
var versionRegex = regexp.MustCompile("Terraform v(.*?)(\\s.*)?\n")

// NewClientWithDefaultVersion creates a new terraform client and pre-fetches the default version
func NewClientWithVersionCache(
	binDir string,
	cacheDir string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloader Downloader,
	usePluginCache bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
	versionCache cache.ExecutionVersionCache,
) (*DefaultClient, error) {
	version, err := getDefaultVersion(defaultVersionStr, defaultVersionFlagName)

	if err != nil {
		return nil, errors.Wrapf(err, "getting default version")
	}

	// warm the cache with this version
	_, err = versionCache.Get(version)

	if err != nil {
		return nil, errors.Wrapf(err, "getting default terraform version %s", defaultVersionStr)
	}

	builder := &CommandBuilder{
		defaultVersion: version,
		versionCache:   versionCache,
	}

	if usePluginCache {
		builder.terraformPluginCacheDir = cacheDir
	}

	asyncClient := &AsyncClient{
		projectCmdOutputHandler: projectCmdOutputHandler,
		commandBuilder:          builder,
	}

	return &DefaultClient{
		defaultVersion:  version,
		binDir:          binDir,
		downloader:      tfDownloader,
		downloadBaseURL: tfDownloadURL,
		AsyncClient:     asyncClient,
		commandBuilder:  builder,
		versionCache:    versionCache,
	}, nil

}

func NewE2ETestClient(
	binDir string,
	cacheDir string,
	tfeToken string,
	tfeHostname string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloader Downloader,
	usePluginCache bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
) (*DefaultClient, error) {
	versionCache := cache.NewLocalBinaryCache("terraform")
	return NewClientWithVersionCache(
		binDir,
		cacheDir,
		defaultVersionStr,
		defaultVersionFlagName,
		tfDownloadURL,
		tfDownloader,
		usePluginCache,
		projectCmdOutputHandler,
		versionCache,
	)
}

func NewClient(
	binDir string,
	cacheDir string,
	defaultVersionStr string,
	defaultVersionFlagName string,
	tfDownloadURL string,
	tfDownloader Downloader,
	usePluginCache bool,
	projectCmdOutputHandler jobs.ProjectCommandOutputHandler,
) (*DefaultClient, error) {
	loader := VersionLoader{
		downloader:  tfDownloader,
		downloadURL: tfDownloadURL,
	}

	versionCache := cache.NewExecutionVersionLayeredLoadingCache(
		"terraform",
		binDir,
		loader.LoadVersion,
	)
	return NewClientWithVersionCache(
		binDir,
		cacheDir,
		defaultVersionStr,
		defaultVersionFlagName,
		tfDownloadURL,
		tfDownloader,
		usePluginCache,
		projectCmdOutputHandler,
		versionCache,
	)
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

func (c *DefaultClient) EnsureVersion(log logging.Logger, v *version.Version) error {
	if v == nil {
		v = c.defaultVersion
	}

	_, err := c.versionCache.Get(v)

	if err != nil {
		return errors.Wrapf(err, "getting version %s", v)
	}

	return nil
}

// See Client.RunCommandWithVersion.
func (c *DefaultClient) RunCommandWithVersion(ctx context.Context, prjCtx command.ProjectContext, path string, args []string, customEnvVars map[string]string, v *version.Version, workspace string) (string, error) {
	// if the feature is enabled, we use the async workflow else we default to the original sync workflow
	// Don't stream terraform show output to outCh
	if len(args) > 0 && isAsyncEligibleCommand(args[0]) {
		outCh := c.RunCommandAsync(ctx, prjCtx, path, args, customEnvVars, v, workspace)

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

	cmd, err := c.commandBuilder.Build(v, workspace, path, args)
	if err != nil {
		return "", err
	}
	envVars := cmd.Env
	for key, val := range customEnvVars {
		envVars = append(envVars, fmt.Sprintf("%s=%s", key, val))
	}
	cmd.Env = envVars
	out, err := cmd.CombinedOutput()
	if err != nil {
		err = errors.Wrapf(err, "running %q in %q", cmd.String(), path)
		prjCtx.Log.ErrorContext(prjCtx.RequestCtx, err.Error())
		return ansi.Strip(string(out)), err
	}
	prjCtx.Log.InfoContext(prjCtx.RequestCtx, fmt.Sprintf("successfully ran %q in %q", cmd.String(), path))

	return ansi.Strip(string(out)), nil
}

type VersionLoader struct {
	downloader  Downloader
	downloadURL string
}

func NewVersionLoader(downloader Downloader, downloadURL string) *VersionLoader {
	return &VersionLoader{
		downloader:  downloader,
		downloadURL: downloadURL,
	}
}

func (l *VersionLoader) LoadVersion(v *version.Version, destPath string) (runtime_models.FilePath, error) {
	urlPrefix := fmt.Sprintf("%s/terraform/%s/terraform_%s", l.downloadURL, v.String(), v.String())
	binURL := fmt.Sprintf("%s_%s_%s.zip", urlPrefix, runtime.GOOS, runtime.GOARCH)
	checksumURL := fmt.Sprintf("%s_SHA256SUMS", urlPrefix)
	fullSrcURL := fmt.Sprintf("%s?checksum=file:%s", binURL, checksumURL)
	if err := l.downloader.GetAny(destPath, fullSrcURL); err != nil {
		return runtime_models.LocalFilePath(""), errors.Wrapf(err, "downloading terraform version %s at %q", v.String(), fullSrcURL)
	}

	binPath := filepath.Join(destPath, "terraform")

	return runtime_models.LocalFilePath(binPath), nil

}

func isAsyncEligibleCommand(cmd string) bool {
	for _, validCmd := range LogStreamingValidCmds {
		if validCmd == cmd {
			return true
		}
	}
	return false
}

func getDefaultVersion(overrideVersion string, versionFlagName string) (*version.Version, error) {
	if overrideVersion != "" {
		v, err := version.NewVersion(overrideVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing version %s", overrideVersion)
		}

		return v, nil
	}

	// look for the binary directly on disk and query the version
	// we shouldn't really be doing this, but don't want to break existing clients.
	// this implementation assumes that versions in the format our cache assumes
	// and if thats the case we won't be redownloading the version of this binary to our cache
	localPath, err := exec.LookPath("terraform")
	if err != nil {
		return nil, fmt.Errorf("terraform not found in $PATH. Set --%s or download terraform from https://www.terraform.io/downloads.html", versionFlagName)
	}

	return getVersion(localPath)
}

func getVersion(tfBinary string) (*version.Version, error) {
	versionOutBytes, err := exec.Command(tfBinary, "version").Output() // #nosec
	versionOutput := string(versionOutBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "running terraform version: %s", versionOutput)
	}
	match := versionRegex.FindStringSubmatch(versionOutput)
	if len(match) <= 1 {
		return nil, fmt.Errorf("could not parse terraform version from %s", versionOutput)
	}
	return version.NewVersion(match[1])
}

type DefaultDownloader struct{}

// See go-getter.GetFile.
func (d *DefaultDownloader) GetFile(dst, src string, opts ...getter.ClientOption) error {
	return getter.GetFile(dst, src, opts...)
}

// See go-getter.GetFile.
func (d *DefaultDownloader) GetAny(dst, src string, opts ...getter.ClientOption) error {
	return getter.GetAny(dst, src, opts...)
}
