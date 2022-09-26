package metrics

import (
	"strings"
	"time"

	"github.com/uber-go/tally/v4"
)

type customTagReporter struct {
	tally.StatsReporter

	separator string
}

// https://github.com/influxdata/telegraf/blob/master/plugins/inputs/statsd/README.md#influx-statsd
func (r *customTagReporter) taggedName(name string, tags map[string]string) string {
	var b strings.Builder
	b.WriteString(name)
	for k, v := range tags {
		b.WriteString(r.separator)
		b.WriteString(replaceChars(k))
		b.WriteByte('=')
		b.WriteString(replaceChars(v))
	}
	return b.String()
}

func (r *customTagReporter) ReportCounter(name string, tags map[string]string, value int64) {
	r.StatsReporter.ReportCounter(r.taggedName(name, tags), nil, value)
}

func (r *customTagReporter) ReportGauge(name string, tags map[string]string, value float64) {
	r.StatsReporter.ReportGauge(r.taggedName(name, tags), nil, value)
}

func (r *customTagReporter) ReportTimer(name string, tags map[string]string, interval time.Duration) {
	r.StatsReporter.ReportTimer(r.taggedName(name, tags), nil, interval)
}

func (r *customTagReporter) ReportHistogramValueSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound float64, samples int64) {
	r.StatsReporter.ReportHistogramValueSamples(r.taggedName(name, tags), nil, buckets, bucketLowerBound, bucketUpperBound, samples)
}

func (r *customTagReporter) ReportHistogramDurationSamples(name string, tags map[string]string, buckets tally.Buckets, bucketLowerBound, bucketUpperBound time.Duration, samples int64) {
	r.StatsReporter.ReportHistogramDurationSamples(r.taggedName(name, tags), nil, buckets, bucketLowerBound, bucketUpperBound, samples)
}

// Replace problematic characters in tags.
func replaceChars(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '.', ':', '|', '-', '=':
			b.WriteByte('_')
		default:
			b.WriteByte(s[i])
		}
	}
	return b.String()
}