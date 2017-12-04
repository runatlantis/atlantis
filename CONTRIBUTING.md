# Developing

## Running Atlantis Locally
Get the source code:
```
go get github.com/hootsuite/atlantis
```
This will clone Atlantis into `$GOPATH/src/github.com/hootsuite/atlantis` (where `$GOPATH` defaults to `~/go`).

Go to that directory:
```
cd $GOPATH/src/github.com/hootsuite/atlantis
```

Compile Atlantis:
```
go install
```

Run Atlantis:
```
atlantis server --gh-user <your username> --gh-token <your token> --log-level debug
```
If you get an error like `command not found: atlantis`, ensure that `$GOPATH/bin` is in your `$PATH`.

## Calling Your Local Atlantis From GitHub
- Create a test terraform repository in your GitHub.
- Create a personal access token for Atlantis. See [Create a GitHub token](https://github.com/hootsuite/atlantis#create-a-github-token).
- Start Atlantis in server mode using that token:
```
atlantis server --gh-user <your username> --gh-token <your token> --log-level debug
```
- Download ngrok from https://ngrok.com/download. This will enable you to expose Atlantis running on your laptop to the internet so GitHub can call it.
- When you've downloaded and extracted ngrok, run it on port `4141`:
```
ngrok http 4141
```
- Create a WebHook in your repo and use the `https` url that `ngrok` printed out after running `ngrok http 4141`. Be sure to append `/events` so your webhook url looks something like `https://efce3bcd.ngrok.io/events`. See [Add GitHub Webhook](https://github.com/hootsuite/atlantis#add-github-webhook).
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
- use our testing utility for easier-to-read assertions: `import . "github.com/hootsuite/atlantis/testing"` and then use `Assert()`, `Equals()` and `Ok()`
- don't try to describe the whole test by its function name. Instead use `t.Log` statements:
```go
// don't do this
func TestLockingWhenThereIsAnExistingLockForNewEnv(t *testing.T) {
    ...

// do this
func TestLockingExisting(t *testing.T) {
    	t.Log("if there is an existing lock, lock should...")
        ...
       	t.Log("...succeed if the new project has a different path") {
             // optionally wrap in a block so it's easier to read
       }
```
- each test should have a `t.Log` that describes what the current state is and what should happen (like a behavioural test)

# Glossary
* **Run**: Encompasses the two steps (plan and apply) for modifying infrastructure in a specific environment
* **Project Lock**: When a run has started but is not yet completed, the infrastructure and environment that's being modified is "locked" against
other runs being started for the same set of infrastructure and environment. We determine what infrastructure is being modified by combining the
repository name, the directory in the repository at which the terraform commands need to be run, and the environment that's being modified
* **Project Path**: The path relative to the repository's root at which terraform commands need to be executed for this Run

# Creating a New Release
1. Update version number in
    1. `main.go`
    1. `website/src/themes/kube/layouts/index.html`
1. Update `CHANGELOG.md` with latest release number and information
1. Create a pull request and merge to master
1. Run `make release`
1. Go to https://github.com/hootsuite/atlantis/releases and click "Draft a new release"
    1. Prefix version with `v`
    1. Description should just be `See CHANGELOG` with link to CHANGELOG at that release
    1. Drag in binaries made with `make release`
1. Run `make generate-website-html` and `make upload-website-html` to update website
