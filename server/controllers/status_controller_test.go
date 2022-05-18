package controllers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/runatlantis/atlantis/server/controllers"
	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

func TestStatusController_Startup(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	d := &controllers.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result controllers.StatusResponse
	body, err := io.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, false, result.ShuttingDown)
	Equals(t, 0, result.InProgressOps)
}

func TestStatusController_InProgress(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	dr.StartOp()

	d := &controllers.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result controllers.StatusResponse
	body, err := io.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, false, result.ShuttingDown)
	Equals(t, 1, result.InProgressOps)
}

func TestStatusController_Shutdown(t *testing.T) {
	logger := logging.NewNoopLogger(t)
	r, _ := http.NewRequest("GET", "/status", bytes.NewBuffer(nil))
	w := httptest.NewRecorder()
	dr := &events.Drainer{}
	dr.ShutdownBlocking()

	d := &controllers.StatusController{
		Logger:  logger,
		Drainer: dr,
	}
	d.Get(w, r)

	var result controllers.StatusResponse
	body, err := io.ReadAll(w.Result().Body)
	Ok(t, err)
	Equals(t, 200, w.Result().StatusCode)
	err = json.Unmarshal(body, &result)
	Ok(t, err)
	Equals(t, true, result.ShuttingDown)
	Equals(t, 0, result.InProgressOps)
}
