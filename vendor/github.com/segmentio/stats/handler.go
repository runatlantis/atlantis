package stats

import "time"

// The Handler interface is implemented by types that produce measures to
// various metric collection backends.
type Handler interface {
	// HandleMeasures is called by the Engine on which the handler was set
	// whenever new measures are produced by the program. The first argument
	// is the time at which the measures were taken.
	//
	// The method must treat the list of measures as read-only values, and
	// must not retain pointers to any of the measures or their sub-fields
	// after returning.
	HandleMeasures(time time.Time, measures ...Measure)
}

// Flusher is an interface implemented by measure handlers in order to flush
// any buffered data.
type Flusher interface {
	Flush()
}

func flush(h Handler) {
	if f, ok := h.(Flusher); ok {
		f.Flush()
	}
}

// HandleFunc is a type alias making it possible to use simple functions as
// measure handlers.
type HandlerFunc func(time.Time, ...Measure)

// HandleMeasures calls f, satisfies the Handler interface.
func (f HandlerFunc) HandleMeasures(time time.Time, measures ...Measure) {
	f(time, measures...)
}

// MultiHandler constructs a handler which dispatches measures to all given
// handlers.
func MultiHandler(handlers ...Handler) Handler {
	multi := make([]Handler, 0, len(handlers))

	for _, h := range handlers {
		if h != nil {
			if m, ok := h.(*multiHandler); ok {
				multi = append(multi, m.handlers...) // flatten multi handlers
			} else {
				multi = append(multi, h)
			}
		}
	}

	if len(multi) == 1 {
		return multi[0]
	}

	return &multiHandler{handlers: multi}
}

type multiHandler struct {
	handlers []Handler
}

func (m *multiHandler) HandleMeasures(time time.Time, measures ...Measure) {
	for _, h := range m.handlers {
		h.HandleMeasures(time, measures...)
	}
}

func (m *multiHandler) Flush() {
	for _, h := range m.handlers {
		flush(h)
	}
}

// Discard is a handler that doesn't do anything with the measures it receives.
var Discard = &discard{}

type discard struct{}

func (*discard) HandleMeasures(time time.Time, measures ...Measure) {}
