# Contributing <!-- omit in toc -->

# Table of Contents <!-- omit in toc -->
- [Reporting Issues](#reporting-issues)
- [Reporting Security Issues](#reporting-security-issues)
- [Updating The Website](#updating-the-website)
- [Developing](#developing)
  - [Running Atlantis Locally](#running-atlantis-locally)
  - [Running Atlantis With Local Changes](#running-atlantis-with-local-changes)
    - [Rebuilding](#rebuilding)
  - [Running Tests Locally](#running-tests-locally)
  - [Running Tests In Docker](#running-tests-in-docker)
  - [Calling Your Local Atlantis From GitHub](#calling-your-local-atlantis-from-github)
  - [Code Style](#code-style)
    - [Logging](#logging)
    - [Errors](#errors)
    - [Testing](#testing)
    - [Mocks](#mocks)
- [Backporting Fixes](#backporting-fixes)
  - [Manual Backporting Fixes](#manual-backporting-fixes)
- [Creating a New Release](#creating-a-new-release)

# Reporting Issues
* When reporting issues, please include the output of `atlantis version`.
* Also include the steps required to reproduce the problem if possible and applicable. This information will help us review and fix your issue faster.
* When sending lengthy log-files, consider posting them as a gist (https://gist.github.com). Don't forget to remove sensitive data from your logfiles before posting (you can replace those parts with "REDACTED").

# Reporting Security Issues
We take security issues seriously. Please report a security vulnerability to the maintainers using [private vulnerability reporting](https://github.com/runatlantis/atlantis/security/advisories/new).

# Updating The Website
* To view the generated website locally, run `npm website:dev` and then
open your browser to http://localhost:8080.
* The website will be regenerated when your pull request is merged to main.

# Developing

## Running Atlantis Locally
* Clone the repo from https://github.com/runatlantis/atlantis/
* Compile Atlantis:
    ```sh
    go install
    ```
* Run Atlantis:
    ```sh
    atlantis server --gh-user <your username> --gh-token <your token> --repo-allowlist <your repo> --gh-webhook-secret <your webhook secret> --log-level debug
    ```
    If you get an error like `command not found: atlantis`, ensure that `$GOPATH/bin` is in your `$PATH`.

## Running Atlantis With Local Changes
Docker compose is set up to start an atlantis container and ngrok container in the same network in order to expose the atlantis instance to the internet.  In order to do this, create a file in the repository called `atlantis.env` and add the required env vars for the atlantis server configuration.

e.g.

```sh
NGROK_AUTH=1234567890

ATLANTIS_GH_APP_ID=123
ATLANTIS_GH_APP_KEY_FILE="/.ssh/somekey.pem"
ATLANTIS_GH_WEBHOOK_SECRET=12345
```

Note: `~/.ssh` is mounted to allow for referencing any local ssh keys.

Following this just run:

```sh
make build-service
docker-compose up --detach
docker-compose logs --follow
```

### Rebuilding
If the ngrok container is restarted, the url will change which is a hassle. Fortunately, when we make a code change, we can rebuild and restart the atlantis container easily without disrupting ngrok.

e.g.

```sh
make build-service
docker-compose up --detach --build
```

## Running Tests Locally
`make test`. If you want to run the integration tests that actually run real `terraform` commands, run `make test-all`.

## Running Tests In Docker
```sh
docker run --rm -v $(pwd):/go/src/github.com/runatlantis/atlantis -w /go/src/github.com/runatlantis/atlantis ghcr.io/runatlantis/testing-env:latest make test
```

Or to run the integration tests

```sh
docker run --rm -v $(pwd):/go/src/github.com/runatlantis/atlantis -w /go/src/github.com/runatlantis/atlantis ghcr.io/runatlantis/testing-env:latest make test-all
```

## Calling Your Local Atlantis From GitHub
- Create a test terraform repository in your GitHub.
- Create a personal access token for Atlantis. See [Create a GitHub token](https://github.com/runatlantis/atlantis/tree/main/runatlantis.io/docs/access-credentials.md#generating-an-access-token).
- Start Atlantis in server mode using that token:
```sh
atlantis server --gh-user <your username> --gh-token <your token> --repo-allowlist <your repo> --gh-webhook-secret <your webhook secret> --log-level debug
```
- Download ngrok from https://ngrok.com/download. This will enable you to expose Atlantis running on your laptop to the internet so GitHub can call it.
- When you've downloaded and extracted ngrok, run it on port `4141`:
```sh
ngrok http 4141
```
- Create a Webhook in your repo and use the `https` url that `ngrok` printed out after running `ngrok http 4141`. Be sure to append `/events` so your webhook url looks something like `https://efce3bcd.ngrok.io/events`. See [Add GitHub Webhook](https://github.com/runatlantis/atlantis/blob/main/runatlantis.io/docs/configuring-webhooks.md#configuring-webhooks).
- Create a pull request and type `atlantis help`. You should see the request in the `ngrok` and Atlantis logs and you should also see Atlantis comment back.

## Code Style

### Logging
- `ctx.Log` should be available in most methods. If not, pass it down.
- levels:
    - debug is for developers of atlantis
    - info is for users (expected that people run on info level)
    - warn is for something that might be a problem but we're not sure
    - error is for something that's definitely a problem
- **ALWAYS** logs should be all lowercase (when printed, the first letter of each line will be automatically capitalized)
- **ALWAYS** quote any string variables using %q in the fmt string, ex. `ctx.Log.Info("cleaning clone dir %q", dir)` => `Cleaning clone directory "/tmp/atlantis/lkysow/atlantis-terraform-test/3"`
- **NEVER** use colons "`:`" in a log since that's used to separate error descriptions and causes
  - if you need to have a break in your log, either use `-` or `,` ex. `failed to clean directory, continuing regardless`

### Errors
- **ALWAYS** use lowercase unless the word requires it
- **ALWAYS** use `errors.Wrap(err, "additional context...")"` instead of `fmt.Errorf("additional context: %s", err)`
because it is less likely to result in mistakes and gives us the ability to trace call stacks
- **NEVER** use the words "error occurred when...", or "failed to..." or "unable to...", etc. Instead, describe what was occurring at
time of the error, ex. "cloning repository", "creating AWS session". This will prevent errors from looking like
```
Error setting up workspace: failed to run git clone: could find git
```

and will instead look like
```
Error: setting up workspace: running git clone: no executable "git"
```
This is easier to read and more consistent

### Testing
- place tests under `{package under test}_test` to enforce testing the external interfaces
- if you need to test internally i.e. access non-exported stuff, call the file `{file under test}_internal_test.go`
- use our testing utility for easier-to-read assertions: `import . "github.com/runatlantis/atlantis/testing"` and then use `Assert()`, `Equals()` and `Ok()`

### Mocks
We use [pegomock](https://github.com/petergtz/pegomock) for mocking. If you're
modifying any interfaces that are mocked, you'll need to regen the mocks for that
interface.

Install using `go install github.com/petergtz/pegomock/v4/pegomock@latest`

If you see errors like:
```
# github.com/runatlantis/atlantis/server/events [github.com/runatlantis/atlantis/server/events.test]
server/events/project_command_builder_internal_test.go:567:5: cannot use workingDir (type *MockWorkingDir) as type WorkingDir in field value:
	*MockWorkingDir does not implement WorkingDir (missing ListAllFiles method)
```

Then you've likely modified an interface and now need to update the mocks.

Each interface that is mocked has a `go:generate` command above it, e.g.
```go
//go:generate pegomock generate -m --package mocks -o mocks/mock_project_command_builder.go ProjectCommandBuilder

type ProjectCommandBuilder interface {
	BuildAutoplanCommands(ctx *command.Context) ([]command.ProjectContext, error)
}
```

To regen the mock, run `go generate` on that file, e.g.
```sh
go generate server/events/project_command_builder.go
```

Alternatively, you can run `make go-generate` to execute `go generate` across all packages


# Backporting Fixes
Atlantis now uses a [cherry-pick-bot](https://github.com/googleapis/repo-automation-bots/tree/main/packages/cherry-pick-bot) from Google. The bot assists in maintaining changes across releases branches by easily cherry-picking changes via pull requests.

Maintainers and Core Contributors can add a comment to a pull request:

```sh
/cherry-pick target-branch-name
```

target-branch-name is the branch to cherry-pick to. cherry-pick-bot will cherry-pick the merged commit to a new branch (created from the target branch) and open a new pull request to the target branch.

The bot will immediately try to cherry-pick a merged PR. On unmerged pull request, it will not do anything immediately, but wait until merge. You can comment multiple times on a PR for multiple release branches.

## Manual Backporting Fixes
The bot will fail to cherry-pick if the feature branches' git history is not linear (merge commits instead of rebase). In that case, you will need to manually cherry-pick the squashed merged commit from main to the release branch

1. Switch to the release branch intended for the fix.
1. Run `git cherry-pick <sha>` with the commit hash from the main branch.
1. Push the newly cherry-picked commit up to the remote release branch.

# Creating a New Release
1. (Major/Minor release only) Create a new release branch `release-x.y`
1. Go to https://github.com/runatlantis/atlantis/releases and click "Draft a new release"
    1. Prefix version with `v` and increment based on last release.
    1. The title of the release is the same as the tag (ex. v0.2.2)
    1. Fill in description by clicking on the "Generate Release Notes" button.
        1. You may have to manually move around some commit titles as they are determined by PR labels (see .github/labeler.yml & .github/release.yml)
    1. (Latest Major/Minor branches only) Make sure the release is set as latest
        1. Don't set "latest release" for patches on older release branches.
1. Check and update the default version in `Chart.yaml` in [the official Helm chart](https://github.com/runatlantis/helm-charts/blob/main/charts/atlantis/values.yaml) as needed.
