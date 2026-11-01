// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package websocket

import "time"

// SetPongWait shortens the keepalive timing so tests do not wait a minute.
func (w *Writer) SetPongWait(d time.Duration) { w.pongWait = d }

// CheckOriginFunc exposes checkOriginFunc to the external Multiplexor tests.
var CheckOriginFunc = checkOriginFunc
