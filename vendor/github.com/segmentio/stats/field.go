package stats

import "strconv"

// A Field is a key/value type that represents a single metric in a Measure.
type Field struct {
	Name  string
	Value Value
}

// MakeField constructs and returns a new Field from name, value, and ftype.
func MakeField(name string, value interface{}, ftype FieldType) Field {
	f := Field{Name: name, Value: ValueOf(value)}
	f.setType(ftype)
	return f
}

// Type returns the type of f.
func (f Field) Type() FieldType {
	return FieldType(f.Value.pad)
}

func (f *Field) setType(t FieldType) {
	// We pack the field type into the value's padding space to make copies and
	// assignments of fields more time efficient.
	// Here are the results of a microbenchmark showing the performance of
	// a simple assignment for a Field type of 40 bytes (with a Type field) vs
	// an assignment of a Tag type (32 bytes).
	//
	// $ go test -v -bench . -run _
	// BenchmarkAssign40BytesStruct-4   	1000000000	         2.20 ns/op
	// BenchmarkAssign32BytesStruct-4   	2000000000	         0.31 ns/op
	//
	// There's an order of magnitude difference, so the optimization is worth it.
	f.Value.pad = int32(t)
}

func (f Field) String() string {
	return f.Type().String() + ":" + f.Name + "=" + f.Value.String()
}

// FieldType is an enumeration of the different metric types that may be set on
// a Field value.
type FieldType int32

const (
	// Counter represents incrementing counter metrics.
	Counter FieldType = iota

	// Gauge represents metrics that snapshot a value that may increase and
	// decrease.
	Gauge

	// Histogram represents metrics to observe the distribution of values.
	Histogram
)

func (t FieldType) String() string {
	switch t {
	case Counter:
		return "counter"
	case Gauge:
		return "gauge"
	case Histogram:
		return "histogram"
	}
	return ""
}

func (t FieldType) GoString() string {
	switch t {
	case Counter:
		return "stats.Counter"
	case Gauge:
		return "stats.Gauge"
	case Histogram:
		return "stats.Histogram"
	default:
		return "stats.FieldType(" + strconv.Itoa(int(t)) + ")"
	}
}

func copyFields(fields []Field) []Field {
	if len(fields) == 0 {
		return nil
	}
	cfields := make([]Field, len(fields))
	copy(cfields, fields)
	return cfields
}
