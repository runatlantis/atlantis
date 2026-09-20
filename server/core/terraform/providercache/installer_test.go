// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0

package providercache

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"

	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
)

// testCoordinate is a fixed provider coordinate used across these tests.
var testCoordinate = Coordinate{
	Host: registryHost, Namespace: "hashicorp", Type: "null", Version: "3.2.1", OS: "linux", Arch: "amd64",
}

// newTestInstaller builds an installer whose fetch simply serves files out of
// a local map, keyed by URL, bypassing the HTTP layer entirely so these tests
// can focus on verification and publication logic.
func newTestInstaller(t *testing.T, files map[string][]byte) *installer {
	t.Helper()
	cacheDir := t.TempDir()
	return &installer{
		log:       logging.NewNoopLogger(t),
		mirrorDir: filepath.Join(t.TempDir(), "mirror"),
		fetch: func(_ context.Context, rawURL string) (string, error) {
			data, ok := files[rawURL]
			if !ok {
				t.Fatalf("unexpected fetch of %s", rawURL)
			}
			path := filepath.Join(cacheDir, filepath.Base(rawURL))
			Ok(t, os.WriteFile(path, data, 0600))
			return path, nil
		},
	}
}

// validMeta returns a "find a provider package" response, and the raw file
// contents it references, for a correctly signed archive.
func validMeta(t *testing.T) (meta map[string]any, files map[string][]byte) {
	t.Helper()
	archive := validArchive(t)
	shasums, signature := signedShasums(t, archive)
	sum := sha256.Sum256(archive)
	_, publicKeyArmor := testSigningKey()

	meta = map[string]any{
		"filename":              archiveFilename,
		"download_url":          "https://upstream.example/archive.zip",
		"shasums_url":           "https://upstream.example/SHA256SUMS",
		"shasums_signature_url": "https://upstream.example/SHA256SUMS.sig",
		"shasum":                hex.EncodeToString(sum[:]),
		"signing_keys": map[string]any{
			"gpg_public_keys": []any{
				map[string]any{"ascii_armor": publicKeyArmor},
			},
		},
	}
	files = map[string][]byte{
		"https://upstream.example/archive.zip":    archive,
		"https://upstream.example/SHA256SUMS":     shasums,
		"https://upstream.example/SHA256SUMS.sig": signature,
	}
	return meta, files
}

func TestInstaller_PublishesVerifiedProvider(t *testing.T) {
	meta, files := validMeta(t)
	in := newTestInstaller(t, files)

	Ok(t, in.install(context.Background(), testCoordinate, meta))
	Assert(t, in.isPublished(testCoordinate), "expected coordinate to be published")

	installed, err := os.ReadFile(filepath.Join(testCoordinate.mirrorPath(in.mirrorDir), providerBinaryName))
	Ok(t, err)
	Equals(t, "#!/bin/sh\necho fake provider\n", string(installed))
}

func TestInstaller_RejectsChecksumMismatch(t *testing.T) {
	meta, files := validMeta(t)
	// Tamper with the archive so it no longer matches the signed SHA256SUMS
	// entry, without touching that entry or its signature.
	files["https://upstream.example/archive.zip"] = append(bytes.Clone(files["https://upstream.example/archive.zip"]), 0x00)

	in := newTestInstaller(t, files)
	err := in.install(context.Background(), testCoordinate, meta)
	Assert(t, err != nil, "expected a checksum mismatch error")
	Assert(t, !in.isPublished(testCoordinate), "a checksum-mismatched archive must never be published")
}

func TestInstaller_RejectsForgedSignature(t *testing.T) {
	meta, files := validMeta(t)

	// Sign the (otherwise identical) SHA256SUMS with a different keypair than
	// the one advertised in signing_keys, simulating a forged signature.
	otherKey, err := openpgp.NewEntity("Attacker", "", "attacker@example.com", nil)
	Ok(t, err)
	var forged bytes.Buffer
	Ok(t, openpgp.DetachSign(&forged, otherKey, bytes.NewReader(files["https://upstream.example/SHA256SUMS"]), nil))
	files["https://upstream.example/SHA256SUMS.sig"] = forged.Bytes()

	in := newTestInstaller(t, files)
	err = in.install(context.Background(), testCoordinate, meta)
	Assert(t, err != nil, "expected a signature verification error")
	Assert(t, !in.isPublished(testCoordinate), "an unverifiably-signed archive must never be published")
}

func TestInstaller_RejectsMissingSigningKeys(t *testing.T) {
	meta, files := validMeta(t)
	meta["signing_keys"] = map[string]any{"gpg_public_keys": []any{}}

	in := newTestInstaller(t, files)
	err := in.install(context.Background(), testCoordinate, meta)
	Assert(t, err != nil, "expected an error when no signing keys are offered")
	Assert(t, !in.isPublished(testCoordinate), "an unverifiable archive must never be published")
}

func TestInstaller_EnsureInstalledDedupesConcurrentCalls(t *testing.T) {
	meta, files := validMeta(t)
	in := newTestInstaller(t, files)

	const n = 15
	done := make(chan struct{}, n)
	for range n {
		go func() {
			in.EnsureInstalled(testCoordinate, meta)
			done <- struct{}{}
		}()
	}
	for range n {
		<-done
	}

	waitForInstaller(t, in, testCoordinate)
	Equals(t, int64(1), in.installs.Load())

	// EnsureInstalled's outer isPublished check and singleflight's own
	// "forget on completion" both make a second call for this coordinate a
	// no-op once published - but only once that second call actually reaches
	// its own isPublished re-check. Under heavy scheduler contention, a
	// straggler among the 15 goroutines above can have read isPublished as
	// false (correctly, at that moment) before being descheduled, and only
	// resume - past the point where it would now see the coordinate as
	// published - after this test would otherwise have already returned and
	// torn down its t.TempDir()-backed files out from under it. Give any such
	// straggler a generous window to reach its own safe early-return first.
	time.Sleep(200 * time.Millisecond)
}

// Test that a failed install is recorded per coordinate (grouped under its
// host by LastErrors), and that a later successful install of that same
// coordinate clears it - so a caller that gave up waiting only ever sees a
// currently-true failure, not a stale one from an attempt that has since
// succeeded.
func TestInstaller_LastErrorsRecordsAndClearsOnSuccess(t *testing.T) {
	// recordResult is called by EnsureInstalled around install, not by
	// install itself, so drive this through EnsureInstalled (and poll for
	// completion, same as the concurrency/timeout tests) rather than calling
	// install directly.
	meta, files := validMeta(t)
	meta["signing_keys"] = map[string]any{"gpg_public_keys": []any{}} // fail closed
	in := newTestInstaller(t, files)

	in.EnsureInstalled(testCoordinate, meta)
	waitForInstalls(t, in, 1)
	errs := in.LastErrors(testCoordinate.Host)
	Assert(t, len(errs) == 1, "expected exactly one recorded error, got %v", errs)
	Assert(t, strings.Contains(errs[0], "refusing to install an unverifiable package"), "unexpected recorded error: %s", errs[0])
	Equals(t, []string(nil), in.LastErrors("some.other.host"))

	// A later successful install of the same coordinate clears the record.
	okMeta, _ := validMeta(t)
	in.EnsureInstalled(testCoordinate, okMeta)
	waitForInstaller(t, in, testCoordinate)
	Equals(t, []string(nil), in.LastErrors(testCoordinate.Host))
}

// waitForInstalls blocks until in.installs has reached at least n completed
// attempts.
func waitForInstalls(t *testing.T, in *installer, n int64) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for in.installs.Load() < n {
		Assert(t, time.Now().Before(deadline), "installs never reached %d (at %d)", n, in.installs.Load())
		time.Sleep(5 * time.Millisecond)
	}
}

// Test that EnsureInstalled bounds a single install attempt with
// installTimeout: a fetch that never returns on its own (simulating a
// stalled upstream download, which downloadClient's own Timeout: 0 will not
// catch - see the package doc comment) doesn't wedge the coordinate forever,
// and it's left unpublished so a later attempt can retry it.
func TestInstaller_EnsureInstalledRespectsInstallTimeout(t *testing.T) {
	meta, _ := validMeta(t)
	in := &installer{
		log:            logging.NewNoopLogger(t),
		mirrorDir:      filepath.Join(t.TempDir(), "mirror"),
		installTimeout: 20 * time.Millisecond,
		fetch: func(ctx context.Context, _ string) (string, error) {
			<-ctx.Done() // never resolves on its own, like a stalled download.
			return "", ctx.Err()
		},
	}

	start := time.Now()
	in.EnsureInstalled(testCoordinate, meta)
	// EnsureInstalled itself must return immediately regardless of
	// installTimeout - it never blocks the HTTP handler that calls it.
	Assert(t, time.Since(start) < 100*time.Millisecond, "EnsureInstalled blocked its caller")

	deadline := time.Now().Add(2 * time.Second)
	for in.installs.Load() == 0 {
		Assert(t, time.Now().Before(deadline), "the timed-out install never completed")
		time.Sleep(5 * time.Millisecond)
	}
	Assert(t, !in.isPublished(testCoordinate), "a timed-out install must never be published")
	errs := in.LastErrors(testCoordinate.Host)
	Assert(t, len(errs) == 1, "expected the timeout to be recorded as this coordinate's last error, got %v", errs)
}

func waitForInstaller(t *testing.T, in *installer, c Coordinate) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		if in.isPublished(c) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("coordinate %s was never published", c.key())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
