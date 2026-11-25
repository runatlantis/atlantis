// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"time"

	statsd5 "github.com/cactus/go-statsd-client/v5/statsd"
	statsd6 "github.com/cactus/go-statsd-client/v6/statsd"
)

type statsdV6SenderAdapter struct {
	sender statsd6.StatSender
}

func (s statsdV6SenderAdapter) Inc(stat string, value int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Inc(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) Dec(stat string, value int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Dec(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) Gauge(stat string, value int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Gauge(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) GaugeDelta(stat string, value int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.GaugeDelta(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) Timing(stat string, delta int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Timing(stat, delta, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) TimingDuration(stat string, delta time.Duration, rate float32, tags ...statsd5.Tag) error {
	return s.sender.TimingDuration(stat, delta, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) Set(stat string, value string, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Set(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) SetInt(stat string, value int64, rate float32, tags ...statsd5.Tag) error {
	return s.sender.SetInt(stat, value, rate, convertStatsdTags(tags)...)
}

func (s statsdV6SenderAdapter) Raw(stat string, value string, rate float32, tags ...statsd5.Tag) error {
	return s.sender.Raw(stat, value, rate, convertStatsdTags(tags)...)
}

type statsdV6StatterAdapter struct {
	statsdV6SenderAdapter
	statter statsd6.Statter
}

var _ statsd5.Statter = statsdV6StatterAdapter{}

func newStatsdV6StatterAdapter(statter statsd6.Statter) statsd5.Statter {
	return statsdV6StatterAdapter{
		statsdV6SenderAdapter: statsdV6SenderAdapter{sender: statter},
		statter:               statter,
	}
}

func (s statsdV6StatterAdapter) NewSubStatter(prefix string) statsd5.SubStatter {
	sub := s.statter.NewSubStatter(prefix)
	return statsdV6SubStatterAdapter{
		statsdV6SenderAdapter: statsdV6SenderAdapter{sender: sub},
		subStatter:            sub,
	}
}

func (s statsdV6StatterAdapter) SetPrefix(prefix string) {
	s.statter.SetPrefix(prefix)
}

func (s statsdV6StatterAdapter) Close() error {
	return s.statter.Close()
}

type statsdV6SubStatterAdapter struct {
	statsdV6SenderAdapter
	subStatter statsd6.SubStatter
}

var _ statsd5.SubStatter = statsdV6SubStatterAdapter{}

func (s statsdV6SubStatterAdapter) SetSamplerFunc(fn statsd5.SamplerFunc) {
	s.subStatter.SetSamplerFunc(func(rate float32) bool {
		return fn(rate)
	})
}

func (s statsdV6SubStatterAdapter) NewSubStatter(prefix string) statsd5.SubStatter {
	sub := s.subStatter.NewSubStatter(prefix)
	return statsdV6SubStatterAdapter{
		statsdV6SenderAdapter: statsdV6SenderAdapter{sender: sub},
		subStatter:            sub,
	}
}

func convertStatsdTags(tags []statsd5.Tag) []statsd6.Tag {
	if len(tags) == 0 {
		return nil
	}
	converted := make([]statsd6.Tag, len(tags))
	for i, tag := range tags {
		converted[i] = statsd6.Tag(tag)
	}
	return converted
}
