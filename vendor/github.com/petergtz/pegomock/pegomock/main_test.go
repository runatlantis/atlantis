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
	"go/build"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	main "github.com/petergtz/pegomock/pegomock"
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
		WriteFile(joinPath(packageDir, "http_request_handler.go"),
			`package pegomocktest; import "net/http"; type RequestHandler interface {  Handler(r *http.Request) }`)
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

				main.Run(cmd("pegomock generate MyDisplay"), os.Stdout, os.Stdin, app, done)

				Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context(`with args "VendorDisplay""`, func() {

			It(`generates a file mock_vendordisplay_test.go that contains 'import ( vendored_package "github.com/petergtz/vendored_package" )'`, func() {

				main.Run(cmd("pegomock generate -m VendorDisplay"), os.Stdout, os.Stdin, app, done)
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
				main.Run(cmd("pegomock generate pegomocktest/subpackage SubDisplay"), os.Stdout, os.Stdin, app, done)

				Expect(joinPath(packageDir, "mock_subdisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("with args mydisplay.go", func() {
			It(`generates a file mock_mydisplay_test.go that contains "package pegomocktest_test"`, func() {
				main.Run(cmd("pegomock generate mydisplay.go"), os.Stdout, os.Stdin, app, done)

				Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("with args -d mydisplay.go", func() {
			It(`prints out debug information on stdout`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate -d mydisplay.go"), &buf, os.Stdin, app, done)
				Expect(buf.String()).To(ContainSubstring("- method Show"))
			})
		})

		Context("with args -o where output override is a path with a non-existing directory", func() {
			It(`creates an output directory before code generation`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate mydisplay.go -o testoutput/test.go"), &buf, os.Stdin, app, done)
				Expect(joinPath(packageDir, "testoutput/test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

		Context("with args --output-dir", func() {
			It(`creates the mocks in output dir with the dir's basename as package name`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate MyDisplay --output-dir fakes"), &buf, os.Stdin, app, done)
				Expect(joinPath(packageDir, "fakes/mock_mydisplay.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package fakes")))
			})
		})

		Context("with args --output-dir and --package", func() {
			It(`creates the mocks in output dir with the specified package name`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate MyDisplay --output-dir fakes --package other"), &buf, os.Stdin, app, done)

				Expect(joinPath(packageDir, "fakes/mock_mydisplay.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package other")))
			})
		})

		Context("with args for specifying matcher directory", func() {
			It(`creates matchers in the specified directory`, func() {
				var buf bytes.Buffer
				main.Run(cmd("pegomock generate --generate-matchers --matchers-dir custom/matcher/dir VendorDisplay"), &buf, os.Stdin, app, done)

				Expect(joinPath(packageDir, "custom/matcher/dir")).To(BeADirectory())
				Expect(joinPath(packageDir, "custom/matcher/dir/vendored_package_interface.go")).To(BeAnExistingFile())
			})
		})

		Context("with too many args", func() {

			It(`reports an error and the usage`, func() {
				var buf bytes.Buffer
				Expect(func() {
					main.Run(cmd("pegomock generate with too many args"), &buf, os.Stdin, app, done)
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
				go main.Run(cmd("pegomock watch"), os.Stdout, os.Stdin, app, done)
				Eventually(func() string { return "interfaces_to_mock" }, "3s").Should(BeAnExistingFile())
			})
		})

		Context("after populating interfaces_to_mock with an actual interface", func() {
			It(`Eventually creates a file mock_mydisplay_test.go starting with "package pegomocktest_test"`, func() {
				WriteFile(joinPath(packageDir, "interfaces_to_mock"), "MyDisplay")

				go main.Run(cmd("pegomock watch"), os.Stdout, os.Stdin, app, done)

				Eventually(joinPath(packageDir, "mock_mydisplay_test.go"), "3s").Should(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))
			})
		})

	})

	Describe(`"remove" command`, func() {
		Context("there are no mock files", func() {
			It("removes mock files in current directory only", func() {
				var buf bytes.Buffer

				main.Run(cmd("pegomock remove -n"), &buf, os.Stdin, app, done)

				Expect(buf.String()).To(ContainSubstring(`No files to remove.`))
			})
		})

		Context("there are some mock files", func() {
			BeforeEach(func() {
				main.Run(cmd("pegomock generate MyDisplay"), os.Stdout, os.Stdin, app, done)
				Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package pegomocktest_test")))

				main.Run(cmd("pegomock generate --output-dir "+subPackageDir+" pegomocktest/subpackage SubDisplay"), os.Stdout, os.Stdin, app, done)

				Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(SatisfyAll(
					BeAnExistingFile(),
					BeAFileContainingSubString("package subpackage")))
			})

			Context("Non-interactive", func() {
				Context("non-recursive", func() {
					It("removes mock files in current directory only", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove -n"), &buf, os.Stdin, app, done)

						Expect(buf.String()).To(ContainSubstring(`Deleting the following files:
` + build.Default.GOPATH + `/src/pegomocktest/mock_mydisplay_test.go`))
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).NotTo(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(BeAnExistingFile())
					})
				})

				Context("recursive", func() {
					It("removes mock files recursively", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove -n -r"), &buf, os.Stdin, app, done)

						Expect(buf.String()).To(ContainSubstring(`Deleting the following files:
` + build.Default.GOPATH + `/src/pegomocktest/mock_mydisplay_test.go
` + build.Default.GOPATH + `/src/pegomocktest/subpackage/mock_subdisplay.go`))
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).NotTo(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).NotTo(BeAnExistingFile())
					})

					Context("custom matchers were generated", func() {
						BeforeEach(func() {
							main.Run(cmd("pegomock generate -m RequestHandler"), os.Stdout, os.Stdin, app, done)
							Expect(joinPath(packageDir, "mock_requesthandler_test.go")).To(SatisfyAll(
								BeAnExistingFile(),
								BeAFileContainingSubString("package pegomocktest_test")))

							Expect(joinPath(packageDir, "matchers", "ptr_to_http_request.go")).To(SatisfyAll(
								BeAnExistingFile(),
								BeAFileContainingSubString("package matchers"),
							))
							main.Run(cmd("pegomock generate -m --output-dir "+subPackageDir+" RequestHandler"), os.Stdout, os.Stdin, app, done)

							Expect(joinPath(subPackageDir, "mock_requesthandler.go")).To(SatisfyAll(
								BeAnExistingFile(),
								BeAFileContainingSubString("package subpackage")))
							Expect(joinPath(subPackageDir, "matchers", "ptr_to_http_request.go")).To(SatisfyAll(
								BeAnExistingFile(),
								BeAFileContainingSubString("package matchers"),
							))

						})

						It("removes matchers and matchers dir too", func() {
							var buf bytes.Buffer

							main.Run(cmd("pegomock remove -n -r"), &buf, os.Stdin, app, done)

							Expect(buf.String()).To(Equal(`Deleting the following files:
`+build.Default.GOPATH+`/src/pegomocktest/matchers
`+build.Default.GOPATH+`/src/pegomocktest/matchers/ptr_to_http_request.go
`+build.Default.GOPATH+`/src/pegomocktest/mock_mydisplay_test.go
`+build.Default.GOPATH+`/src/pegomocktest/mock_requesthandler_test.go
`+build.Default.GOPATH+`/src/pegomocktest/subpackage/matchers
`+build.Default.GOPATH+`/src/pegomocktest/subpackage/matchers/ptr_to_http_request.go
`+build.Default.GOPATH+`/src/pegomocktest/subpackage/mock_requesthandler.go
`+build.Default.GOPATH+`/src/pegomocktest/subpackage/mock_subdisplay.go
`), buf.String())
							Expect(joinPath(packageDir, "mock_mydisplay_test.go")).NotTo(BeAnExistingFile())
							Expect(joinPath(subPackageDir, "mock_subdisplay.go")).NotTo(BeAnExistingFile())
						})
					})
				})

				Context("Silent", func() {
					It("removes mock files, but provides no output", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove -n --silent"), &buf, os.Stdin, app, done)

						Expect(buf.String()).To(BeEmpty())
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).NotTo(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(BeAnExistingFile())
					})
				})

				Context("with path", func() {
					It("removes mock files in path", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove -n "+subPackageDir), &buf, os.Stdin, app, done)

						Expect(buf.String()).To(ContainSubstring(`Deleting the following files:
` + build.Default.GOPATH + `/src/pegomocktest/subpackage/mock_subdisplay.go`))
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).NotTo(BeAnExistingFile())
					})
				})
			})

			Context("Interactive", func() {
				Context("confirming with yes", func() {
					It("removes mock files", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove"), &buf, strings.NewReader("yes\n"), app, done)

						Expect(buf.String()).To(ContainSubstring(`Will delete the following files:
` + build.Default.GOPATH + `/src/pegomocktest/mock_mydisplay_test.go
Continue? [y/n]:`))
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).NotTo(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(BeAnExistingFile())
					})
				})

				Context("confirming with no", func() {
					It("does not remove mock files", func() {
						var buf bytes.Buffer

						main.Run(cmd("pegomock remove"), &buf, strings.NewReader("no\n"), app, done)

						Expect(buf.String()).To(ContainSubstring(`Will delete the following files:
` + build.Default.GOPATH + `/src/pegomocktest/mock_mydisplay_test.go
Continue? [y/n]:`))
						Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(BeAnExistingFile())
						Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(BeAnExistingFile())
					})
				})
			})

			Context("dry-run", func() {
				It("removes no mock files, but provides files that would be deleted", func() {
					var buf bytes.Buffer

					main.Run(cmd("pegomock remove --dry-run"), &buf, os.Stdin, app, done)

					Expect(buf.String()).To(ContainSubstring(`Would delete the following files:
` + build.Default.GOPATH + `/src/pegomocktest/mock_mydisplay_test.go`))
					Expect(joinPath(packageDir, "mock_mydisplay_test.go")).To(BeAnExistingFile())
					Expect(joinPath(subPackageDir, "mock_subdisplay.go")).To(BeAnExistingFile())
				})
			})
		})
	})

	Context("with some unknown command", func() {
		It(`reports an error and the usage`, func() {
			var buf bytes.Buffer
			kingpin.CommandLine.Terminate(nil)
			kingpin.CommandLine.Writer(&buf)

			main.Run(cmd("pegomock some unknown command"), &buf, os.Stdin, app, done)
			Expect(buf.String()).To(ContainSubstring("error"))
		})
	})

})

func cmd(line string) []string {
	return strings.Split(line, " ")
}
