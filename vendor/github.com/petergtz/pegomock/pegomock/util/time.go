package util

import "time"

// Ticker repeatedly calls cb with a delay in between calls. It stops doing This
// When a element is sent to the done channel.
func Ticker(cb func(), delay time.Duration, done chan bool) {
	for {
		select {
		case <-done:
			return

		default:
			cb()
			time.Sleep(delay)
		}
	}
}
