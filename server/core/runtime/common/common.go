package common

import (
	"os"
	"os/exec"
	"strings"
)

// Looks for any argument in commandArgs that has been overridden by an entry in extra args and replaces them
// any extraArgs that are not used as overrides are added yo the end of the final string slice
func DeDuplicateExtraArgs(commandArgs []string, extraArgs []string) []string {
	// work if any of the core args have been overridden
	finalArgs := []string{}
	usedExtraArgs := []string{}
	for _, arg := range commandArgs {
		override := ""
		prefix := arg
		argSplit := strings.Split(arg, "=")
		if len(argSplit) == 2 {
			prefix = argSplit[0]
		}
		for _, extraArgOrig := range extraArgs {
			extraArg := extraArgOrig
			if strings.HasPrefix(extraArg, prefix) {
				override = extraArgOrig
				break
			}
			if strings.HasPrefix(extraArg, "--") {
				extraArg = extraArgOrig[1:]
				if strings.HasPrefix(extraArg, prefix) {
					override = extraArgOrig
					break
				}
			}
			if strings.HasPrefix(prefix, "--") {
				prefixWithoutDash := prefix[1:]
				if strings.HasPrefix(extraArg, prefixWithoutDash) {
					override = extraArgOrig
					break
				}
			}

		}
		if override != "" {
			finalArgs = append(finalArgs, override)
			usedExtraArgs = append(usedExtraArgs, override)
		} else {
			finalArgs = append(finalArgs, arg)
		}
	}
	// add any extra args that are not overrides
	for _, extraArg := range extraArgs {
		if !stringInSlice(usedExtraArgs, extraArg) {
			finalArgs = append(finalArgs, extraArg)
		}
	}
	return finalArgs
}

// returns true if a file at the passed path exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// returns true if the given file is tracked by git
func IsFileTracked(cloneDir string, filename string) (bool, error) {
	cmd := exec.Command("git", "ls-files", filename)
	cmd.Dir = cloneDir

	output, err := cmd.CombinedOutput()

	if err != nil {
		return false, err
	}
	return len(output) > 0, nil

}

func stringInSlice(stringSlice []string, target string) bool {
	for _, value := range stringSlice {
		if value == target {
			return true
		}
	}
	return false
}
