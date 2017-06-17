package main

import (
	"context"
	"log"
	"os"
	"strings"

	"fmt"

	"github.com/google/go-github/github"
	"github.com/hashicorp/go-multierror"
)

var defaultAtlantisURL = "http://localhost:4141"
var projectTypes = []Project{
	Project{"standalone", "run plan", "run apply"},
	Project{"standalone-with-env", "run plan staging", "run apply staging"},
}

type Project struct {
	Name         string
	PlanCommand  string
	ApplyCommand string
}

func main() {

	githubUsername := os.Getenv("GITHUB_USERNAME")
	if githubUsername == "" {
		log.Fatalf("GITHUB_USERNAME cannot be empty")
	}
	githubPassword := os.Getenv("GITHUB_PASSWORD")
	if githubPassword == "" {
		log.Fatalf("GITHUB_PASSWORD cannot be empty")
	}
	atlantisURL := os.Getenv("ATLANTIS_URL")
	if atlantisURL == "" {
		atlantisURL = defaultAtlantisURL
	}
	// add /hooks to the url
	atlantisURL = fmt.Sprintf("%s/hooks", atlantisURL)
	ownerName := os.Getenv("GITHUB_REPO_OWNER_NAME")
	if ownerName == "" {
		ownerName = "anubhavmishra"
	}
	repoName := os.Getenv("GITHUB_REPO_NAME")
	if repoName == "" {
		repoName = "atlantis-tests"
	}
	// using https to clone the repo
	repoUrl := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUsername, githubPassword, ownerName, repoName)
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

	// create github client
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(githubUsername),
		Password: strings.TrimSpace(githubPassword),
	}
	ghClient := github.NewClient(tp.Client())

	githubClient := &GithubClient{client: ghClient, ctx: context.Background(), username: githubUsername}

	// we create atlantis hook once for the repo, since the atlantis server can handle multiple requests
	log.Printf("creating atlantis webhook with %s url", atlantisURL)
	hookID, err := createAtlantisWebhook(githubClient, ownerName, repoName, atlantisURL)
	if err != nil {
		log.Fatalf("error creating atlantis webhook: %v", err)
	}

	// create e2e test
	e2e := E2ETester{
		githubClient: githubClient,
		repoUrl:      repoUrl,
		ownerName:    ownerName,
		repoName:     repoName,
		hookID:       hookID,
		cloneDirRoot: cloneDirRoot,
	}

	// start e2e tests
	results, err := startTests(e2e)
	log.Printf("Test Results\n---------------------------\n")
	for _, result := range results {
		fmt.Printf("Project Type: %s \n", result.projectType)
		fmt.Printf("Pull Request Link: %s \n", result.githubPullRequestURL)
		fmt.Printf("Atlantis Run Status: %s \n", result.testResult)
		fmt.Println("---------------------------")
	}
	if err != nil {
		log.Fatalf(fmt.Sprintf("%s", err))
	}

}

func createAtlantisWebhook(g *GithubClient, ownerName string, repoName string, hookURL string) (int, error) {
	// create atlantis hook
	atlantisHook := &github.Hook{
		Name:   github.String("web"),
		Events: []string{"issue_comment", "pull_request", "push"},
		Config: map[string]interface{}{
			"url":          hookURL,
			"content_type": "json",
		},
		Active: github.Bool(true),
	}

	// moved to github.go
	hook, _, err := g.client.Repositories.CreateHook(g.ctx, ownerName, repoName, atlantisHook)
	if err != nil {
		return 0, err
	}
	log.Println(hook.GetURL())

	return hook.GetID(), nil
}

func deleteAtlantisHook(g *GithubClient, ownerName string, repoName string, hookID int) error {
	_, err := g.client.Repositories.DeleteHook(g.ctx, ownerName, repoName, hookID)
	if err != nil {
		return err
	}
	log.Printf("deleted webhook id %d", hookID)

	return nil
}

func cleanDir(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return nil
}

func startTests(e2e E2ETester) ([]*E2EResult, error) {
	var testResults []*E2EResult
	var testErrors *multierror.Error
	// delete webhook when we are done running tests
	defer deleteAtlantisHook(e2e.githubClient, e2e.ownerName, e2e.repoName, e2e.hookID)

	for _, projectType := range projectTypes {
		log.Printf("starting e2e test for project type %q", projectType.Name)
		e2e.projectType = projectType
		// start e2e test
		result, err := e2e.Start()
		testResults = append(testResults, result)
		testErrors = multierror.Append(testErrors, err)
	}

	return testResults, testErrors.ErrorOrNil()
}
