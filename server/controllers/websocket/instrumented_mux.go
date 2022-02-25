package websocket

import (
	"net/http"

	"github.com/uber-go/tally"
)

type InstrumentedMultiplexor struct {
	Multiplexor

	NumWsConnections tally.Counter
}

func NewInstrumentedMultiplexor(multiplexor Multiplexor, statsScope tally.Scope) Multiplexor {
	return &InstrumentedMultiplexor{
		Multiplexor:      multiplexor,
		NumWsConnections: statsScope.SubScope("websocket").Counter("connections"),
	}
}

func (i *InstrumentedMultiplexor) Handle(w http.ResponseWriter, r *http.Request) error {
	i.NumWsConnections.Inc(1)
	return i.Multiplexor.Handle(w, r)
}
