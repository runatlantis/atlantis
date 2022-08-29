package temporal

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/client"
	temporaltally "go.temporal.io/sdk/contrib/tally"
	"go.temporal.io/sdk/workflow"
	"logur.dev/logur"
)

func NewClient(scope tally.Scope, logger logur.Logger, cfg valid.Temporal) (client.Client, error) {
	opts := client.Options{
		Namespace:          cfg.Namespace,
		MetricsHandler:     temporaltally.NewMetricsHandler(scope),
		Logger:             logur.LoggerToKV(logger),
		ContextPropagators: []workflow.ContextPropagator{&ctxPropagator{}},
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
