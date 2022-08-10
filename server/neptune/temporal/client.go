package temporal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/client"
	temporaltally "go.temporal.io/sdk/contrib/tally"
	"logur.dev/logur"
)

var namespace = "atlantis"

func NewClient(scope tally.Scope, logger logur.Logger, cfg valid.Temporal) (client.Client, error) {
	opts := client.Options{
		HostPort:       fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Namespace:      namespace,
		MetricsHandler: temporaltally.NewMetricsHandler(scope),
		Logger:         logur.LoggerToKV(logger),
	}

	if cfg.UseSystemCACert {
		certs, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}
		opts.ConnectionOptions = client.ConnectionOptions{
			TLS: &tls.Config{
				RootCAs:    certs,
				MinVersion: tls.VersionTLS12,
			},
		}
	}

	if cfg.Host != "" || cfg.Port != "" {
		opts.HostPort = fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	}

	return client.Dial(opts)
}
