package activities

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-version"
	"github.com/palantir/go-githubapp/githubapp"
	"github.com/pkg/errors"
	legacy_tf "github.com/runatlantis/atlantis/server/core/terraform"
	"github.com/runatlantis/atlantis/server/neptune/github"
	"github.com/runatlantis/atlantis/server/neptune/temporalworker/config"
	"github.com/runatlantis/atlantis/server/neptune/terraform"
	repo "github.com/runatlantis/atlantis/server/neptune/workflows/internal/github"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/github/link"
	"github.com/runatlantis/atlantis/server/neptune/workflows/internal/root"
	"github.com/uber-go/tally/v4"
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

func NewDeploy(config githubapp.Config, scope tally.Scope) (*Deploy, error) {
	return &Deploy{
		dbActivities: &dbActivities{},
	}, nil
}

type Terraform struct {
	*terraformActivities
	*executeCommandActivities
	*workerInfoActivity
	*notifyActivities
	*cleanupActivities
}

func NewTerraform(config config.TerraformConfig, dataDir string, serverURL *url.URL, streamHandler streamHandler) (*Terraform, error) {
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

	tfClient, err := terraform.NewAsyncClient(
		tfClientConfig,
		config.DefaultVersionStr,
		&legacy_tf.DefaultDownloader{},
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
	}, nil
}

type Github struct {
	*githubActivities
}

type LinkBuilder interface {
	BuildDownloadLinkFromArchive(archiveURL *url.URL, root root.Root, repo repo.Repo, revision string) string
}

func NewGithub(config githubapp.Config, scope tally.Scope, dataDir string) (*Github, error) {
	clientCreator, err := githubapp.NewDefaultCachingClientCreator(
		config,
		githubapp.WithClientMiddleware(
			github.ClientMetrics(scope.SubScope("app")),
		))

	if err != nil {
		return nil, errors.Wrap(err, "initializing client creator")
	}
	return &Github{
		githubActivities: &githubActivities{
			ClientCreator: clientCreator,
			DataDir:       dataDir,
			LinkBuilder:   link.Builder{},
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
