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
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"
)

type E2ETester struct {
	vcsClient    VCSClient
	hookID       int64
	cloneDirRoot string
	projectType  Project
}

type E2EResult struct {
	projectType    string
	pullRequestURL string
	testResult     string
}

var testFileData = `
resource "null_resource" "hello" {
}
`

// nolint: gosec
func (t *E2ETester) Start(ctx context.Context) (*E2EResult, error) {
	cloneDir := fmt.Sprintf("%s/%s-test", t.cloneDirRoot, t.projectType.Name)
	branchName := fmt.Sprintf("%s-%s", t.projectType.Name, time.Now().Format("20060102150405"))
	testFileName := fmt.Sprintf("%s.tf", t.projectType.Name)
	e2eResult := &E2EResult{}
	e2eResult.projectType = t.projectType.Name

	// create the directory and parents if necessary
	log.Printf("creating dir %q", cloneDir)
	if err := os.MkdirAll(cloneDir, 0700); err != nil {
		return e2eResult, fmt.Errorf("failed to create dir %q prior to cloning, attempting to continue: %v", cloneDir, err)
	}

	err := t.vcsClient.Clone(cloneDir)
	if err != nil {
		return e2eResult, err
	}

	// checkout a new branch for the project
	log.Printf("checking out branch %q", branchName)
	checkoutCmd := exec.Command("git", "checkout", "-b", branchName)
	checkoutCmd.Dir = cloneDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		return e2eResult, fmt.Errorf("failed to git checkout branch %q: %v: %s", branchName, err, string(output))
	}

	// write a file for running the tests
	randomData := []byte(testFileData)
	filePath := fmt.Sprintf("%s/%s/%s", cloneDir, t.projectType.Name, testFileName)
	log.Printf("creating file to commit %q", filePath)
	err = os.WriteFile(filePath, randomData, 0644)
	if err != nil {
		return e2eResult, fmt.Errorf("couldn't write file %s: %v", filePath, err)
	}

	// add the file
	log.Printf("git add file %q", filePath)
	addCmd := exec.Command("git", "add", filePath)
	addCmd.Dir = cloneDir
	if output, err := addCmd.CombinedOutput(); err != nil {
		return e2eResult, fmt.Errorf("failed to git add file %q: %v: %s", filePath, err, string(output))
	}

	// commit the file
	log.Printf("git commit file %q", filePath)
	commitCmd := exec.Command("git", "commit", "-am", "test commit")
	commitCmd.Dir = cloneDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		return e2eResult, fmt.Errorf("failed to run git commit in %q: %v: %v", cloneDir, err, string(output))
	}

	// push the branch to remote
	log.Printf("git push branch %q", branchName)
	pushCmd := exec.Command("git", "push", "origin", branchName)
	pushCmd.Dir = cloneDir
	if output, err := pushCmd.CombinedOutput(); err != nil {
		return e2eResult, fmt.Errorf("failed to git push branch %q: %v: %s", branchName, err, string(output))
	}

	// create a new pr
	title := fmt.Sprintf("This is a test pull request for atlantis e2e test for %s project type", t.projectType.Name)
	url, pullId, err := t.vcsClient.CreatePullRequest(ctx, title, branchName)

	if err != nil {
		return e2eResult, err
	}

	// set pull request url
	e2eResult.pullRequestURL = url

	log.Printf("created pull request %s", url)

	// defer closing pull request and delete remote branch
	defer cleanUp(ctx, t, pullId, branchName) // nolint: errcheck

	// wait for atlantis to respond to webhook and autoplan.
	time.Sleep(2 * time.Second)

	state := "not started"
	// waiting for atlantis run and finish
	maxLoops := 20
	i := 0
	for ; i < maxLoops && checkStatus(state); i++ {
		time.Sleep(2 * time.Second)
		state, _ = t.vcsClient.GetAtlantisStatus(ctx, branchName)
		if state == "" {
			log.Println("atlantis run hasn't started")
			continue
		}
		log.Printf("atlantis run is in %s state", state)
	}
	if i == maxLoops {
		state = "timed out"
	}

	log.Printf("atlantis run finished with status %q", state)
	e2eResult.testResult = state
	// check if atlantis run was a success
	if state != "success" {
		return e2eResult, fmt.Errorf("atlantis run project type %q failed with %q status", t.projectType.Name, state)
	}

	return e2eResult, nil
}

func checkStatus(state string) bool {
	for _, s := range []string{"success", "error", "failure"} {
		if state == s {
			return false
		}
	}
	return true
}

func cleanUp(ctx context.Context, t *E2ETester, pullRequestNumber int, branchName string) error {
	// clean up
	err := t.vcsClient.ClosePullRequest(ctx, pullRequestNumber)
	if err != nil {
		return err
	}
	log.Printf("closed pull request %d", pullRequestNumber)

	deleteBranchName := fmt.Sprintf("%s/%s", "heads", branchName)
	err = t.vcsClient.DeleteBranch(ctx, deleteBranchName)
	if err != nil {
		return fmt.Errorf("error while deleting branch %s: %v", deleteBranchName, err)
	}
	log.Printf("deleted branch %s", deleteBranchName)

	return nil
}
