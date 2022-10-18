package activities

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/core/runtime/cache"
	legacy_tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/neptune/storage"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/config"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/deployment"
	internal "github.com/runatlantis/atlantis/server/neptune/workflows/activities/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/github/link"
	"github.com/runatlantis/atlantis/server/neptune/workflows/activities/terraform"
)

const (
	// binDirName is the name of the directory inside our data dir where
	// we download binaries.
	BinDirName = "bin"
	// terraformPluginCacheDir is the name of the dir inside our data dir
	// where we tell terraform to cache plugins and modules.
	TerraformPluginCacheDirName = "plugin-cache"
)

// Exported Activites should be here.
// The convention should be one exported struct per workflow
// This guarantees function naming uniqueness within a given workflow
// which is a requirement at a per worker level
//
// Note: This doesn't prevent issues with naming duplication that can come up when
// registering multiple workflows to the same worker
type Deploy struct {
	*dbActivities
}

func NewDeploy(deploymentStoreCfg valid.StoreConfig) (*Deploy, error) {

	storageClient, err := storage.NewClient(deploymentStoreCfg)
	if err != nil {
		return nil, errors.Wrap(err, "intializing stow client")
	}

	deploymentStore, err := deployment.NewStore(storageClient)
	if err != nil {
		return nil, errors.Wrap(err, "initializing deployment info store")
	}

	return &Deploy{
		dbActivities: &dbActivities{
			DeploymentInfoStore: deploymentStore,
		},
	}, nil
}

type Terraform struct {
	*terraformActivities
	*executeCommandActivities
	*workerInfoActivity
	*cleanupActivities
	*jobActivities
}

type StreamCloser interface {
	streamer
	closer
}

type TerraformOptions struct {
	VersionCache cache.ExecutionVersionCache
}

func NewTerraform(config config.TerraformConfig, dataDir string, serverURL *url.URL, streamHandler StreamCloser, opts ...TerraformOptions) (*Terraform, error) {
	binDir, err := mkSubDir(dataDir, BinDirName)
	if err != nil {
		return nil, err
	}

	cacheDir, err := mkSubDir(dataDir, TerraformPluginCacheDirName)
	if err != nil {
		return nil, err
	}

	defaultTfVersion, err := version.NewVersion(config.DefaultVersionStr)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing version %s", config.DefaultVersionStr)
	}

	tfClientConfig := terraform.ClientConfig{
		BinDir:        binDir,
		CacheDir:      cacheDir,
		TfDownloadURL: config.DownloadURL,
	}

	downloader := &legacy_tf.DefaultDownloader{}

	loader := legacy_tf.NewVersionLoader(downloader, config.DownloadURL)

	var versionCache cache.ExecutionVersionCache
	for _, o := range opts {
		versionCache = o.VersionCache
	}

	if versionCache == nil {
		versionCache = cache.NewExecutionVersionLayeredLoadingCache(
			"terraform",
			binDir,
			loader.LoadVersion,
		)
	}

	tfClient, err := terraform.NewAsyncClient(
		tfClientConfig,
		config.DefaultVersionStr,
		downloader,
		versionCache,
	)
	if err != nil {
		return nil, err
	}

	return &Terraform{
		executeCommandActivities: &executeCommandActivities{},
		workerInfoActivity: &workerInfoActivity{
			ServerURL: serverURL,
		},
		terraformActivities: &terraformActivities{
			TerraformClient:  tfClient,
			StreamHandler:    streamHandler,
			DefaultTFVersion: defaultTfVersion,
		},
		jobActivities: &jobActivities{
			StreamCloser: streamHandler,
		},
	}, nil
}

type Github struct {
	*githubActivities
}

type LinkBuilder interface {
	BuildDownloadLinkFromArchive(archiveURL *url.URL, root terraform.Root, repo internal.Repo, revision string) string
}

func NewGithub(client githubClient, dataDir string, getter gogetter) (*Github, error) {
	return &Github{
		githubActivities: &githubActivities{
			Client:      client,
			DataDir:     dataDir,
			LinkBuilder: link.Builder{},
			Getter:      getter,
		},
	}, nil
}

func mkSubDir(parentDir string, subDir string) (string, error) {
	fullDir := filepath.Join(parentDir, subDir)
	if err := os.MkdirAll(fullDir, 0700); err != nil {
		return "", errors.Wrapf(err, "unable to creare dir %q", fullDir)
	}

	return fullDir, nil
}
