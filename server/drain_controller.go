package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
)

// DrainController handles all requests relating to Atlantis drainage (to shutdown properly).
type DrainController struct {
	Logger  *logging.SimpleLogger
	Drainer *events.Drainer
}

type DrainResponse struct {
	DrainStarted             bool `json:"started"`
	DrainCompleted           bool `json:"completed"`
	OngoingOperationsCounter int  `json:"ongoingOperations"`
}

// Get is the GET /drain route. It renders the current drainage status.
func (d *DrainController) Get(w http.ResponseWriter, r *http.Request) {
	d.respondStatus(http.StatusOK, w)
}

// Post is the POST /drain route. It asks atlantis to finish all ongoing operations and to refuse to start new ones.
func (d *DrainController) Post(w http.ResponseWriter, r *http.Request) {
	d.Drainer.StartDrain()
	d.respondStatus(http.StatusCreated, w)
}

func (d *DrainController) respondStatus(responseCode int, w http.ResponseWriter) {
	data, err := json.MarshalIndent(&DrainResponse{
		DrainStarted:             d.Drainer.DrainStarted,
		DrainCompleted:           d.Drainer.DrainCompleted,
		OngoingOperationsCounter: d.Drainer.OngoingOperationsCounter,
	}, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating status json response: %s", err)
		return
	}
	d.Logger.Log(logging.Info, "Drain status: %s", string(data))
	w.WriteHeader(responseCode)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data) // nolint: errcheck
}
