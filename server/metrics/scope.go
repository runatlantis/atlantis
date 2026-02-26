// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
	tallyprom "github.com/uber-go/tally/v4/prometheus"
	tallystatsd "github.com/uber-go/tally/v4/statsd"
)

func NewLoggingScope(logger logging.SimpleLogging, statsNamespace string) (tally.Scope, io.Closer, error) {
	scope, _, closer, err := NewScope(valid.Metrics{}, logger, statsNamespace)
	return scope, closer, err
}

func NewScope(cfg valid.Metrics, logger logging.SimpleLogging, statsNamespace string) (tally.Scope, tally.BaseStatsReporter, io.Closer, error) {
	reporter, err := newReporter(cfg, logger)

	if err != nil {
		return nil, nil, nil, fmt.Errorf("initializing stats reporter: %w", err)
	}

	scopeOpts := tally.ScopeOptions{
		Prefix: statsNamespace,
		SanitizeOptions: &tally.SanitizeOptions{
			NameCharacters: tally.ValidCharacters{
				Ranges:     tally.AlphanumericRange,
				Characters: tally.UnderscoreCharacters,
			},
			KeyCharacters: tally.ValidCharacters{
				Ranges:     tally.AlphanumericRange,
				Characters: tally.UnderscoreCharacters,
			},
			ValueCharacters: tally.ValidCharacters{
				Ranges: []tally.SanitizeRange{
					{rune('a'), rune('z')},
					{rune('A'), rune('Z')},
					{rune('0'), rune('9')},
					{rune('-'), rune('.')},
				},
			},
			ReplacementCharacter: tally.DefaultReplacementCharacter,
		},
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
		return nil, fmt.Errorf("initializing statsd client: %w", err)
	}

	return tallystatsd.NewReporter(client, tallystatsd.Options{}), nil
}
