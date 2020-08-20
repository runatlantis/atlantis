package stats

import (
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Measure is a type that represents a single measure made by the application.
// Measures are identified by a name, a set of fields that define what has been
// instrumented, and a set of tags representing different dimensions of the
// measure.
//
// Implementations of the Handler interface receive lists of measures produced
// by the application, and assume the tags will be sorted.
type Measure struct {
	Name   string
	Fields []Field
	Tags   []Tag
}

// Clone creates and returns a deep copy of m. The original and returned values
// and do not share any pointers to mutable types (but may share string values
// for example).
func (m Measure) Clone() Measure {
	return Measure{
		Name:   m.Name,
		Fields: copyFields(m.Fields),
		Tags:   copyTags(m.Tags),
	}
}

func (m Measure) String() string {
	return "{ " + m.Name + "(" + strings.Join(stringFields(m.Fields), ", ") + ") [" + strings.Join(stringTags(m.Tags), ", ") + "] }"
}

func stringFields(fields []Field) []string {
	s := make([]string, len(fields))

	for i, f := range fields {
		s[i] = f.String()
	}

	return s
}

func stringTags(tags []Tag) []string {
	s := make([]string, len(tags))

	for i, t := range tags {
		s[i] = t.String()
	}

	return s
}

// MakeMeasures takes a struct value or a pointer to a struct value as argument
// and extracts and returns the list of measures that it represented.
//
// The rules for converting values to measure are:
//
//  1. All fields exposing a 'metric' tag are expected to be of type bool, int,
//  int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr,
//  float32, float64, or time.Duration, and represent fields of the measures.
//  The struct fields may also define a 'type' tag with a value of "counter",
//  "gauge" or "histogram" to tune the behavior of the measure handlers.
//
//  2. All fields exposing a 'tag' tag are expected to be of type string and
//  represent tags of the measures.
//
//  3. All struct fields are searched recursively for fields matching rule (1)
//  and (2). Tags found within a struct are inherited by measures generated from
//  sub-fields, they may also be overwritten.
//
func MakeMeasures(prefix string, value interface{}, tags ...Tag) []Measure {
	if !TagsAreSorted(tags) {
		SortTags(tags)
	}
	return makeMeasures(nil, prefix, reflect.ValueOf(value), tags...)
}

func makeMeasures(cache *measureCache, prefix string, value reflect.Value, tags ...Tag) []Measure {
	return appendMeasures(nil, cache, prefix, value, tags...)
}

func appendMeasures(m []Measure, cache *measureCache, prefix string, v reflect.Value, tags ...Tag) []Measure {
	var p reflect.Value
	// The optimized routines for generating Measure values need to have the
	// address of the value, which means it has to be addressable. In the event
	// where it's not we have to make a copy of the value to be able to safely
	// get a pointer.
	switch {
	case v.Kind() == reflect.Ptr:
		p = v
		v = v.Elem()
	case v.CanAddr():
		p = v.Addr()
	default:
		p = reflect.New(v.Type())
		p.Elem().Set(v)
	}

	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		for i, n := 0, v.Len(); i != n; i++ {
			m = appendMeasures(m, cache, prefix, v.Index(i), tags...)
		}
		return m
	}

	var ptr = unsafe.Pointer(p.Pointer())
	var typ = v.Type()
	var mf []measureFuncs
	var ok bool

	if cache != nil {
		mf, ok = cache.lookup(typ)
	}

	if !ok {
		mf = makeMeasureFuncs(typ, prefix)
		if cache != nil {
			cache.set(typ, mf)
		}
	}

	used := len(m)
	free := cap(m) - used
	need := len(mf)

	if free < need {
		m = append(make([]Measure, 0, used+need), m...)
	}

	m = m[:used+need]

	for i := range mf {
		m[used+i].set(ptr, mf[i], tags...)
	}

	return m
}

func (m *Measure) set(ptr unsafe.Pointer, mf measureFuncs, tags ...Tag) {
	m.Name = mf.name

	if cap(m.Fields) < len(mf.fields) {
		m.Fields = make([]Field, len(mf.fields))
	} else {
		m.Fields = m.Fields[:len(mf.fields)]
	}

	for i := range m.Fields {
		m.Fields[i] = mf.fields[i](ptr)
	}

	n0 := cap(m.Tags)
	n1 := len(mf.tags)
	n2 := len(tags)
	n3 := n1 + n2

	if n0 < n3 {
		m.Tags = make([]Tag, n3)
	} else {
		m.Tags = m.Tags[:n3]
	}

	i1 := 0
	i2 := 0

	for i := range m.Tags {
		switch {
		case i1 == n1:
			m.Tags[i] = tags[i2]
			i2++

		case i2 == n2:
			m.Tags[i] = mf.tags[i1](ptr)
			i1++

		default:
			t1 := mf.tags[i1](ptr)
			t2 := tags[i2]

			if t1.Name < t2.Name {
				m.Tags[i] = t1
				i1++
			} else {
				m.Tags[i] = t2
				i2++
			}
		}
	}
}

func (m *Measure) reset() {
	for i := range m.Fields {
		m.Fields[i] = Field{}
	}

	for i := range m.Tags {
		m.Tags[i] = Tag{}
	}

	m.Name = ""
	m.Fields = m.Fields[:0]
	m.Tags = m.Tags[:0]
}

type measureFuncs struct {
	name   string
	fields []func(unsafe.Pointer) Field
	tags   []func(unsafe.Pointer) Tag
}

func makeMeasureFuncs(typ reflect.Type, prefix string) []measureFuncs {
	return appendMeasureFuncs(nil, typ, prefix, nil, 0)
}

func appendMeasureFuncs(measures []measureFuncs, typ reflect.Type, name string, tags tagFuncMap, offset uintptr) []measureFuncs {
	tags = tags.copy()
	kind := typ.Kind()

	switch kind {
	case reflect.Struct, reflect.Array:
	default:
		return measures
	}

	if kind == reflect.Array {
		elemType := typ.Elem()
		elemSize := elemType.Size()

		for ai, an := 0, typ.Len(); ai < an; ai++ {
			measures = appendMeasureFuncs(measures, elemType, name, tags, offset+(uintptr(ai)*elemSize))
		}

		return measures
	}

	for i, n := 0, typ.NumField(); i != n; i++ {
		field := typ.Field(i)

		if tag := field.Tag.Get("tag"); len(tag) != 0 {
			switch field.Type {
			case stringType:
				sf := structField{typ: field.Type, off: offset + field.Offset}
				tags[tag] = makeTagFunc(sf, tag)
			default:
				panic("unsupported value type found for metric tags of " + concat(name, tag) + ": " + field.Type.String())
			}
		}
	}

	mf := measureFuncs{name: name, tags: tags.funcs()}

	for i, n := 0, typ.NumField(); i != n; i++ {
		field := typ.Field(i)
		metric := field.Tag.Get("metric")

		switch field.Type.Kind() {
		case reflect.Struct, reflect.Array:
			measures = appendMeasureFuncs(measures, field.Type, concat(name, metric), tags, offset+field.Offset)

		default:
			if len(metric) != 0 {
				sf := structField{typ: field.Type, off: offset + field.Offset}
				t := makeFieldType(field.Tag.Get("type"))
				f := makeFieldFunc(sf, metric, t)
				if f == nil {
					panic("unsupported value type found for metric " + concat(name, metric) + ": " + field.Type.String())
				}
				mf.fields = append(mf.fields, f)
			}
		}
	}

	if len(mf.fields) != 0 {
		measures = append(measures, mf)
	}

	return measures
}

func makeFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	switch sf.typ {
	case boolType:
		return makeBoolFieldFunc(sf, name, ftype)
	case intType:
		return makeIntFieldFunc(sf, name, ftype)
	case int8Type:
		return makeInt8FieldFunc(sf, name, ftype)
	case int16Type:
		return makeInt16FieldFunc(sf, name, ftype)
	case int32Type:
		return makeInt32FieldFunc(sf, name, ftype)
	case int64Type:
		return makeInt64FieldFunc(sf, name, ftype)
	case uintType:
		return makeUintFieldFunc(sf, name, ftype)
	case uint8Type:
		return makeUint8FieldFunc(sf, name, ftype)
	case uint16Type:
		return makeUint16FieldFunc(sf, name, ftype)
	case uint32Type:
		return makeUint32FieldFunc(sf, name, ftype)
	case uint64Type:
		return makeUint64FieldFunc(sf, name, ftype)
	case uintptrType:
		return makeUintptrFieldFunc(sf, name, ftype)
	case float32Type:
		return makeFloat32FieldFunc(sf, name, ftype)
	case float64Type:
		return makeFloat64FieldFunc(sf, name, ftype)
	case durationType:
		return makeDurationFieldFunc(sf, name, ftype)
	default:
		return nil
	}
}

func makeBoolFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return boolValue(sf.bool(ptr)) })
}

func makeIntFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return intValue(sf.int(ptr)) })
}

func makeInt8FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return int8Value(sf.int8(ptr)) })
}

func makeInt16FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return int16Value(sf.int16(ptr)) })
}

func makeInt32FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return int32Value(sf.int32(ptr)) })
}

func makeInt64FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return int64Value(sf.int64(ptr)) })
}

func makeUintFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uintValue(sf.uint(ptr)) })
}

func makeUint8FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uint8Value(sf.uint8(ptr)) })
}

func makeUint16FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uint16Value(sf.uint16(ptr)) })
}

func makeUint32FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uint32Value(sf.uint32(ptr)) })
}

func makeUint64FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uint64Value(sf.uint64(ptr)) })
}

func makeUintptrFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return uintptrValue(sf.uintptr(ptr)) })
}

func makeFloat32FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return float32Value(sf.float32(ptr)) })
}

func makeFloat64FieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return float64Value(sf.float64(ptr)) })
}

func makeDurationFieldFunc(sf structField, name string, ftype FieldType) func(unsafe.Pointer) Field {
	return makeAnyFieldFunc(name, ftype, func(ptr unsafe.Pointer) Value { return durationValue(sf.duration(ptr)) })
}

func makeAnyFieldFunc(name string, ftype FieldType, valueOf func(unsafe.Pointer) Value) func(unsafe.Pointer) Field {
	return func(ptr unsafe.Pointer) Field {
		f := Field{Name: name, Value: valueOf(ptr)}
		f.setType(ftype)
		return f
	}
}

func makeFieldType(mtype string) FieldType {
	switch mtype {
	case "counter":
		return Counter
	case "gauge":
		return Gauge
	default:
		return Histogram
	}
}

type tagFuncMap map[string](func(unsafe.Pointer) Tag)

func (tags tagFuncMap) copy() tagFuncMap {
	cpy := make(tagFuncMap, len(tags))

	for name, fn := range tags {
		cpy[name] = fn
	}

	return cpy
}

func (tags tagFuncMap) namedTagFuncs() []namedTagFunc {
	namedTags := make([]namedTagFunc, 0, len(tags))

	for name, fn := range tags {
		namedTags = append(namedTags, namedTagFunc{name: name, fn: fn})
	}

	return namedTags
}

func (tags tagFuncMap) funcs() []func(unsafe.Pointer) Tag {
	namedTags := tags.namedTagFuncs()
	sort.Sort(tagFuncByName(namedTags))

	fns := make([]func(unsafe.Pointer) Tag, len(namedTags))

	for i, tag := range namedTags {
		fns[i] = tag.fn
	}

	return fns
}

type namedTagFunc struct {
	name string
	fn   func(unsafe.Pointer) Tag
}

func makeTagFunc(sf structField, name string) func(unsafe.Pointer) Tag {
	return func(ptr unsafe.Pointer) Tag { return Tag{Name: name, Value: sf.string(ptr)} }
}

type tagFuncByName []namedTagFunc

func (t tagFuncByName) Len() int               { return len(t) }
func (t tagFuncByName) Less(i int, j int) bool { return t[i].name < t[j].name }
func (t tagFuncByName) Swap(i int, j int)      { t[i], t[j] = t[j], t[i] }

func concat(prefix string, suffix string) string {
	if len(prefix) == 0 {
		return suffix
	}
	if len(suffix) == 0 {
		return prefix
	}
	return prefix + "." + suffix
}

type measuresBuffer struct {
	measures []Measure
}

var measurePool = sync.Pool{
	New: func() interface{} {
		return &measuresBuffer{measures: make([]Measure, 0, 32)}
	},
}

type measureCache struct {
	cache unsafe.Pointer
}

func (c *measureCache) lookup(typ reflect.Type) ([]measureFuncs, bool) {
	m := c.load()
	if m == nil {
		return nil, false
	}
	mf, ok := (*m)[typ]
	return mf, ok
}

func (c *measureCache) set(typ reflect.Type, mf []measureFuncs) {
	for {
		m1 := c.load()
		m2 := map[reflect.Type][]measureFuncs{typ: mf}

		if m1 != nil {
			for t, f := range *m1 {
				m2[t] = f
			}
		}

		if c.compareAndSwap(m1, &m2) {
			break
		}
	}
}

func (c *measureCache) load() *map[reflect.Type][]measureFuncs {
	return (*map[reflect.Type][]measureFuncs)(atomic.LoadPointer(&c.cache))
}

func (c *measureCache) compareAndSwap(old *map[reflect.Type][]measureFuncs, new *map[reflect.Type][]measureFuncs) bool {
	return atomic.CompareAndSwapPointer(&c.cache,
		unsafe.Pointer(old),
		unsafe.Pointer(new),
	)
}
