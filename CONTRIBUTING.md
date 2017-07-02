## Testing
* place tests under `{package under test}_test` to enforce testing the external interfaces
* if you need to test internally i.e. access non-exported stuff, call the file `{file under test}_internal_test.go`
* use `testing_util` for easier-to-read assertions: `import . "github.com/hootsuite/atlantis/testing_util"`
* don't try to describe the whole test by its function name. Instead use `t.Log` statements:
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
* each test should have a `t.Log` that describes what the current state is and what should happen (like a behavioural test)
