// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.
//
// This file implements the versioned JSON value envelope (design §619). Values
// use versioned JSON envelopes. The envelope frame — kind and version — is
// decoded strictly and fails closed: an unrecognized kind or a version outside
// the reader's declared range is a hard error, never a best-effort parse, so a
// binary must explicitly implement every record version it claims to read.
//
// The inner payload is decoded leniently (unknown fields are ignored) on
// purpose: some records wrap Atlantis core models (ProjectLock, PullStatus) that
// gain fields over time, and an additive model change within the same record
// version must not turn every stored record unreadable on an older binary. The
// kind+version gate is the strict boundary; forward-incompatible payload changes
// are expressed by bumping the record version, not by field-level strictness.
package etcd

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// envelope wraps every stored value with its record kind and version. The kind
// guards against a key being decoded as the wrong record type; the version
// selects the decoder.
type envelope struct {
	Kind    string          `json:"kind"`
	Version int             `json:"version"`
	Payload json.RawMessage `json:"payload"`
}

// encodeValue serializes payload into a versioned envelope. kind and version
// identify the record so a reader can refuse an unexpected type or version.
func encodeValue(kind string, version int, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling %s payload: %w", kind, err)
	}
	return json.Marshal(envelope{Kind: kind, Version: version, Payload: raw})
}

// decodeValue decodes a versioned envelope. The frame is strict — it fails
// closed on a kind mismatch or a version outside [minVersion, maxVersion], and
// absence of a known version is never treated as an empty record. The inner
// payload is decoded leniently (see the file header): additive fields within a
// recognized version are tolerated so wrapped Atlantis models can evolve.
func decodeValue(data []byte, kind string, minVersion, maxVersion int, payload any) error {
	var env envelope
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&env); err != nil {
		return fmt.Errorf("decoding %s envelope: %w", kind, err)
	}
	if env.Kind != kind {
		return fmt.Errorf("expected record kind %q but stored value is %q", kind, env.Kind)
	}
	if env.Version < minVersion || env.Version > maxVersion {
		return fmt.Errorf("%s record version %d is outside readable range [%d,%d]; this binary does not implement it", kind, env.Version, minVersion, maxVersion)
	}
	if err := json.Unmarshal(env.Payload, payload); err != nil {
		return fmt.Errorf("decoding %s payload v%d: %w", kind, env.Version, err)
	}
	return nil
}
