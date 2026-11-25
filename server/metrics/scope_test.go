// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics_test

import (
	"errors"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	"github.com/runatlantis/atlantis/server/metrics"
)

var (
	prometheusConfig = valid.Metrics{
		Prometheus: &valid.Prometheus{
			Endpoint: "/metrics",
		},
	}
)

func TestNewScope_PrometheusTaggingCapabilities(t *testing.T) {
	scope, _, _, err := metrics.NewScope(prometheusConfig, nil, "test")
	if err != nil {
		t.Fatalf("got an error: %s", err.Error())
	}

	scope.Tagged(map[string]string{
		"base_repo": "runatlantis/atlantis",
		"pr_number": "2687",
	})

	want := true
	got := scope.Capabilities().Tagging()
	if want != got {
		t.Errorf("Scope does not have Capability to do Tagging")
	}
}

func TestNewScope_StatsdReporterEmitsCounter(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening for statsd packet: %s", err)
	}
	defer conn.Close() // nolint: errcheck

	udpAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok {
		t.Fatalf("expected UDP listener address, got %T", conn.LocalAddr())
	}

	scope, reporter, closer, err := metrics.NewScope(valid.Metrics{
		Statsd: &valid.Statsd{
			Host: udpAddr.IP.String(),
			Port: strconv.Itoa(udpAddr.Port),
		},
	}, nil, "test")
	if err != nil {
		t.Fatalf("got an error: %s", err.Error())
	}
	defer closer.Close() // nolint: errcheck
	if reporter == nil {
		t.Fatal("expected statsd reporter")
	}

	scope.Counter("events").Inc(1)

	buf := make([]byte, 1024)
	deadline := time.Now().Add(3 * time.Second)
	var packets []string
	for {
		if err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond)); err != nil {
			t.Fatalf("setting read deadline: %s", err)
		}
		n, _, err := conn.ReadFrom(buf)
		if err == nil {
			got := string(buf[:n])
			packets = append(packets, got)
			if strings.Contains(got, "test_events:1|c") {
				return
			}
			continue
		}
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() && time.Now().Before(deadline) {
			continue
		}
		t.Fatalf("reading statsd packet: %s; packets: %q", err, packets)
	}
}
