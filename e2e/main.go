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

package main

import (
	"context"
	"log"
	"os"

	"fmt"

	multierror "github.com/hashicorp/go-multierror"
)

var defaultAtlantisURL = "http://localhost:4141"
var projectTypes = []Project{
	{"standalone", "atlantis apply -d standalone"},
	{"standalone-with-workspace", "atlantis apply -d standalone-with-workspace -w staging"},
}

type Project struct {
	Name         string
	ApplyCommand string
}

func main() {

	atlantisURL := os.Getenv("ATLANTIS_URL")
	if atlantisURL == "" {
		atlantisURL = defaultAtlantisURL
	}
	// add /events to the url
	atlantisURL = fmt.Sprintf("%s/events", atlantisURL)

	cloneDirRoot := os.Getenv("CLONE_DIR")
	if cloneDirRoot == "" {
		cloneDirRoot = "/tmp/atlantis-tests"
	}

	// clean workspace
	log.Printf("cleaning workspace %s", cloneDirRoot)
	err := cleanDir(cloneDirRoot)
	if err != nil {
		log.Fatalf("failed to clean dir %q before cloning, attempting to continue: %v", cloneDirRoot, err)
	}

	githubClient := NewGithubClient()
	ctx := context.Background()
	// we create atlantis hook once for the repo, since the atlantis server can handle multiple requests
	log.Printf("creating atlantis webhook with %s url", atlantisURL)
	hookID, err := githubClient.CreateAtlantisWebhook(ctx, atlantisURL)
	if err != nil {
		log.Fatalf("error creating atlantis webhook: %v", err)
	}

	// create e2e test
	e2e := E2ETester{
		vcsClient:    githubClient,
		hookID:       hookID,
		cloneDirRoot: cloneDirRoot,
	}

	// start e2e tests
	results, err := startTests(ctx, e2e)
	log.Printf("Test Results\n---------------------------\n")
	for _, result := range results {
		fmt.Printf("Project Type: %s \n", result.projectType)
		fmt.Printf("Pull Request Link: %s \n", result.pullRequestURL)
		fmt.Printf("Atlantis Run Status: %s \n", result.testResult)
		fmt.Println("---------------------------")
	}
	if err != nil {
		log.Fatalf(fmt.Sprintf("%s", err))
	}

}

func cleanDir(path string) error {
	return os.RemoveAll(path)
}

func startTests(ctx context.Context, e2e E2ETester) ([]*E2EResult, error) {
	var testResults []*E2EResult
	var testErrors *multierror.Error
	// delete webhook when we are done running tests
	defer e2e.vcsClient.DeleteAtlantisHook(ctx, e2e.hookID) // nolint: errcheck

	for _, projectType := range projectTypes {
		log.Printf("starting e2e test for project type %q", projectType.Name)
		e2e.projectType = projectType
		// start e2e test
		result, err := e2e.Start(ctx)
		testResults = append(testResults, result)
		testErrors = multierror.Append(testErrors, err)
	}

	return testResults, testErrors.ErrorOrNil()
}
