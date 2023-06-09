package metrics

import (
	"time"

	"github.com/runatlantis/atlantis/server/logging"
	tally "github.com/uber-go/tally/v4"
)

// newLoggingReporter returns a tally reporter that logs to the provided logger at debug level. This is useful for
// local development where the usual sinks are not available.
func newLoggingReporter(logger logging.SimpleLogging) tally.StatsReporter {
	return &debugReporter{log: logger}
}

type debugReporter struct {
	log logging.SimpleLogging
}

// Capabilities interface.

func (r *debugReporter) Reporting() bool {
	return true
}

func (r *debugReporter) Tagging() bool {
	return true
}

func (r *debugReporter) Capabilities() tally.Capabilities {
	return r
}

// Reporter interface.

func (r *debugReporter) Flush() {
	// Silence.
}

func (r *debugReporter) ReportCounter(name string, tags map[string]string, value int64) {
	log := r.log.With("name", name, "value", value, "tags", tags, "type", "counter")
	log.Debug("counter")
}

func (r *debugReporter) ReportGauge(name string, tags map[string]string, value float64) {
	log := r.log.With("name", name, "value", value, "tags", tags, "type", "gauge")
	log.Debug("gauge")
}

func (r *debugReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	log := r.log.With("name", name, "value", interval, "tags", tags, "type", "timer")
	log.Debug("timer")
}

func (r *debugReporter) ReportHistogramValueSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	log := r.log.With(
		"name", name,
		"buckets", buckets.AsValues(),
		"bucketLowerBound", bucketLowerBound,
		"bucketUpperBound", bucketUpperBound,
		"samples", samples,
		"tags", tags,
		"type", "valueHistogram",
	)
	log.Debug("histogram")
}

func (r *debugReporter) ReportHistogramDurationSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	log := r.log.With(
		"name", name,
		"buckets", buckets.AsValues(),
		"bucketLowerBound", bucketLowerBound,
		"bucketUpperBound", bucketUpperBound,
		"samples", samples,
		"tags", tags,
		"type", "durationHistogram",
	)
	log.Debug("histogram")
}
