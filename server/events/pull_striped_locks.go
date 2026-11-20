// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"hash/fnv"
	"strconv"
	"sync"
)

// applyResultLocks serializes recording apply results and publishing the
// combined apply status per pull request. It is shared by every
// ApplyCommandRunner in the process.
var applyResultLocks pullStripedLocks

// pullStripedLocks is a fixed set of mutexes keyed by pull request. Two pull
// requests may share a stripe, which only costs a little waiting; memory stays
// bounded no matter how many pull requests Atlantis sees. The zero value is
// ready to use.
type pullStripedLocks struct {
	stripes [64]sync.Mutex
}

// lock locks the stripe for repoFullName/pullNum and returns its unlock.
func (l *pullStripedLocks) lock(repoFullName string, pullNum int) func() {
	h := fnv.New32a()
	_, _ = h.Write([]byte(repoFullName))
	_, _ = h.Write([]byte{'#'})
	_, _ = h.Write([]byte(strconv.Itoa(pullNum)))
	m := &l.stripes[h.Sum32()%uint32(len(l.stripes))]
	m.Lock()
	return m.Unlock
}
