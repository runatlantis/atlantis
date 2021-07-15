package controllers_test

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
	"strconv"
	"testing"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/controllers/templates"
	"github.com/runatlantis/atlantis/server/events/db"
	"github.com/runatlantis/atlantis/server/events/models"
	"github.com/runatlantis/atlantis/server/logging"

	tMocks "github.com/runatlantis/atlantis/server/controllers/templates/mocks"

	"github.com/gorilla/mux"
	. "github.com/petergtz/pegomock"
	"github.com/runatlantis/atlantis/server/events"

	mocks2 "github.com/runatlantis/atlantis/server/events/mocks"
	vcsmocks "github.com/runatlantis/atlantis/server/events/vcs/mocks"
	. "github.com/runatlantis/atlantis/testing"
)

func TestAddChan(t *testing.T) {
	
}

func TestRemoveChan(t *testing.T) {

}

func TestWritingLogLine(t *testing.T) {

}

func TestClearLogLine(t *testing.T) {

}

func TestListen(t *testing.T) {

}

func TestString(t *testing.T) {

}

func TestNewPullInfo(t *testing.T) {

}

func TestGetLogStream(t *testing.T) {

}

func TestGetLogStream_WebSockets(t *testing.T) {

}

func TestRespond(t *testing.T) {

}

func TestRetrievePrStatus(t *testing.T) {

}
