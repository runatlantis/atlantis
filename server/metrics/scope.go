package metrics

import (
	"io"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/pkg/errors"
	"github.com/runatlantis/atlantis/server/events/yaml/valid"
	"github.com/runatlantis/atlantis/server/logging"
	"github.com/uber-go/tally"
	tallystatsd "github.com/uber-go/tally/statsd"
)

func NewLoggingScope(logger logging.SimpleLogging, statsNamespace string) (tally.Scope, io.Closer, error) {
	reporter, err := newReporter(valid.Metrics{}, logger)

	if err != nil {
		return nil, nil, errors.Wrap(err, "initializing stats reporter")
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:   statsNamespace,
		Reporter: reporter,
	}, time.Second)

	return scope, closer, nil
}

func NewScope(cfg valid.Metrics, logger logging.SimpleLogging, statsNamespace string) (tally.Scope, io.Closer, error) {
	reporter, err := newReporter(cfg, logger)

	if err != nil {
		return nil, nil, errors.Wrap(err, "initializing stats reporter")
	}

	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Prefix:   statsNamespace,
		Reporter: reporter,
	}, time.Second)

	return scope, closer, nil
}

func newReporter(cfg valid.Metrics, logger logging.SimpleLogging) (tally.StatsReporter, error) {
	if cfg.Statsd == nil {
		// return logging reporter and proceed
		return newLoggingReporter(logger), nil
	}

	statsdCfg := cfg.Statsd

	client, err := statsd.NewClientWithConfig(&statsd.ClientConfig{
		Address: strings.Join([]string{statsdCfg.Host, statsdCfg.Port}, ":"),
	})

	if err != nil {
		return nil, errors.Wrap(err, "initializing statsd client")
	}

	return tallystatsd.NewReporter(client, tallystatsd.Options{}), nil
}
