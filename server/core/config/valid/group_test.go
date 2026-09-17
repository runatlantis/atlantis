// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package valid_test

import (
	"testing"

	"github.com/runatlantis/atlantis/server/core/config/valid"
	. "github.com/runatlantis/atlantis/testing"
)

func TestValidateGroupName(t *testing.T) {
	cases := []struct {
		group  string
		expErr string
	}{
		{group: "infra", expErr: ""},
		{group: "infra-prod", expErr: ""},
		{group: "infra_v1.2~x", expErr: ""},
		{group: "-infra", expErr: ""},
		{group: "", expErr: "cannot be empty"},
		{group: "in*fra", expErr: "cannot contain glob pattern characters ('*', '?', '[')"},
		{group: "in?fra", expErr: "cannot contain glob pattern characters ('*', '?', '[')"},
		{group: "in[fra", expErr: "cannot contain glob pattern characters ('*', '?', '[')"},
		{group: "infra prod", expErr: "must contain only URL safe characters"},
		{group: "infra%", expErr: "must contain only URL safe characters"},
		// A group is matched verbatim against a selector, so '/' has to be
		// rejected here even though it is legal in a project name.
		{group: "infra/prod", expErr: "must contain only URL safe characters"},
	}
	for _, c := range cases {
		t.Run(c.group, func(t *testing.T) {
			err := valid.ValidateGroupName(c.group)
			if c.expErr == "" {
				Ok(t, err)
				return
			}
			ErrEquals(t, c.expErr, err)
		})
	}
}
