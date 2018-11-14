// Copyright 2017 HootSuite Media Inc.
//
// Licensed under the Apache License, Version 2.0 (the License);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an AS IS BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
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
