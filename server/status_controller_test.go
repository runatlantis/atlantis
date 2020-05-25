package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatusController_Startup(t *testing.T) {
	logger := logging.NewNoopLogger()
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	d := &server.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result server.StatusResponse
	body, err := ioutil.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, false, result.ShuttingDown)
	Equals(t, 0, result.InProgressOps)
}

func TestStatusController_InProgress(t *testing.T) {
	logger := logging.NewNoopLogger()
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	dr.StartOp()

	d := &server.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result server.StatusResponse
	body, err := ioutil.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, false, result.ShuttingDown)
	Equals(t, 1, result.InProgressOps)
}

func TestStatusController_Shutdown(t *testing.T) {
	logger := logging.NewNoopLogger()
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	dr.ShutdownBlocking()

	d := &server.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result server.StatusResponse
	body, err := ioutil.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, true, result.ShuttingDown)
	Equals(t, 0, result.InProgressOps)
}
