package server

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/runatlantis/atlantis/server/events/terraform"
	"github.com/runatlantis/atlantis/server/logging"
	"net/http"
	"strings"
)

type TfOutputsController struct {
	log            *logging.SimpleLogger
	tfOutputHelper terraform.OutputHelper
	wcUpgrader     websocket.Upgrader
}

func (o *TfOutputsController) GetTfOutput(w http.ResponseWriter, r *http.Request) {
	o.wcUpgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	// Creates the websocket by upgrading the request
	c, err := o.wcUpgrader.Upgrade(w, r, nil)
	if err != nil {
		o.log.Err("fail to start websocket server, %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "Fail to start websocket server")
		return
	}
	defer c.Close()

	// Channel to close the method tailing the tf output file
	done := make(chan bool)
	c.SetCloseHandler(func(code int, text string) error {
		// Stopping reading the file (tail -f)
		done <- true
		return nil
	})

	for {
		// Websocket start with the client sending a message
		mt, msg, err := c.ReadMessage()
		if err != nil {
			o.log.Err("fail to read message from the websocket server, %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintln(w, "Fail to read message from the websocket server")
			return
		}
		o.log.Debug("Received %q", string(msg))

		// The message contains the tf output file infos to stream the log in messages
		tfOutputInfos := strings.Split(string(msg), "|")

		// Gets the tf output file name to stream
		tfOutputFileName, err := o.tfOutputHelper.FindOutputFile(
			tfOutputInfos[0],
			tfOutputInfos[1],
			tfOutputInfos[2],
			tfOutputInfos[3],
			tfOutputInfos[4],
			tfOutputInfos[5],
			tfOutputInfos[6],
		)
		if err != nil {
			o.log.Err("can't find file %v", tfOutputInfos)
			err = c.WriteMessage(mt, []byte("file not found"))
			if err != nil {
				o.log.Err("fail to write in the websocket server, %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "Fail to write in the websocket server")
			}
			return
		}

		// Call a go function to start the continue read file method, and as a return the channel  will receive the messages
		// to reply back in the websocket
		fileText := make(chan string)
		go func() {
			err := o.tfOutputHelper.ContinueReadFile(o.log, tfOutputFileName, fileText, done)
			if err != nil {
				err = c.WriteMessage(mt, []byte("failed to tail the file"))
				if err != nil {
					o.log.Err("fail to write in the websocket server, %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, "Fail to write in the websocket server")
				}
				return
			}
		}()

		for {
			select {
			// For a new message in the channel, post it in the websocket
			case text := <-fileText:
				o.log.Debug("receiving new message")
				err = c.WriteMessage(websocket.TextMessage, []byte(text))
				if err != nil {
					o.log.Err("fail to write in the websocket, %v", err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintln(w, "Fail to write in the websocket")
					return
				}
			}
		}
	}

}
