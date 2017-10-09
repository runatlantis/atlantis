package util

import (
	"io/ioutil"
	"os"
)

// WithinWorkingDir changes the current working directory temporarily and
// executes cb within that context.
func WithinWorkingDir(targetPath string, cb func(workingDir string)) {
	origWorkingDir, e := os.Getwd()
	PanicOnError(e)
	e = os.Chdir(targetPath)
	PanicOnError(e)
	defer func() { os.Chdir(origWorkingDir) }()
	cb(targetPath)
}

func WriteFileIfChanged(outputFilepath string, output []byte) bool {
	existingFileContent, err := ioutil.ReadFile(outputFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			err = ioutil.WriteFile(outputFilepath, output, 0664)
			PanicOnError(err)
			return true
		} else {
			panic(err)
		}
	}
	if string(existingFileContent) == string(output) {
		return false
	} else {
		err = ioutil.WriteFile(outputFilepath, output, 0664)
		PanicOnError(err)
		return true
	}
}
