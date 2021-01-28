package slack

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// MaxMessageTextLength is the current maximum message length in number of characters as defined here
	// https://api.slack.com/rtm#limits
	MaxMessageTextLength = 4000
)

// RTM represents a managed websocket connection. It also supports
// all the methods of the `Client` type.
//
// Create this element with Client's NewRTM() or NewRTMWithOptions(*RTMOptions)
type RTM struct {
	idGen        IDGenerator
	pingInterval time.Duration
	pingDeadman  *time.Timer

	// Connection life-cycle
	conn             *websocket.Conn
	IncomingEvents   chan RTMEvent
	outgoingMessages chan OutgoingMessage
	killChannel      chan bool
	disconnected     chan struct{} // disconnected is closed when Disconnect is invoked, regardless of connection state. Allows for ManagedConnection to not leak.
	forcePing        chan bool
	rawEvents        chan json.RawMessage
	wasIntentional   bool
	isConnected      bool

	// Client is the main API, embedded
	Client
	websocketURL string

	// UserDetails upon connection
	info *Info

	// useRTMStart should be set to true if you want to use
	// rtm.start to connect to Slack, otherwise it will use
	// rtm.connect
	useRTMStart bool

	// dialer is a gorilla/websocket Dialer. If nil, use the default
	// Dialer.
	dialer *websocket.Dialer

	// mu is mutex used to prevent RTM connection race conditions
	mu *sync.Mutex
}

// RTMOptions allows configuration of various options available for RTM messaging
//
// This structure will evolve in time so please make sure you are always using the
// named keys for every entry available as per Go 1 compatibility promise adding fields
// to this structure should not be considered a breaking change.
type RTMOptions struct {
	// UseRTMStart set to true in order to use rtm.start or false to use rtm.connect
	// As of 11th July 2017 you should prefer setting this to false, see:
	// https://api.slack.com/changelog/2017-04-start-using-rtm-connect-and-stop-using-rtm-start
	UseRTMStart bool
}

// Disconnect and wait, blocking until a successful disconnection.
func (rtm *RTM) Disconnect() error {
	// avoid RTM disconnect race conditions
	rtm.mu.Lock()
	defer rtm.mu.Unlock()

	// always push into the disconnected channel when invoked,
	// this lets the ManagedConnection() function properly clean up.
	// if the buffer is full then just continue on.
	select {
	case rtm.disconnected <- struct{}{}:
	default:
	}

	if !rtm.isConnected {
		return errors.New("Invalid call to Disconnect - Slack API is already disconnected")
	}

	rtm.killChannel <- true
	return nil
}

// GetInfo returns the info structure received when calling
// "startrtm", holding all channels, groups and other metadata needed
// to implement a full chat client. It will be non-nil after a call to
// StartRTM().
func (rtm *RTM) GetInfo() *Info {
	return rtm.info
}

// SendMessage submits a simple message through the websocket.  For
// more complicated messages, use `rtm.PostMessage` with a complete
// struct describing your attachments and all.
func (rtm *RTM) SendMessage(msg *OutgoingMessage) {
	if msg == nil {
		rtm.Debugln("Error: Attempted to SendMessage(nil)")
		return
	}

	rtm.outgoingMessages <- *msg
}

func (rtm *RTM) resetDeadman() {
	timerReset(rtm.pingDeadman, deadmanDuration(rtm.pingInterval))
}

func deadmanDuration(d time.Duration) time.Duration {
	return d * 4
}
