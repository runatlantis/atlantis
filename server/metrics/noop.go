package metrics

import (
	"io"
	"time"

	tally "github.com/uber-go/tally/v4"
)

// NewNoopScope creates a metrics scope that doesn't do any logging or reporting.
// This is useful for tests where metrics logging can slow down execution significantly.
func NewNoopScope() (tally.Scope, io.Closer, error) {
	reporter := &noopReporter{}
	
	scopeOpts := tally.ScopeOptions{
		Prefix:   "",
		Reporter: reporter,
	}
	
	scope, closer := tally.NewRootScope(scopeOpts, time.Second)
	return scope, closer, nil
}

// noopReporter implements tally.StatsReporter but does nothing.
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
	// No-op
}

func (r *noopReporter) ReportCounter(name string, tags map[string]string, value int64) {
	// No-op
}

func (r *noopReporter) ReportGauge(name string, tags map[string]string, value float64) {
	// No-op
}

func (r *noopReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	// No-op
}

func (r *noopReporter) ReportHistogramValueSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound float64,
	samples int64,
) {
	// No-op
}

func (r *noopReporter) ReportHistogramDurationSamples(
	name string,
	tags map[string]string,
	buckets tally.Buckets,
	bucketLowerBound,
	bucketUpperBound time.Duration,
	samples int64,
) {
	// No-op
}