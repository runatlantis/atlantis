package config

import (
	"io"
	"net/url"

	"github.com/palantir/go-githubapp/githubapp"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally/v4"
)

type AuthConfig struct {
	SslCertFile string
	SslKeyFile  string
}

type ServerConfig struct {
	URL     *url.URL
	Version string
	Port    int
}

type TerraformConfig struct {
	DefaultVersionStr      string
	DefaultVersionFlagName string
	DownloadURL            string
	LogFilters             valid.TerraformLogFilters
}

// Config is TemporalWorker specific user config
type Config struct {
	AuthCfg      AuthConfig
	ServerCfg    ServerConfig
	TemporalCfg  valid.Temporal
	TerraformCfg TerraformConfig
	JobCfg       valid.Jobs

	DataDir     string
	CtxLogger   logging.Logger
	Scope       tally.Scope
	App         githubapp.Config
	StatsCloser io.Closer
}
