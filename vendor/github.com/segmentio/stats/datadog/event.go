package datadog

import (
	"fmt"

	"github.com/segmentio/stats"
)

// EventPriority is an enumeration providing the available datadog event
// priority levels.
type EventPriority string

const (
	EventPriorityNormal EventPriority = "normal"
	EventPriorityLow    EventPriority = "low"
)

// EventAlertType is an enumeration providing the available datadog event
// allert types.
type EventAlertType string

const (
	EventAlertTypeError   EventAlertType = "error"
	EventAlertTypeWarning EventAlertType = "warning"
	EventAlertTypeInfo    EventAlertType = "info"
	EventAlertTypeSuccess EventAlertType = "success"
)

// Event is a representation of a datadog event
type Event struct {
	Title          string
	Text           string
	Ts             int64
	Priority       EventPriority
	Host           string
	Tags           []stats.Tag
	AlertType      EventAlertType
	AggregationKey string
	SourceTypeName string
	EventType      string
}

// String satisfies the fmt.Stringer interface.
func (e Event) String() string {
	return fmt.Sprint(e)
}

// Format satisfies the fmt.Formatter interface.
func (e Event) Format(f fmt.State, _ rune) {
	buf := bufferPool.Get().(*buffer)
	buf.b = appendEvent(buf.b[:0], e)
	f.Write(buf.b)
	bufferPool.Put(buf)
}
