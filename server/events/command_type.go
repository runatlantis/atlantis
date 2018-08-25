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

package events

// CommandName is which command to run.
type CommandName int

const (
	// ApplyCommand is a command to run terraform apply.
	ApplyCommand CommandName = iota
	// PlanCommand is a command to run terraform plan.
	PlanCommand
	// Adding more? Don't forget to update String() below
)

// String returns the string representation of c.
func (c CommandName) String() string {
	switch c {
	case ApplyCommand:
		return "apply"
	case PlanCommand:
		return "plan"
	}
	return ""
}
