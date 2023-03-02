package metrics

import (
	"time"

	"github.com/uber-go/tally/v4"
)

// newNoopReporter returns a tally reporter that does nothing.
func newNoopReporter() tally.StatsReporter {
	return &noopReporter{}
}

type noopReporter struct{}

// Capabilities interface.

func (r *noopReporter) Reporting() bool {
	return true
}

func (r *noopReporter) Tagging() bool {
	return true
}

func (r *noopReporter) Capabilities() tally.Capabilities {
	return r
}

// Reporter interface.

func (r *noopReporter) Flush() {
	// Silence.
}

func (r *noopReporter) ReportCounter(name string, tags map[string]string, value int64) {
	// Silence.
}

func (r *noopReporter) ReportGauge(name string, tags map[string]string, value float64) {
	// Silence.
}

func (r *noopReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	// Silence.
}

func (r *noopReporter) ReportHistogramValueSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	// Silence.
}

func (r *noopReporter) ReportHistogramDurationSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	// Silence.
}
