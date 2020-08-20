package datadog

import (
	"fmt"
	"sync"

	"github.com/segmentio/stats"
)

// MetricType is an enumeration providing symbols to represent the different
// metric types upported by datadog.
type MetricType string

const (
	Counter   MetricType = "c"
	Gauge     MetricType = "g"
	Histogram MetricType = "h"
	Unknown   MetricType = "?"
)

// The Metric type is a representation of the metrics supported by datadog.
type Metric struct {
	Type      MetricType  // the metric type
	Namespace string      // the metric namespace (never populated by parsing operations)
	Name      string      // the metric name
	Value     float64     // the metric value
	Rate      float64     // sample rate, a value between 0 and 1
	Tags      []stats.Tag // the list of tags set on the metric
}

// String satisfies the fmt.Stringer interface.
func (m Metric) String() string {
	return fmt.Sprint(m)
}

// Format satisfies the fmt.Formatter interface.
func (m Metric) Format(f fmt.State, _ rune) {
	buf := bufferPool.Get().(*buffer)
	buf.b = appendMetric(buf.b[:0], m)
	f.Write(buf.b)
	bufferPool.Put(buf)
}

type buffer struct {
	b []byte
}

var bufferPool = sync.Pool{
	New: func() interface{} { return &buffer{make([]byte, 0, 512)} },
}
