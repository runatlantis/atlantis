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
//
package events

//go:generate pegomock generate -m --use-experimental-model-gen --package mocks -o mocks/mock_executor.go Executor

// Executor is the generic interface implemented by each command type:
// help, plan, and apply.
type Executor interface {
	Execute(ctx *CommandContext) CommandResult
}
