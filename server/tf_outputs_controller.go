package server

import (
	"bufio"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	"log"
	"net/http"
)

type TfOutputsController struct {
	log            *logging.SimpleLogger
	tfOutputHelper terraform.OutputHelper
	wcUpgrader     websocket.Upgrader
}

func (o *TfOutputsController) GetTfOutputParams() []string {
	return []string{"createdAt", "fullRepoName", "pullNr", "project", "headCommit", "workspace", "tfCommand"}
}

func (o *TfOutputsController) GetTfOutput(w http.ResponseWriter, r *http.Request) {
	o.log.Debug("starting a websocket for getting a tf output file, request %s", r.URL.String())
	// Get all the parameters from the query string
	paramValues := make(map[string]string, len(o.GetTfOutputParams()))
	mux.Vars()
	for _, param := range o.GetTfOutputParams() {
		ok := false
		paramValues[param], ok = mux.Vars(r)[param]
		if !ok {
			o.log.Warn("no %s parameter found", param)
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "No %s parameter found \n", param)
			return
		}
	}

	tfOutputFileName, err := o.tfOutputHelper.FindOutputFile(
		paramValues["createdAt"],
		paramValues["fullRepoName"],
		paramValues["pullNr"],
		paramValues["project"],
		paramValues["headCommit"],
		paramValues["workspace"],
		paramValues["tfCommand"],
	)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "No tf output file found")
		return
	}

	// writer
	done := make(chan bool)
	var buff bufio.ReadWriter
	go func() {
		o.log.Debug("reading file %q", tfOutputFileName)
		err := o.tfOutputHelper.ContinueReadFile(o.log, tfOutputFileName, &buff, done)
		if err != nil {
			o.log.Err("Fail to tail tf output file, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Fail to tail tf output file")
			return
		}
	}()

	// Creates the websocket
	c, err := o.wcUpgrader.Upgrade(w, r, nil)
	if err != nil {
		o.log.Err("fail to start websocket server, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Fail to start websocket server")
		return
	}
	defer c.Close()

	c.SetCloseHandler(func(code int, text string) error {
		// Stopping reading the file (tail -f)
		done <- true
		return nil
	})

	// Reads the messages written by the tf output file
	scanner := bufio.NewScanner(&buff)
	// For every single message writes a message in the websocket
	for scanner.Scan() {
		err = c.WriteMessage(websocket.TextMessage, scanner.Bytes())
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}
