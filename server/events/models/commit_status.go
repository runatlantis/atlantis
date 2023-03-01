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

// VCSStatus is the result of executing an Atlantis command for the commit.
// In Github the options are: error, failure, pending, success.
// In Gitlab the options are: failed, canceled, pending, running, success.
// We only support Failed, Pending, Success.
type VCSStatus int

const (
	PendingVCSStatus VCSStatus = iota
	SuccessVCSStatus
	FailedVCSStatus
	QueuedVCSStatus
)

func (s VCSStatus) String() string {
	switch s {
	case PendingVCSStatus:
		return "pending"
	case SuccessVCSStatus:
		return "success"
	case FailedVCSStatus:
		return "failed"
	case QueuedVCSStatus:
		return "queued"
	}
	return "failed"
}
