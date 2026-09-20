// Copyright 2024 contributors to runatlantis/atlantis.
// SPDX-License-Identifier: Apache-2.0

package providercache

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp"
	"golang.org/x/sync/singleflight"

	"github.com/runatlantis/atlantis/server/logging"
)

// Coordinate identifies one provider package: a specific provider, version
// and target platform on one registry host.
type Coordinate struct {
	Host, Namespace, Type, Version, OS, Arch string
}

// key uniquely identifies the coordinate for de-duplication purposes.
func (c Coordinate) key() string {
	return strings.Join([]string{c.Host, c.Namespace, c.Type, c.Version, c.OS, c.Arch}, "/")
}

// mirrorPath is where this coordinate is published under mirrorDir, laid out
// exactly like Terraform's own `provider_installation.filesystem_mirror`
// "unpacked" layout (HOSTNAME/NAMESPACE/TYPE/VERSION/TARGET), so Terraform can
// be pointed straight at mirrorDir with no translation.
func (c Coordinate) mirrorPath(mirrorDir string) string {
	return filepath.Join(mirrorDir, c.Host, c.Namespace, c.Type, c.Version, c.OS+"_"+c.Arch)
}

// installer is the sole writer of the provider mirror directory: it fetches,
// verifies (SHA256SUMS + GPG signature) and unpacks provider archives itself,
// so that Terraform only ever reads from the mirror (via a filesystem_mirror
// CLI config) instead of installing into a shared directory itself. That is
// what makes the shared directory safe under concurrency: Terraform's own
// provider installer, which is not safe for concurrent writers (see
// hashicorp/terraform#25849), never touches it.
type installer struct {
	log logging.SimpleLogging

	// mirrorDir is the root of the filesystem_mirror-formatted directory.
	mirrorDir string

	// installTimeout bounds a single install attempt (download + verify +
	// unpack), so a stalled upstream (the download has no other timeout -
	// provider archives can be large) can't wedge a coordinate's singleflight
	// key forever. Falls back to defaultInstallTimeout when zero.
	installTimeout time.Duration

	// fetch resolves an artifact URL (archive/shasums/signature) to an
	// on-disk path, reusing the proxy's existing download cache/dedup
	// (Server.ensureCached) rather than re-implementing it.
	fetch func(ctx context.Context, rawURL string) (string, error)

	// sf ensures only one goroutine ever downloads+verifies+publishes a
	// given coordinate at a time, regardless of how many concurrent archive
	// requests name it. A completed call (success or failure) is forgotten
	// once done, so a later EnsureInstalled for the same still-unpublished
	// coordinate starts a fresh attempt - that's what lets a transient
	// failure be retried rather than being permanent for the process
	// lifetime.
	sf singleflight.Group

	// errMu guards lastErr.
	errMu sync.Mutex
	// lastErr records the most recent install failure per coordinate (see
	// Coordinate.key), so a caller stuck waiting on the mirror can surface
	// the real cause instead of a generic "not found" once it gives up.
	// Cleared on a subsequent successful install of that coordinate.
	lastErr map[string]error

	// installs counts completed install attempts, for tests.
	installs atomic.Int64
}

// defaultInstallTimeout is used when installTimeout is unset.
const defaultInstallTimeout = 2 * time.Minute

// EnsureInstalled makes sure the provider archive for coordinate is (or soon
// will be) verified and unpacked under mirrorDir. It never blocks the caller
// on the network: if an install isn't already in flight for this exact
// coordinate, one is started in the background - deliberately decoupled from
// any single HTTP request's context, since the request that triggered it
// (and every concurrent sibling request for the same coordinate) will
// normally be long gone (they all get a 423 immediately) before the install
// finishes - and EnsureInstalled returns right away either way.
//
// meta is the registry's "find a package" response for coordinate (see
// Server.downloadMetadata), which carries the archive/shasums/signature URLs
// and the signing keys needed to verify them.
func (in *installer) EnsureInstalled(coordinate Coordinate, meta map[string]any) {
	if in.isPublished(coordinate) {
		return
	}
	ch := in.sf.DoChan(coordinate.key(), func() (any, error) {
		// Re-check under the singleflight barrier: a sibling call (or a
		// previous Atlantis process run) may have already published this
		// coordinate while we were waiting to run.
		if in.isPublished(coordinate) {
			return nil, nil
		}
		in.log.Info("provider cache: installing %s into the mirror", coordinate.key())
		timeout := in.installTimeout
		if timeout <= 0 {
			timeout = defaultInstallTimeout
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		err := in.install(ctx, coordinate, meta)
		in.installs.Add(1)
		in.recordResult(coordinate, err)
		return nil, err
	})
	go func() {
		if res := <-ch; res.Err != nil {
			in.log.Err("provider cache: installing %s: %s", coordinate.key(), res.Err)
		}
	}()
}

// recordResult remembers err (or forgets a prior failure, on success) as the
// most recent install outcome for coordinate, so LastErrors can later surface
// it to a caller that gave up waiting on the mirror.
func (in *installer) recordResult(c Coordinate, err error) {
	in.errMu.Lock()
	defer in.errMu.Unlock()
	if err == nil {
		delete(in.lastErr, c.key())
		return
	}
	if in.lastErr == nil {
		in.lastErr = make(map[string]error)
	}
	in.lastErr[c.key()] = err
}

// LastErrors returns the most recent install failure message for every
// coordinate on host with a currently-recorded error (i.e. one whose most
// recent install attempt failed and has not since succeeded).
func (in *installer) LastErrors(host string) []string {
	in.errMu.Lock()
	defer in.errMu.Unlock()
	prefix := host + "/"
	var out []string
	for key, err := range in.lastErr {
		if strings.HasPrefix(key, prefix) {
			out = append(out, key+": "+err.Error())
		}
	}
	return out
}

// isPublished reports whether coordinate is already fully installed in the
// mirror. Publication is atomic (temp dir + rename, see publish), so
// directory existence alone is a reliable completion marker: there is no
// partially-written state a reader can observe.
func (in *installer) isPublished(c Coordinate) bool {
	info, err := os.Stat(c.mirrorPath(in.mirrorDir))
	return err == nil && info.IsDir()
}

// install fetches, verifies and unpacks the provider archive for coordinate,
// then atomically publishes it under mirrorDir. Nothing is published unless
// every check below passes; any failure leaves the mirror untouched.
func (in *installer) install(ctx context.Context, c Coordinate, meta map[string]any) error {
	archiveURL, _ := meta["download_url"].(string)
	shasumsURL, _ := meta["shasums_url"].(string)
	sigURL, _ := meta["shasums_signature_url"].(string)
	filename, _ := meta["filename"].(string)
	expectedSum, _ := meta["shasum"].(string)
	if archiveURL == "" || shasumsURL == "" || sigURL == "" || filename == "" || expectedSum == "" {
		return errors.New("registry metadata is missing required fields")
	}

	keyring, err := signingKeyRing(meta)
	if err != nil {
		return fmt.Errorf("resolving signing keys: %w", err)
	}
	if len(keyring) == 0 {
		// Fail closed: a provider we cannot verify is a provider we will not
		// silently install into a directory every concurrent `terraform
		// init` implicitly trusts.
		return errors.New("registry did not provide any signing keys for this provider; refusing to install an unverifiable package")
	}

	archivePath, err := in.fetch(ctx, archiveURL)
	if err != nil {
		return fmt.Errorf("downloading archive: %w", err)
	}
	shasumsPath, err := in.fetch(ctx, shasumsURL)
	if err != nil {
		return fmt.Errorf("downloading SHA256SUMS: %w", err)
	}
	sigPath, err := in.fetch(ctx, sigURL)
	if err != nil {
		return fmt.Errorf("downloading SHA256SUMS signature: %w", err)
	}

	// #nosec G304 -- these paths are our own cache paths (Server.cachePath),
	// not attacker-controlled input.
	shasums, err := os.ReadFile(shasumsPath)
	if err != nil {
		return fmt.Errorf("reading cached SHA256SUMS: %w", err)
	}
	// #nosec G304 -- see above.
	sig, err := os.ReadFile(sigPath)
	if err != nil {
		return fmt.Errorf("reading cached SHA256SUMS signature: %w", err)
	}
	if _, err := openpgp.CheckDetachedSignature(keyring, bytes.NewReader(shasums), bytes.NewReader(sig), nil); err != nil {
		return fmt.Errorf("verifying SHA256SUMS signature: %w", err)
	}

	wantSum, err := shasumFor(shasums, filename)
	if err != nil {
		return err
	}
	if !strings.EqualFold(wantSum, expectedSum) {
		return fmt.Errorf("registry metadata shasum %q does not match the signed SHA256SUMS entry %q for %s", expectedSum, wantSum, filename)
	}

	// #nosec G304 -- archivePath is our own cache path, not attacker-controlled input.
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("reading cached archive: %w", err)
	}
	gotSum := sha256.Sum256(archiveBytes)
	if !strings.EqualFold(hex.EncodeToString(gotSum[:]), wantSum) {
		return errors.New("downloaded archive does not match its signed SHA256SUMS checksum")
	}

	return in.publish(c, archiveBytes)
}

// signingKeyRing collects every GPG public key the registry offered for this
// provider (meta["signing_keys"]["gpg_public_keys"][].ascii_armor) into a
// single keyring. A provider is trusted if its SHA256SUMS signature verifies
// against any one of them, matching Terraform's own behavior.
func signingKeyRing(meta map[string]any) (openpgp.EntityList, error) {
	signingKeys, _ := meta["signing_keys"].(map[string]any)
	if signingKeys == nil {
		return nil, nil
	}
	rawKeys, _ := signingKeys["gpg_public_keys"].([]any)
	var all openpgp.EntityList
	for _, rk := range rawKeys {
		km, ok := rk.(map[string]any)
		if !ok {
			continue
		}
		armor, _ := km["ascii_armor"].(string)
		if armor == "" {
			continue
		}
		entities, err := openpgp.ReadArmoredKeyRing(strings.NewReader(armor))
		if err != nil {
			return nil, fmt.Errorf("parsing signing key: %w", err)
		}
		all = append(all, entities...)
	}
	return all, nil
}

// shasumFor returns the hex SHA-256 digest recorded for filename in the
// SHA256SUMS file contents shasums (lines of "<hex>  <filename>").
func shasumFor(shasums []byte, filename string) (string, error) {
	for line := range strings.Lines(string(shasums)) {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return strings.ToLower(fields[0]), nil
		}
	}
	return "", fmt.Errorf("no SHA256SUMS entry for %s", filename)
}

// publish atomically makes archiveBytes, already fully verified by the
// caller, visible at coordinate's mirror path: it is unpacked into a temp
// directory alongside the mirror root and then moved into place with a
// single rename, so a concurrent reader (or a concurrent EnsureInstalled
// call, including from another Atlantis process) never observes a partially
// written directory.
func (in *installer) publish(c Coordinate, archiveBytes []byte) error {
	dest := c.mirrorPath(in.mirrorDir)
	if _, err := os.Stat(dest); err == nil {
		// Already published by a sibling call, or a previous run - nothing
		// to do.
		return nil
	}

	if err := os.MkdirAll(in.mirrorDir, 0o700); err != nil {
		return err
	}
	tmpRoot, err := os.MkdirTemp(in.mirrorDir, ".install-*")
	if err != nil {
		return err
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(tmpRoot)
		}
	}()

	if err := unzip(archiveBytes, tmpRoot); err != nil {
		return fmt.Errorf("unpacking archive: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o700); err != nil {
		return err
	}
	if err := os.Rename(tmpRoot, dest); err != nil {
		// A sibling install (possibly from another Atlantis process sharing
		// this mirror dir) may have won the race to publish this exact
		// coordinate; POSIX rename cannot atomically replace a directory
		// that already exists, so that shows up as a rename error, not a
		// pre-check hit. Treat it as success rather than failing the whole
		// install.
		if _, statErr := os.Stat(dest); statErr == nil {
			published = true
			return nil
		}
		return fmt.Errorf("publishing to mirror: %w", err)
	}
	published = true
	return nil
}

// unzip extracts a provider distribution archive (a flat zip of a single
// platform's binary plus metadata files, as HashiCorp's provider archives
// are) into destDir.
func unzip(data []byte, destDir string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		name := filepath.Clean(f.Name)
		if name == "." || filepath.IsAbs(name) || name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive entry %q has an unsafe path", f.Name)
		}
		target := filepath.Join(destDir, name)
		if target != destDir && !strings.HasPrefix(target, destDir+string(filepath.Separator)) {
			return fmt.Errorf("archive entry %q escapes the destination directory", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o700); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
			return err
		}
		if err := extractFile(f, target); err != nil {
			return err
		}
	}
	return nil
}

// extractFile writes one zip entry to target, preserving its executable bit
// (provider archives store the exec permission on the provider binary).
func extractFile(f *zip.File, target string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	mode := f.Mode().Perm()
	if mode == 0 {
		mode = 0o600
	}
	// #nosec G304 -- target is validated to stay under destDir by unzip above.
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	// #nosec G110 -- the archive was checksum + GPG-signature verified
	// before extraction and provider archives are at most a few tens of MB,
	// so a decompression-bomb guard isn't warranted here.
	if _, err := io.Copy(out, rc); err != nil {
		return err
	}
	return out.Close()
}
