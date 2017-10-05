// Copyright 2016 Peter Goetz
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

package testutil

import (
	"fmt"
	"io/ioutil"
	"strings"
)

type fileContentSubStringMatcher struct {
	substring   string
	fileContent []byte
}

func BeAFileContainingSubString(substring string) *fileContentSubStringMatcher {
	return &fileContentSubStringMatcher{substring: substring}
}

func (matcher *fileContentSubStringMatcher) Match(actual interface{}) (bool, error) {
	actualFilePath, ok := actual.(string)
	if !ok {
		return false, fmt.Errorf("Matcher expects the actual file path as string")
	}
	var err error
	matcher.fileContent, err = ioutil.ReadFile(actualFilePath)
	if err != nil {
		return false, fmt.Errorf("File content could not be read due to: %v", err)
	}
	return strings.Contains(string(matcher.fileContent), matcher.substring), nil
}

func (matcher *fileContentSubStringMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n\t%s to contain a substring '%s'\n\nBut got:\n%s",
		actual, matcher.substring, matcher.fileContent)
}

func (matcher *fileContentSubStringMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected:\n\t%s not to contain a substring '%s'\n\nBut got:\n%s",
		actual, matcher.substring, matcher.fileContent)
}
