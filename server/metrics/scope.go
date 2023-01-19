package metrics

import (
	"io"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally"
	tallyprom "github.com/uber-go/tally/prometheus"
	tallystatsd "github.com/uber-go/tally/statsd"
)

func NewLoggingScope(logger logging.SimpleLogging, statsNamespace string) (tally.Scope, io.Closer, error) {
	scope, _, closer, err := NewScope(valid.Metrics{}, logger, statsNamespace)
	return scope, closer, err
}

func NewScope(cfg valid.Metrics, logger logging.SimpleLogging, statsNamespace string) (tally.Scope, tally.BaseStatsReporter, io.Closer, error) {
	reporter, err := newReporter(cfg, logger)

	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "initializing stats reporter")
	}

	scopeOpts := tally.ScopeOptions{
		Prefix: statsNamespace,
	}

	if r, ok := reporter.(tally.StatsReporter); ok {
		scopeOpts.Reporter = r
	} else if r, ok := reporter.(tally.CachedStatsReporter); ok {
		scopeOpts.CachedReporter = r
		scopeOpts.Separator = tallyprom.DefaultSeparator
	}

	scope, closer := tally.NewRootScope(scopeOpts, time.Second)
	return scope, reporter, closer, nil
}

func newReporter(cfg valid.Metrics, logger logging.SimpleLogging) (tally.BaseStatsReporter, error) {

	// return statsd metrics if configured
	if cfg.Statsd != nil {
		return newStatsReporter(cfg)
	}

	// return prometheus metrics if configured
	if cfg.Prometheus != nil {
		return tallyprom.NewReporter(tallyprom.Options{}), nil
	}

	// return logging reporter and proceed
	return newLoggingReporter(logger), nil

}

func newStatsReporter(cfg valid.Metrics) (tally.StatsReporter, error) {

	statsdCfg := cfg.Statsd

	client, err := statsd.NewClientWithConfig(&statsd.ClientConfig{
		Address: strings.Join([]string{statsdCfg.Host, statsdCfg.Port}, ":"),
	})

	if err != nil {
		return nil, errors.Wrap(err, "initializing statsd client")
	}

	return tallystatsd.NewReporter(client, tallystatsd.Options{}), nil
}
