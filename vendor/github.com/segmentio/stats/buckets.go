package stats

import "strings"

// Key is a type used to uniquely identify metrics.
type Key struct {
	Measure string
	Field   string
}

// HistogramBuckets is a map type storing histogram buckets.
type HistogramBuckets map[Key][]Value

// Set sets a set of buckets to the given list of sorted values.
func (b HistogramBuckets) Set(key string, buckets ...interface{}) {
	v := make([]Value, len(buckets))

	for i, b := range buckets {
		v[i] = ValueOf(b)
	}

	b[makeKey(key)] = v
}

// Buckets is a registry where histogram buckets are placed. Some metric
// collection backends need to have histogram buckets defined by the program
// (like Prometheus), a common pattern is to use the init function of a package
// to register buckets for the various histograms that it produces.
var Buckets = HistogramBuckets{}

func makeKey(s string) Key {
	measure, field := splitMeasureField(s)
	return Key{Measure: measure, Field: field}
}

func splitMeasureField(s string) (measure string, field string) {
	if i := strings.LastIndexByte(s, ':'); i >= 0 {
		measure, field = s[:i], s[i+1:]
	} else {
		measure = s
	}
	return
}
