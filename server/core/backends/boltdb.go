// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package backends

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	bolt "go.etcd.io/bbolt"
)

// openBoltDB opens (creating if needed) the atlantis.db bolt file under
// dataDir. Bucket creation and migrations belong to the store drivers.
func openBoltDB(dataDir string) (*bolt.DB, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("creating data dir: %w", err)
	}
	db, err := bolt.Open(path.Join(dataDir, "atlantis.db"), 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if err.Error() == "timeout" {
			return nil, errors.New("starting BoltDB: timeout (a possible cause is another Atlantis instance already running)")
		}
		return nil, fmt.Errorf("starting BoltDB: %w", err)
	}
	return db, nil
}
