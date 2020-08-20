package stats

import (
	"os"
	"path/filepath"
	"reflect"
	"sync"
	"time"
)

// An Engine carries the context for producing metrics, it is configured by
// setting the exported fields or using the helper methods to create sub-engines
// that inherit the configuration of the base they were created from.
//
// The program must not modify the engine's handler, prefix, or tags after it
// started using it. If changes need to be made new engines must be created by
// calls to WithPrefix or WithTags.
type Engine struct {
	// The measure handler that the engine forwards measures to.
	Handler Handler

	// A prefix set on all metric names produced by the engine.
	Prefix string

	// A list of tags set on all metrics produced by the engine.
	//
	// The list of tags has to be sorted. This is automatically managed by the
	// helper methods WithPrefix, WithTags and the NewEngine function. A program
	// that manipulates this field directly has to respect this requirement.
	Tags []Tag

	// This cache keeps track of the generated measure structures to avoid
	// rebuilding them every time a same measure type is seen by the engine.
	//
	// The cached values include the engine prefix in the measure names, which
	// is why the cache must be local to the engine.
	cache measureCache
}

// NewEngine creates and returns a new engine configured with prefix, handler,
// and tags.
func NewEngine(prefix string, handler Handler, tags ...Tag) *Engine {
	return &Engine{
		Handler: handler,
		Prefix:  prefix,
		Tags:    SortTags(copyTags(tags)),
	}
}

// Register adds handler to eng.
func (eng *Engine) Register(handler Handler) {
	if eng.Handler == Discard {
		eng.Handler = handler
	} else {
		eng.Handler = MultiHandler(eng.Handler, handler)
	}
}

// Flush flushes eng's handler (if it implements the Flusher interface).
func (eng *Engine) Flush() {
	flush(eng.Handler)
}

// WithPrefix returns a copy of the engine with prefix appended to eng's current
// prefix and tags set to the merge of eng's current tags and those passed as
// argument. Both eng and the returned engine share the same handler.
func (eng *Engine) WithPrefix(prefix string, tags ...Tag) *Engine {
	return &Engine{
		Handler: eng.Handler,
		Prefix:  eng.makeName(prefix),
		Tags:    eng.makeTags(tags),
	}
}

// WithTags returns a copy of the engine with tags set to the merge of eng's
// current tags and those passed as arguments. Both eng and the returned engine
// share the same handler.
func (eng *Engine) WithTags(tags ...Tag) *Engine {
	return eng.WithPrefix("", tags...)
}

// Incr increments by one the counter identified by name and tags.
func (eng *Engine) Incr(name string, tags ...Tag) {
	eng.Add(name, 1, tags...)
}

// Incr increments by one the counter identified by name and tags.
func (eng *Engine) IncrAt(time time.Time, name string, tags ...Tag) {
	eng.AddAt(time, name, 1, tags...)
}

// Add increments by value the counter identified by name and tags.
func (eng *Engine) Add(name string, value interface{}, tags ...Tag) {
	eng.measure(time.Now(), name, value, Counter, tags...)
}

// Add increments by value the counter identified by name and tags.
func (eng *Engine) AddAt(t time.Time, name string, value interface{}, tags ...Tag) {
	eng.measure(t, name, value, Counter, tags...)
}

// Set sets to value the gauge identified by name and tags.
func (eng *Engine) Set(name string, value interface{}, tags ...Tag) {
	eng.measure(time.Now(), name, value, Gauge, tags...)
}

// Set sets to value the gauge identified by name and tags.
func (eng *Engine) SetAt(t time.Time, name string, value interface{}, tags ...Tag) {
	eng.measure(t, name, value, Gauge, tags...)
}

// Observe reports value for the histogram identified by name and tags.
func (eng *Engine) Observe(name string, value interface{}, tags ...Tag) {
	eng.measure(time.Now(), name, value, Histogram, tags...)
}

// Observe reports value for the histogram identified by name and tags.
func (eng *Engine) ObserveAt(t time.Time, name string, value interface{}, tags ...Tag) {
	eng.measure(t, name, value, Histogram, tags...)
}

// Clock returns a new clock identified by name and tags.
func (eng *Engine) Clock(name string, tags ...Tag) *Clock {
	return eng.ClockAt(name, time.Now(), tags...)
}

// ClockAt returns a new clock identified by name and tags with a specified
// start time.
func (eng *Engine) ClockAt(name string, start time.Time, tags ...Tag) *Clock {
	cpy := make([]Tag, len(tags), len(tags)+1) // clock always appends a stamp.
	copy(cpy, tags)
	return &Clock{
		name:  name,
		first: start,
		last:  start,
		tags:  cpy,
		eng:   eng,
	}
}

func (eng *Engine) measure(t time.Time, name string, value interface{}, ftype FieldType, tags ...Tag) {
	name, field := splitMeasureField(name)
	mp := measureArrayPool.Get().(*[1]Measure)

	m := &(*mp)[0]
	m.Name = eng.makeName(name) // TODO: figure out how to optimize this
	m.Fields = append(m.Fields[:0], MakeField(field, value, ftype))
	m.Tags = append(m.Tags[:0], eng.Tags...)
	m.Tags = append(m.Tags, tags...)

	if len(tags) != 0 && !TagsAreSorted(m.Tags) {
		SortTags(m.Tags)
	}

	eng.Handler.HandleMeasures(t, (*mp)[:]...)

	for i := range m.Fields {
		m.Fields[i] = Field{}
	}

	for i := range m.Tags {
		m.Tags[i] = Tag{}
	}

	m.Name = ""
	measureArrayPool.Put(mp)
}

func (eng *Engine) makeName(name string) string {
	return concat(eng.Prefix, name)
}

func (eng *Engine) makeTags(tags []Tag) []Tag {
	return SortTags(concatTags(eng.Tags, tags))
}

var measureArrayPool = sync.Pool{
	New: func() interface{} { return new([1]Measure) },
}

// Report calls ReportAt with time.Now() as first argument.
func (eng *Engine) Report(metrics interface{}, tags ...Tag) {
	eng.ReportAt(time.Now(), metrics, tags...)
}

// ReportAt reports a set of metrics for a given time. The metrics must be of
// type struct, pointer to struct, or a slice or array to one of those. See
// MakeMeasures for details about how to make struct types exposing metrics.
func (eng *Engine) ReportAt(time time.Time, metrics interface{}, tags ...Tag) {
	var tb *tagsBuffer

	if len(tags) == 0 {
		// fast path for the common case where there are no dynamic tags
		tags = eng.Tags
	} else {
		tb = tagsPool.Get().(*tagsBuffer)
		tb.append(tags...)
		tb.append(eng.Tags...)
		tb.sort()
		tags = tb.tags
	}

	mb := measurePool.Get().(*measuresBuffer)
	mb.measures = appendMeasures(mb.measures[:0], &eng.cache, eng.Prefix, reflect.ValueOf(metrics), tags...)

	ms := mb.measures
	eng.Handler.HandleMeasures(time, ms...)

	for i := range ms {
		ms[i].reset()
	}

	if tb != nil {
		tb.reset()
		tagsPool.Put(tb)
	}

	measurePool.Put(mb)
}

// DefaultEngine is the engine used by global helper functions.
var DefaultEngine = NewEngine(progname(), Discard)

// Register adds handler to the default engine.
func Register(handler Handler) {
	DefaultEngine.Register(handler)
}

// Flush flushes the default engine.
func Flush() {
	DefaultEngine.Flush()
}

// WithPrefix returns a copy of the engine with prefix appended to default
// engine's current prefix and tags set to the merge of eng's current tags and
// those passed as argument. Both the default engine and the returned engine
// share the same handler.
func WithPrefix(prefix string, tags ...Tag) *Engine {
	return DefaultEngine.WithPrefix(prefix, tags...)
}

// WithTags returns a copy of the engine with tags set to the merge of the
// default engine's current tags and those passed as arguments. Both the default
// engine and the returned engine share the same handler.
func WithTags(tags ...Tag) *Engine {
	return DefaultEngine.WithTags(tags...)
}

// Incr increments by one the counter identified by name and tags.
func Incr(name string, tags ...Tag) {
	DefaultEngine.Incr(name, tags...)
}

// IncrAt increments by one the counter identified by name and tags.
func IncrAt(time time.Time, name string, tags ...Tag) {
	DefaultEngine.IncrAt(time, name, tags...)
}

// Add increments by value the counter identified by name and tags.
func Add(name string, value interface{}, tags ...Tag) {
	DefaultEngine.Add(name, value, tags...)
}

// AddAt increments by value the counter identified by name and tags.
func AddAt(time time.Time, name string, value interface{}, tags ...Tag) {
	DefaultEngine.AddAt(time, name, value, tags...)
}

// Set sets to value the gauge identified by name and tags.
func Set(name string, value interface{}, tags ...Tag) {
	DefaultEngine.Set(name, value, tags...)
}

// SetAt sets to value the gauge identified by name and tags.
func SetAt(time time.Time, name string, value interface{}, tags ...Tag) {
	DefaultEngine.SetAt(time, name, value, tags...)
}

// Observe reports value for the histogram identified by name and tags.
func Observe(name string, value interface{}, tags ...Tag) {
	DefaultEngine.Observe(name, value, tags...)
}

// ObserveAt reports value for the histogram identified by name and tags.
func ObserveAt(time time.Time, name string, value interface{}, tags ...Tag) {
	DefaultEngine.ObserveAt(time, name, value, tags...)
}

// Report is a helper function that delegates to DefaultEngine.
func Report(metrics interface{}, tags ...Tag) {
	DefaultEngine.Report(metrics, tags...)
}

// ReportAt is a helper function that delegates to DefaultEngine.
func ReportAt(time time.Time, metrics interface{}, tags ...Tag) {
	DefaultEngine.ReportAt(time, metrics, tags...)
}

func progname() (name string) {
	if args := os.Args; len(args) != 0 {
		name = filepath.Base(args[0])
	}
	return
}
