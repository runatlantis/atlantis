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
// Package bootstrap is used by the bootstrap command as a quick-start of
// Atlantis.
package bootstrap

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/briandowns/spinner"
	"github.com/google/go-github/github"
	"github.com/mitchellh/colorstring"
	"github.com/pkg/errors"
)

var terraformExampleRepoOwner = "runatlantis"
var terraformExampleRepo = "atlantis-example"
var bootstrapDescription = `[white]Welcome to Atlantis bootstrap!

This mode walks you through setting up and using Atlantis. We will
- fork an example terraform project to your username
- install terraform (if not already in your PATH)
- install ngrok so we can expose Atlantis to GitHub
- start Atlantis

[bold]Press Ctrl-c at any time to exit
`
var pullRequestBody = "In this pull request we will learn how to use atlantis. There are various commands that are available to you:\n" +
	"* Start by typing `atlantis help` in the comments.\n" +
	"* Next, lets plan by typing `atlantis plan` in the comments. That will run a `terraform plan`.\n" +
	"* Now lets apply that plan. Type `atlantis apply` in the comments. This will run a `terraform apply`.\n" +
	"\nThank you for trying out atlantis. For more info on running atlantis in production see https://github.com/runatlantis/atlantis"

// Start begins the bootstrap process.
// nolint: errcheck
func Start() error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	colorstring.Println(bootstrapDescription)
	colorstring.Print("\n[white][bold]GitHub username: ")
	fmt.Scanln(&githubUsername)
	if githubUsername == "" {
		return fmt.Errorf("please enter a valid github username")
	}
	colorstring.Println(`
[white]To continue, we need you to create a GitHub personal access token
with [green]"repo" [white]scope so we can fork an example terraform project.

Follow these instructions to create a token (we don't store any tokens):
[green]https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/#creating-a-token
[white]- use "atlantis" for the token description
- add "repo" scope
- copy the access token
`)
	// Read github token, check for error later.
	colorstring.Print("[white][bold]GitHub access token (will be hidden): ")
	githubToken, _ = readPassword()
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(githubUsername),
		Password: strings.TrimSpace(githubToken),
	}
	githubClient := &Client{client: github.NewClient(tp.Client()), ctx: context.Background()}

	// Fork terraform example repo.
	colorstring.Println("\n[white]=> forking repo ")
	s.Start()
	if err := githubClient.CreateFork(terraformExampleRepoOwner, terraformExampleRepo); err != nil {
		return errors.Wrapf(err, "forking repo %s/%s", terraformExampleRepoOwner, terraformExampleRepo)
	}
	if !githubClient.CheckForkSuccess(terraformExampleRepoOwner, terraformExampleRepo) {
		return fmt.Errorf("didn't find forked repo %s/%s. fork unsuccessful", terraformExampleRepoOwner, terraformExampleRepoOwner)
	}
	s.Stop()
	colorstring.Println("[green]=> fork completed!")

	// Detect terraform and install it if not installed.
	_, err := exec.LookPath("terraform")
	if err != nil {
		colorstring.Println("[yellow]=> terraform not found in $PATH.")
		colorstring.Println("[white]=> downloading terraform ")
		s.Start()
		terraformDownloadURL := fmt.Sprintf("%s/terraform/%s/terraform_%s_%s_%s.zip", hashicorpReleasesURL, terraformVersion, terraformVersion, runtime.GOOS, runtime.GOARCH)
		if err = downloadAndUnzip(terraformDownloadURL, "/tmp/terraform.zip", "/tmp"); err != nil {
			return errors.Wrapf(err, "downloading and unzipping terraform")
		}
		colorstring.Println("[green]=> downloaded terraform successfully!")
		s.Stop()

		var terraformCmd *exec.Cmd
		terraformCmd, err = executeCmd("mv", []string{"/tmp/terraform", "/usr/local/bin/"})
		if err != nil {
			return errors.Wrapf(err, "moving terraform binary into /usr/local/bin")
		}
		err = terraformCmd.Wait()
		if err != nil {
			return errors.Wrapf(err, "moving terraform binary into /usr/local/bin")
		}
		colorstring.Println("[green]=> installed terraform successfully at /usr/local/bin")
	} else {
		colorstring.Println("[green]=> terraform found in $PATH!")
	}

	// Download ngrok.
	colorstring.Println("[white]=> downloading ngrok  ")
	s.Start()
	ngrokURL := fmt.Sprintf("%s/ngrok-stable-%s-%s.zip", ngrokDownloadURL, runtime.GOOS, runtime.GOARCH)
	if err = downloadAndUnzip(ngrokURL, "/tmp/ngrok.zip", "/tmp"); err != nil {
		return errors.Wrapf(err, "downloading and unzipping ngrok")
	}
	s.Stop()
	colorstring.Println("[green]=> downloaded ngrok successfully!")

	// Create ngrok tunnel.
	colorstring.Println("[white]=> creating secure tunnel")
	s.Start()

	// We use a config file so we can set ngrok's API port (web_addr). We use
	// the API to get the public URL and if there's already ngrok running, it
	// will just choose a random API port and we won't be able to get the right
	// url.
	ngrokConfig := fmt.Sprintf(`
web_addr: %s
tunnels:
  atlantis:
    addr: %d
    bind_tls: true
    proto: http
`, ngrokAPIURL, atlantisPort)

	ngrokConfigFile, err := ioutil.TempFile("", "")
	if err != nil {
		return errors.Wrap(err, "creating ngrok config file")
	}
	err = ioutil.WriteFile(ngrokConfigFile.Name(), []byte(ngrokConfig), 0600)
	if err != nil {
		return errors.Wrap(err, "writing ngrok config file")
	}

	ngrokCmd, err := executeCmd("/tmp/ngrok", []string{"start", "atlantis", "--config", ngrokConfigFile.Name()})
	if err != nil {
		return errors.Wrapf(err, "creating ngrok tunnel")
	}

	ngrokErrChan := make(chan error, 10)
	go func() {
		ngrokErrChan <- ngrokCmd.Wait()
	}()
	// When this function returns, ngrok tunnel should be stopped.
	defer ngrokCmd.Process.Kill()

	// Wait for the tunnel to be up.
	time.Sleep(2 * time.Second)
	s.Stop()
	colorstring.Println("[green]=> started tunnel!")
	tunnelURL, err := getTunnelAddr()
	if err != nil {
		return errors.Wrapf(err, "getting tunnel url")
	}
	s.Stop()

	// Start atlantis server.
	colorstring.Println("[white]=> starting atlantis server")
	s.Start()
	atlantisCmd, err := executeCmd(os.Args[0], []string{"server", "--gh-user", githubUsername, "--gh-token", githubToken, "--data-dir", "/tmp/atlantis/data", "--atlantis-url", tunnelURL, "--repo-whitelist", fmt.Sprintf("github.com/%s/%s", githubUsername, terraformExampleRepo)})
	if err != nil {
		return errors.Wrapf(err, "creating atlantis server")
	}

	atlantisErrChan := make(chan error, 10)
	go func() {
		atlantisErrChan <- atlantisCmd.Wait()
	}()
	// When this function returns atlantis server should be stopped.
	defer atlantisCmd.Process.Kill()
	colorstring.Printf("[green]=> atlantis server is now securely exposed at [bold][underline]%s\n", tunnelURL)
	fmt.Println("")

	// Create atlantis webhook.
	colorstring.Println("[white]=> creating atlantis webhook")
	s.Start()
	err = githubClient.CreateWebhook(githubUsername, terraformExampleRepo, fmt.Sprintf("%s/events", tunnelURL))
	if err != nil {
		return errors.Wrapf(err, "creating atlantis webhook")
	}
	s.Stop()
	colorstring.Println("[green]=> atlantis webhook created!")

	// Create a new pr in the example repo.
	colorstring.Println("[white]=> creating a new pull request")
	s.Start()
	pullRequestURL, err := githubClient.CreatePullRequest(githubUsername, terraformExampleRepo, "example", "master")
	if err != nil {
		return errors.Wrapf(err, "creating new pull request for repo %s/%s", githubUsername, terraformExampleRepo)
	}
	s.Stop()
	colorstring.Println("[green]=> pull request created!")

	// Open new pull request in the browser.
	colorstring.Println("[white]=> opening pull request")
	s.Start()
	time.Sleep(2 * time.Second)
	_, err = executeCmd("open", []string{pullRequestURL})
	if err != nil {
		colorstring.Printf("[red]=> opening pull request failed. please go to: %s on the browser", pullRequestURL)
	}
	s.Stop()

	// Wait for ngrok and atlantis server process to finish.
	colorstring.Println("[_green_][light_green]atlantis is running ")
	s.Start()
	colorstring.Println("[green] [press Ctrl-c to exit]")

	// Wait for SIGINT or SIGTERM signals meaning the user has Ctrl-C'd the
	// bootstrap process and want's to stop.
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	colorstring.Println("\n[red]shutdown signal received, exiting....")
	colorstring.Println("\n[green]Thank you for using atlantis :) \n[white]For more information about how to use atlantis in production go to: https://github.com/runatlantis/atlantis")
	return nil
}
