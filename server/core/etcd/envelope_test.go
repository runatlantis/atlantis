// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package etcd

import (
	"testing"

	. "github.com/runatlantis/atlantis/testing"
)

type sample struct {
	A string `json:"a"`
	B int    `json:"b"`
}

func TestEnvelope_RoundTrip(t *testing.T) {
	in := sample{A: "x", B: 3}
	data, err := encodeValue("sample", 1, in)
	Ok(t, err)
	var out sample
	Ok(t, decodeValue(data, "sample", 1, 1, &out))
	Equals(t, in, out)
}

func TestEnvelope_KindMismatch(t *testing.T) {
	data, err := encodeValue("sample", 1, sample{A: "x"})
	Ok(t, err)
	var out sample
	ErrContains(t, "expected record kind", decodeValue(data, "other", 1, 1, &out))
}

func TestEnvelope_VersionOutOfRange(t *testing.T) {
	data, err := encodeValue("sample", 2, sample{A: "x"})
	Ok(t, err)
	var out sample
	ErrContains(t, "outside readable range", decodeValue(data, "sample", 1, 1, &out))
}

// TestEnvelope_UnknownFieldStrict proves decoding is strict: an unexpected
// envelope field is rejected rather than silently ignored.
func TestEnvelope_UnknownFieldStrict(t *testing.T) {
	var out sample
	ErrContains(t, "decoding sample envelope", decodeValue([]byte(`{"kind":"sample","version":1,"payload":{},"extra":true}`), "sample", 1, 1, &out))
}
