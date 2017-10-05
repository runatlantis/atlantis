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

package watch

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/petergtz/pegomock/pegomock/filehandling"
	"github.com/petergtz/pegomock/pegomock/util"
)

const wellKnownInterfaceListFile = "interfaces_to_mock"

var join = strings.Join

type MockFileUpdater struct {
	recursive   bool
	targetPaths []string
	lastErrors  map[string]string
}

func NewMockFileUpdater(targetPaths []string, recursive bool) *MockFileUpdater {
	return &MockFileUpdater{
		targetPaths: targetPaths,
		recursive:   recursive,
		lastErrors:  make(map[string]string),
	}
}

func (updater *MockFileUpdater) Update() {
	for _, targetPath := range updater.targetPaths {
		if updater.recursive {
			filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
				if err == nil && info.IsDir() {
					util.WithinWorkingDir(path, updater.updateMockFiles)
				}
				return nil
			})
		} else {
			util.WithinWorkingDir(targetPath, updater.updateMockFiles)
		}
	}
}

func (updater *MockFileUpdater) updateMockFiles(targetPath string) {
	if _, err := os.Stat(wellKnownInterfaceListFile); os.IsNotExist(err) {
		return
	}
	for _, lineParts := range linesIn(wellKnownInterfaceListFile) {
		lineCmd := kingpin.New("What should go in here", "And what should go in here")
		destination := lineCmd.Flag("output", "Output file; defaults to mock_<interface>_test.go.").Short('o').String()
		packageOut := lineCmd.Flag("package", "Package of the generated code; defaults to the package from which pegomock was executed suffixed with _test").Default(filepath.Base(targetPath) + "_test").String()
		selfPackage := lineCmd.Flag("self_package", "If set, the package this mock will be part of.").String()
		lineArgs := lineCmd.Arg("args", "A (optional) Go package path + space-separated interface or a .go file").Required().Strings()

		_, parseErr := lineCmd.Parse(lineParts)
		if parseErr != nil {
			fmt.Println("Error while trying to generate mock for line", join(lineParts, " "), ":", parseErr)
			continue
		}
		defer func() {
			err := recover()
			if err != nil {
				if updater.lastErrors[errorKey(*lineArgs)] != fmt.Sprint(err) {
					fmt.Println("Error while trying to generate mock for", join(lineParts, " "), ":", err)
					updater.lastErrors[errorKey(*lineArgs)] = fmt.Sprint(err)
				}
			}
		}()

		util.PanicOnError(util.ValidateArgs(*lineArgs))
		sourceArgs, err := util.SourceArgs(*lineArgs)
		util.PanicOnError(err)

		generatedMockSourceCode := filehandling.GenerateMockSourceCode(sourceArgs, *packageOut, *selfPackage, false, os.Stdout, false)
		mockFilePath := filehandling.OutputFilePath(sourceArgs, ".", *destination)
		hasChanged := util.WriteFileIfChanged(mockFilePath, generatedMockSourceCode)

		if hasChanged || updater.lastErrors[errorKey(*lineArgs)] != "" {
			fmt.Println("(Re)generated mock for", errorKey(*lineArgs), "in", mockFilePath)
		}
		delete(updater.lastErrors, errorKey(*lineArgs))
	}
}

func errorKey(args []string) string {
	return join(args, "_")
}

func CreateWellKnownInterfaceListFilesIfNecessary(targetPaths []string) {
	for _, targetPath := range targetPaths {
		CreateWellKnownInterfaceListFileIfNecessary(targetPath)
	}
}

func CreateWellKnownInterfaceListFileIfNecessary(targetPath string) {
	file, err := os.OpenFile(filepath.Join(targetPath, wellKnownInterfaceListFile), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		if os.IsExist(err) {
			return
		}
		panic(err)
	}
	defer file.Close()
	file.WriteString("### List here all interfaces you would like to mock. One per line.\n")
}

func linesIn(file string) (result [][]string) {
	content, err := ioutil.ReadFile(file)
	util.PanicOnError(err)
	for _, line := range strings.Split(string(content), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") || line == "" {
			continue
		}
		parts := regexp.MustCompile(`\s`).Split(line, -1)
		// TODO: do validation here like in main
		result = append(result, parts)
	}
	return
}
