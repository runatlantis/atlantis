// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package runtime

import "github.com/runatlantis/atlantis/server/core/planstore"

// Type aliases for backward compatibility after moving plan store types
// to the planstore package.
type PlanStore = planstore.PlanStore
type LocalPlanStore = planstore.LocalPlanStore
type S3PlanStoreConfig = planstore.S3PlanStoreConfig

var (
	NewS3PlanStore           = planstore.NewS3PlanStore
	NewS3PlanStoreWithClient = planstore.NewS3PlanStoreWithClient
	ErrRestoreNotSupported   = planstore.ErrRestoreNotSupported
)
