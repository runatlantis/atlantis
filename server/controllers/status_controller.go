package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
)

// StatusController handles the status of Atlantis.
type StatusController struct {
	Logger  logging.SimpleLogging
	Drainer *events.Drainer
}

type StatusResponse struct {
	ShuttingDown  bool `json:"shutting_down"`
	InProgressOps int  `json:"in_progress_operations"`
}

// Get is the GET /status route.
func (d *StatusController) Get(w http.ResponseWriter, r *http.Request) {
	status := d.Drainer.GetStatus()
	data, err := json.MarshalIndent(&StatusResponse{
		ShuttingDown:  status.ShuttingDown,
		InProgressOps: status.InProgressOps,
	}, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error creating status json response: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data) // nolint: errcheck
}
