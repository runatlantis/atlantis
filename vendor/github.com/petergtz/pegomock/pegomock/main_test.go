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

package main_test

import (
	"bytes"
	"os"
	"go/build"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/pegomock/pegomock"
	. "github.com/petergtz/pegomock/pegomock/testutil"

	"testing"
)

var (
	joinPath = filepath.Join
)

func TestPegomock(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI Suite")
}

var _ = Describe("CLI", func() {

	var (
		packageDir, subPackageDir, vendorPackageDir string
		app                                         *kingpin.Application
		origWorkingDir                              string
		done                                        chan bool = make(chan bool)
	)

	BeforeEach(func() {
		packageDir = joinPath(build.Default.GOPATH, "src", "pegomocktest")
		Expect(os.MkdirAll(packageDir, 0755)).To(Succeed())
		subPackageDir = joinPath(packageDir, "subpackage")
		Expect(os.MkdirAll(subPackageDir, 0755)).To(Succeed())
		vendorPackageDir = joinPath(packageDir, "vendor", "github.com", "petergtz", "vendored_package")
		Expect(os.MkdirAll(vendorPackageDir, 0755)).To(Succeed())

		var e error
		origWorkingDir, e = os.Getwd()
		Expect(e).NotTo(HaveOccurred())
		os.Chdir(packageDir)

		WriteFile(joinPath(packageDir, "mydisplay.go"),
			"package pegomocktest; type MyDisplay interface {  Show(something string) }")
		WriteFile(joinPath(subPackageDir, "subdisplay.go"),
			"package subpackage; type SubDisplay interface {  ShowMe() }")
		WriteFile(joinPath(vendorPackageDir, "iface.go"),
			`package vendored_package; type Interface interface{ Foobar() }`)
		WriteFile(joinPath(packageDir, "vendordisplay.go"), `package pegomocktest
			import ( "github.com/petergtz/vendored_package" )
			type VendorDisplay interface { Show(something vendored_package.Interface) }`)

		app = kingpin.New("pegomock", "Generates mocks based on interfaces.")
		app.Terminate(func(int) { panic("Unexpected terminate") })
	})

	AfterEach(func() {
		Expect(os.RemoveAll(packageDir)).To(Succeed())
		os.Chdir(origWorkingDir)
	})

	Describe(`"generate" command`, func() {

		Context(`with args "MyDisplay"`, func() {

			It(`generates a file mock_mydisplay_test.go that contains "package pegomocktest_test"`, func() {
				// The rationale behind this is:
				// mocks should always be part of test packages, because we don't
				// want them to be part of the production code.
				// But to be useful, they must still reside in the package, where
				// they are actually used.

				main.Run(cmd("pegomock generate MyDisplay"), os.Stdout, app, done)

				Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context(`with args "VendorDisplay""`, func() {

			It(`generates a file mock_vendordisplay_test.go that contains 'import ( vendored_package "github.com/petergtz/vendored_package" )'`, func() {

				main.Run(cmd("pegomock generate -m VendorDisplay"), os.Stdout, app, done)
				Expect(joinPath(packageDir, "mock_vendordisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString(`vendored_package "github.com/petergtz/vendored_package"`)))
				Expect(joinPath(packageDir, "matchers", "vendored_package_interface.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString(`vendored_package "github.com/petergtz/vendored_package"`)))
			})
		})

		Context(`with args "pegomocktest/subpackage SubDisplay"`, func() {
			It(`generates a file mock_subdisplay_test.go in "pegomocktest" that contains "package pegomocktest_test"`, func() {
				main.Run(cmd("pegomock generate pegomocktest/subpackage SubDisplay"), os.Stdout, app, done)

				Expect(joinPath(packageDir, "mock_subdisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("with args mydisplay.go", func() {
			It(`generates a file mock_mydisplay_test.go that contains "package pegomocktest_test"`, func() {
				main.Run(cmd("pegomock generate mydisplay.go"), os.Stdout, app, done)

				Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("with args -d mydisplay.go", func() {
			It(`prints out debug information on stdout`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate -d mydisplay.go"), &buf, app, done)
				Expect(buf.String()).To(ContainSubstring("- method Show"))
			})
		})

		Context("with too many args", func() {

			It(`reports an error and the usage`, func() {
				var buf bytes.Buffer
				Expect(func() {
					main.Run(cmd("pegomock generate with too many args"), &buf, app, done)
				}).To(Panic())

				Expect(buf.String()).To(ContainSubstring("Please provide exactly 1 interface or 1 package + 1 interface"))
				Expect(buf.String()).To(ContainSubstring("usage"))
			})
		})

	})

	Describe(`"watch" command`, func() {

		AfterEach(func(testDone Done) { done <- true; close(testDone) }, 3)

		Context("with no further action", func() {
			It(`Creates a template file interfaces_to_mock in the current directory`, func() {
				go main.Run(cmd("pegomock watch"), os.Stdout, app, done)
				Eventually(func() string { return "interfaces_to_mock" }, "3s").Should(BeAnExistingFile())
			})
		})

		Context("after populating interfaces_to_mock with an actual interface", func() {
			It(`Eventually creates a file mock_mydisplay_test.go starting with "package pegomocktest_test"`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay")

				go main.Run(cmd("pegomock watch"), os.Stdout, app, done)

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

	})

	Context("with some unknown command", func() {
		It(`reports an error and the usage`, func() {
			var buf bytes.Buffer
			kingpin.CommandLine.Terminate(nil)
			kingpin.CommandLine.Writer(&buf)

			main.Run(cmd("pegomock some unknown command"), &buf, app, done)
			Expect(buf.String()).To(ContainSubstring("error"))
		})
	})

})

func cmd(line string) []string {
	return strings.Split(line, " ")
}
