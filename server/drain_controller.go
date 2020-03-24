package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/runatlantis/atlantis/server/logging"
)

// DrainController handles all requests relating to Atlantis drainage (to shutdown properly).
type DrainController struct {
	Logger                   *logging.SimpleLogger
	DrainStarted             bool
	DrainCompleted           bool
	OngoingOperationsCounter int
	mutex                    sync.Mutex
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
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.DrainStarted = true
	if d.OngoingOperationsCounter == 0 {
		d.DrainCompleted = true
	}
	d.respondStatus(http.StatusCreated, w)
}

// Try to add an operation as ongoing. Return true if the operation is allowed to start, false if it should be rejected.
func (d *DrainController) TryAddNewOngoingOperation() bool {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.DrainStarted {
		return false
	} else {
		d.OngoingOperationsCounter += 1
		return true
	}
}

// Consider on operation as completed.
func (d *DrainController) RemoveOngoingOperation() {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.OngoingOperationsCounter -= 1
	if d.OngoingOperationsCounter < 0 {
		d.Logger.Log(logging.Warn, "Drain OngoingOperationsCounter became below 0, this is a bug")
		d.OngoingOperationsCounter = 0
	}
	if d.DrainStarted && d.OngoingOperationsCounter == 0 {
		d.DrainCompleted = true
	}
}

func (d *DrainController) respondStatus(responseCode int, w http.ResponseWriter) {
	data, err := json.MarshalIndent(&DrainResponse{
		DrainStarted:             d.DrainStarted,
		DrainCompleted:           d.DrainCompleted,
		OngoingOperationsCounter: d.OngoingOperationsCounter,
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
