// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"
)

// Webhook is a repository webhook as the e2e tests see it.
type Webhook struct {
	ID  int64
	URL string
}

// tunnelHostSuffixes are the ngrok domains scripts/e2e.sh gets tunnel URLs on.
var tunnelHostSuffixes = []string{
	".ngrok.app",
	".ngrok.dev",
	".ngrok.io",
	".ngrok-free.app",
	".ngrok-free.dev",
}

// isLeftoverE2EHook reports whether a webhook was created by an earlier e2e
// run: it points at the Atlantis events endpoint through an ngrok tunnel.
// Hooks with any other URL are never touched.
func isLeftoverE2EHook(hookURL string) bool {
	u, err := url.Parse(hookURL)
	if err != nil || u.Scheme != "https" || u.Path != "/events" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	for _, suffix := range tunnelHostSuffixes {
		if strings.HasSuffix(host, suffix) && len(host) > len(suffix) {
			return true
		}
	}
	return false
}

// sleep waits for d, or less if ctx is done, so an interrupted run reaches
// its cleanup before the runner kills it.
func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

// deleteLeftoverHooks removes webhooks left by earlier e2e runs that were
// cancelled or killed before they could clean up. GitHub allows 20 webhooks
// per event, so leftovers eventually stop new runs from creating theirs.
//
// The e2e workflow runs one job per VCS at a time (see the concurrency groups
// in .github/workflows/test.yml), so no other live run owns these hooks.
func deleteLeftoverHooks(ctx context.Context, vcs VCSClient) error {
	hooks, err := vcs.ListAtlantisHooks(ctx)
	if err != nil {
		return fmt.Errorf("listing webhooks: %w", err)
	}
	deleted := 0
	for _, h := range hooks {
		if !isLeftoverE2EHook(h.URL) {
			continue
		}
		if err := vcs.DeleteAtlantisHook(ctx, h.ID); err != nil {
			return fmt.Errorf("deleting leftover webhook %d: %w", h.ID, err)
		}
		deleted++
	}
	log.Printf("deleted %d leftover e2e webhooks", deleted)
	return nil
}
