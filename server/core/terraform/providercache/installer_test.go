// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0

package providercache

import (
	"archive/zip"
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
	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/packet"

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

// Test that unzip refuses every entry whose path would resolve outside the
// destination directory ("zip slip"), across a range of payload shapes, and
// that destDir is left empty in each case - nothing gets written before the
// unsafe entry is rejected.
func TestUnzip_RejectsPathTraversal(t *testing.T) {
	malicious := []string{
		"../outside.txt",
		"../../../../etc/passwd",
		"a/../../outside.txt",
		"..",
		"a/b/../../../outside.txt",
	}
	for _, name := range malicious {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			zw := zip.NewWriter(&buf)
			w, err := zw.Create(name)
			Ok(t, err)
			_, err = w.Write([]byte("pwned"))
			Ok(t, err)
			Ok(t, zw.Close())

			destDir := t.TempDir()
			err = unzip(buf.Bytes(), destDir)
			Assert(t, err != nil, "expected entry %q to be rejected", name)

			entries, err := os.ReadDir(destDir)
			Ok(t, err)
			Equals(t, 0, len(entries))
		})
	}
}

// Test that a symlink entry is refused outright, rather than having its
// target path written as if it were the symlink's file content (which could
// otherwise be used to stage a later entry writing through it).
func TestUnzip_RejectsSymlinkEntry(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	hdr := &zip.FileHeader{Name: "link"}
	hdr.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(hdr)
	Ok(t, err)
	_, err = w.Write([]byte("/etc"))
	Ok(t, err)
	Ok(t, zw.Close())

	destDir := t.TempDir()
	err = unzip(buf.Bytes(), destDir)
	Assert(t, err != nil, "expected the symlink entry to be rejected")
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

// Test that a signature made validly by a key that has since expired (by
// wall-clock time, at verification) is still accepted - this is a
// regression test for a real bug found via local end-to-end testing against
// the actual registry.terraform.io and HashiCorp's real (long-lived, and at
// the time of that test, expired-by-this-sandbox's-clock) signing key:
// go-crypto's CheckDetachedSignature checks key expiry against wall-clock
// "now" by default, which would permanently break every future install the
// moment any registry rotates its signing key, even for archives it validly
// signed while the key was still current. install must instead verify as of
// the signature's own creation time.
func TestInstaller_AcceptsSignatureFromSinceExpiredKey(t *testing.T) {
	// A key that's only valid for 1 second from creation.
	signer, err := openpgp.NewEntity("Short-Lived Registry Key", "", "test@example.com", &packet.Config{KeyLifetimeSecs: 1})
	Ok(t, err)
	var pubKeyArmor bytes.Buffer
	w, err := armor.Encode(&pubKeyArmor, openpgp.PublicKeyType, nil)
	Ok(t, err)
	Ok(t, signer.Serialize(w))
	Ok(t, w.Close())

	archive := validArchive(t)
	sum := sha256.Sum256(archive)
	shasums := []byte(hex.EncodeToString(sum[:]) + "  " + archiveFilename + "\n")
	var sig bytes.Buffer
	// Signed while the key is still comfortably valid.
	Ok(t, openpgp.DetachSign(&sig, signer, bytes.NewReader(shasums), nil))

	meta := map[string]any{
		"filename":              archiveFilename,
		"download_url":          "https://upstream.example/archive.zip",
		"shasums_url":           "https://upstream.example/SHA256SUMS",
		"shasums_signature_url": "https://upstream.example/SHA256SUMS.sig",
		"shasum":                hex.EncodeToString(sum[:]),
		"signing_keys": map[string]any{
			"gpg_public_keys": []any{
				map[string]any{"ascii_armor": pubKeyArmor.String()},
			},
		},
	}
	files := map[string][]byte{
		"https://upstream.example/archive.zip":    archive,
		"https://upstream.example/SHA256SUMS":     shasums,
		"https://upstream.example/SHA256SUMS.sig": sig.Bytes(),
	}
	in := newTestInstaller(t, files)

	// Let the key's 1-second lifetime lapse before verifying, so a
	// wall-clock-based check would see it as expired.
	time.Sleep(2 * time.Second)

	Ok(t, in.install(context.Background(), testCoordinate, meta))
	Assert(t, in.isPublished(testCoordinate), "a validly-signed archive must still install after its signing key's lifetime has since elapsed")
}

func TestInstaller_RejectsMissingSigningKeys(t *testing.T) {
	meta, files := validMeta(t)
	meta["signing_keys"] = map[string]any{"gpg_public_keys": []any{}}

	in := newTestInstaller(t, files)
	err := in.install(context.Background(), testCoordinate, meta)
	Assert(t, err != nil, "expected an error when no signing keys are offered")
	Assert(t, !in.isPublished(testCoordinate), "an unverifiable archive must never be published")
}

// Test that install refuses a registry-supplied artifact URL that isn't
// https (or loopback http) - the same SSRF guard handleArtifact applies to
// the shasums/signature artifact endpoint (allowedArtifactURL) - for all
// three of the download/shasums/signature URLs, not just the archive one,
// and never touches fetch for a rejected URL (a compromised or
// misconfigured registry can't use any of the three fields to make the
// server issue an arbitrary request, e.g. to a cloud metadata endpoint).
func TestInstaller_RejectsUnsafeArtifactURLs(t *testing.T) {
	fields := []string{"download_url", "shasums_url", "shasums_signature_url"}
	unsafe := []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://internal.example.com/secret",
		"ftp://upstream.example/archive.zip",
	}
	for _, field := range fields {
		for _, badURL := range unsafe {
			t.Run(field+"="+badURL, func(t *testing.T) {
				meta, _ := validMeta(t)
				meta[field] = badURL
				in := &installer{
					log:       logging.NewNoopLogger(t),
					mirrorDir: filepath.Join(t.TempDir(), "mirror"),
					fetch: func(_ context.Context, rawURL string) (string, error) {
						t.Fatalf("fetch must never be called for a rejected url, got %s", rawURL)
						return "", nil
					},
				}
				err := in.install(context.Background(), testCoordinate, meta)
				Assert(t, err != nil, "expected an unsafe %s to be rejected", field)
				Assert(t, !in.isPublished(testCoordinate), "must never publish when an artifact url is rejected")
			})
		}
	}
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

// Test that EnsureInstalled on a coordinate that's already published bumps
// its mirror directory's mtime instead of merely no-op'ing - this is the only
// "still in active use" signal a warm coordinate ever gets (see
// touchPublished's doc comment), since Terraform reads the mirror directly
// on every subsequent init and never asks the proxy again.
func TestInstaller_EnsureInstalledTouchesAlreadyPublishedCoordinate(t *testing.T) {
	meta, files := validMeta(t)
	in := newTestInstaller(t, files)
	Ok(t, in.install(context.Background(), testCoordinate, meta))

	mirrorPath := testCoordinate.mirrorPath(in.mirrorDir)
	old := time.Now().Add(-2 * time.Hour)
	Ok(t, os.Chtimes(mirrorPath, old, old))

	// Synchronous: the already-published branch of EnsureInstalled never
	// touches the network or the singleflight group, so no waiting is needed.
	in.EnsureInstalled(testCoordinate, meta)

	info, err := os.Stat(mirrorPath)
	Ok(t, err)
	Assert(t, info.ModTime().After(old), "expected EnsureInstalled to bump the published coordinate's mtime past %s, got %s", old, info.ModTime())
}

// mkMirrorDir creates coordinate's mirror directory with a single fake file
// in it (a real install always publishes at least one file; sweepMirrorLevel
// only ever recurses into and prunes directories, so an empty one wouldn't
// exercise the same path) and backdates it to age old.
func mkMirrorDir(t *testing.T, mirrorDir string, c Coordinate, age time.Duration) string {
	t.Helper()
	dir := c.mirrorPath(mirrorDir)
	Ok(t, os.MkdirAll(dir, 0o700))
	Ok(t, os.WriteFile(filepath.Join(dir, providerBinaryName), []byte("fake"), 0o600))
	when := time.Now().Add(-age)
	Ok(t, os.Chtimes(dir, when, when))
	return dir
}

// Test that sweepMirror removes an installed provider version whose mirror
// directory mtime is older than maxAge, while leaving a fresher sibling
// version (same host/namespace/type, different version) untouched.
func TestInstaller_SweepMirrorRemovesExpiredKeepsFresh(t *testing.T) {
	in := &installer{log: logging.NewNoopLogger(t), mirrorDir: t.TempDir()}

	expired := Coordinate{Host: registryHost, Namespace: "hashicorp", Type: "null", Version: "3.2.1", OS: "linux", Arch: "amd64"}
	fresh := Coordinate{Host: registryHost, Namespace: "hashicorp", Type: "null", Version: "3.3.0", OS: "linux", Arch: "amd64"}
	expiredDir := mkMirrorDir(t, in.mirrorDir, expired, 2*time.Hour)
	freshDir := mkMirrorDir(t, in.mirrorDir, fresh, time.Minute)

	Ok(t, in.sweepMirror(time.Hour, time.Now()))

	Assert(t, !in.isPublished(expired), "expected the expired provider version to be removed")
	_, err := os.Stat(expiredDir)
	Assert(t, os.IsNotExist(err), "expected %s to no longer exist", expiredDir)

	Assert(t, in.isPublished(fresh), "expected the fresh provider version to survive the sweep")
	_, err = os.Stat(freshDir)
	Ok(t, err)
	// The type directory ("null") is still shared with the surviving fresh
	// version, so it must not have been pruned away.
	_, err = os.Stat(filepath.Dir(filepath.Dir(expiredDir)))
	Ok(t, err)
}

// Test that removing the only installed provider version under a host prunes
// every now-empty ancestor directory it leaves behind (version, type,
// namespace, host) - mirroring the manual cleanup an operator would
// otherwise have to run by hand on the old shared plugin-cache dir.
func TestInstaller_SweepMirrorPrunesEmptyAncestors(t *testing.T) {
	in := &installer{log: logging.NewNoopLogger(t), mirrorDir: t.TempDir()}

	c := Coordinate{Host: registryHost, Namespace: "hashicorp", Type: "null", Version: "3.2.1", OS: "linux", Arch: "amd64"}
	mkMirrorDir(t, in.mirrorDir, c, 2*time.Hour)

	Ok(t, in.sweepMirror(time.Hour, time.Now()))

	hostDir := filepath.Join(in.mirrorDir, c.Host)
	_, err := os.Stat(hostDir)
	Assert(t, os.IsNotExist(err), "expected the now-empty host directory %s to be pruned", hostDir)

	// mirrorDir itself is never pruned - only its contents.
	_, err = os.Stat(in.mirrorDir)
	Ok(t, err)
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
