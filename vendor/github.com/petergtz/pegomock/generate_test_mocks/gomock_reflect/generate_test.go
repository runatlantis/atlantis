// Copyright 2015 Peter Goetz
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mockgen_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/petergtz/pegomock/pegomock/filehandling"
)

func TestMockGeneration(t *testing.T) { RunSpecs(t, "Generating mocks with GoMock-reflect") }

var _ = It("Generate mocks", func() {
	filehandling.GenerateMockFile(
		[]string{"github.com/petergtz/pegomock/test_interface", "Display"},
		"../../mock_display_test.go", "pegomock_test",
		"", false, os.Stdout, false, true)
})
