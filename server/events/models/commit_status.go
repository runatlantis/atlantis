// Copyright 2017 HootSuite Media Inc.
// SPDX-License-Identifier: Apache-2.0
// Modified hereafter by contributors to runatlantis/atlantis.

package models

// CommitStatus is the result of executing an Atlantis command for the commit.
// In Github the options are: error, failure, pending, success.
// In Gitlab the options are: failed, canceled, pending, running, success.
// We only support Failed, Pending, Success.
type CommitStatus int

const (
	PendingCommitStatus CommitStatus = iota
	SuccessCommitStatus
	FailedCommitStatus
)

func (s CommitStatus) String() string {
	switch s {
	case PendingCommitStatus:
		return "pending"
	case SuccessCommitStatus:
		return "success"
	case FailedCommitStatus:
		return "failed"
	}
	return "failed"
}
