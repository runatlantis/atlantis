// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package metrics

import tally "github.com/uber-go/tally/v4"

func InitCounter(scope tally.Scope, name string) {
	s := scope.Counter(name)
	s.Inc(0)
}
