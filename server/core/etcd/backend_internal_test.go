// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// TestBackend_FatalEmbeddedErrorFailsReady proves that a fatal embedded-server
// error recorded after startup makes the backend permanently unready — the
// readiness probe short-circuits on it before touching the client (design §783).
func TestBackend_FatalEmbeddedErrorFailsReady(t *testing.T) {
	b := &backend{requestTimeout: time.Second, probeKey: "/x"}
	b.setFatal(errors.New("embedded boom"))
	err := b.Ready(context.Background())
	if err == nil {
		t.Fatal("expected Ready to fail after a fatal embedded error")
	}
	if !strings.Contains(err.Error(), "embedded etcd server failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestBackend_SetFatalKeepsFirst proves only the first fatal error is retained.
func TestBackend_SetFatalKeepsFirst(t *testing.T) {
	b := &backend{}
	b.setFatal(errors.New("first"))
	b.setFatal(errors.New("second"))
	if b.fatal().Error() != "first" {
		t.Fatalf("expected the first fatal error to stand, got %v", b.fatal())
	}
}
